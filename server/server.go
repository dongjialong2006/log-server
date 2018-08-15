package server

import (
	"crypto/md5"
	"fmt"
	"log-server/config"
	"log-server/util/etcd"
	"log-server/util/network"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/dongjialong2006/log"
	"golang.org/x/net/context"
)

type Server struct {
	sync.RWMutex
	opt  *option
	ctx  context.Context
	srv  net.Listener
	log  *log.Entry
	etcd *clientv3.Client
}

const defaultIP = "127.0.0.1"

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		ctx: context.Background(),
		log: log.New("server"),
		opt: &option{},
	}

	var err error = nil
	if s.etcd, err = etcd.NewClient(context.Background(), cfg.Store, cfg.User, cfg.Pwd); nil != err {
		return nil, err
	}

	if cfg.Port <= 0 {
		cfg.Port = 55505
	}

	if err = s.listener(cfg); nil != err {
		return nil, err
	}

	if "" != cfg.Path {
		go s.watch(cfg.Path, cfg.Identity, cfg.Separator)
	}

	s.opt.ttl = cfg.Ttl
	if cfg.Ttl < 0 {
		s.opt.ttl = 60480
	} else if 0 == cfg.Ttl {
		s.opt.ttl = 2592000
	}

	s.opt.UpdateIdentity(cfg.Identity)
	s.opt.UpdateTimeout("3")

	return s, nil
}

func (s *Server) listener(cfg *config.Config) error {
	var ip string = ""
	tmp, err := network.LocalIP("eth0")
	if nil != err {
		ip, err = network.ExternalIP()
		if nil != err {
			s.log.Warnf("no valid ip was found, th system will use default ip:127.0.0.1")
			ip = defaultIP
		}
	} else {
		ip = tmp.String()
	}

	s.opt.addr = fmt.Sprintf("%s:%d", ip, cfg.Port)

	if s.srv, err = net.Listen("tcp", s.opt.addr); err != nil {
		return err
	}

	s.opt.namespace = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%d", ip, cfg.Port))))

	return nil
}

func (s *Server) Stop() {
	if nil != s.srv {
		s.srv.Close()
	}

	s.Lock()
	defer s.Unlock()
	if nil != s.etcd {
		s.etcd.Close()
		s.etcd = nil
	}
}

func (s *Server) Serve() error {
	s.log.Debugf("log-server starting tcp server listener addr:%s.", s.srv.Addr().String())
	go s.notify()
	go s.sync()
	go s.close()

	for {
		conn, err := s.srv.Accept()
		if err != nil {
			return err
		}

		go s.handle(conn)
	}

	return nil
}

func (s *Server) close() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			if s.opt.GetStop() == "true" {
				s.Stop()
				return
			}
		}
	}
}

func (s *Server) initDir() error {
	resp, err := s.etcd.Get(s.ctx, fmt.Sprintf("/%s/configs", s.opt.namespace), clientv3.WithPrefix())
	if nil != err {
		return err
	}

	if 0 == resp.Count {
		if _, err = s.etcd.Put(s.ctx, fmt.Sprintf("/%s/configs/timeout", s.opt.namespace), s.opt.GetTimeout()); nil != err {
			return err
		}

		if _, err = s.etcd.Put(s.ctx, fmt.Sprintf("/%s/configs/ttl", s.opt.namespace), fmt.Sprintf("%d", s.opt.ttl)); nil != err {
			return err
		}
	}

	resp, err = s.etcd.Get(s.ctx, fmt.Sprintf("/%s/configs", s.opt.namespace), clientv3.WithPrefix())
	if nil != err {
		return err
	}

	for _, key := range resp.Kvs {
		if strings.HasSuffix(string(key.Key), "ttl") {
			tmp, _ := strconv.Atoi(string(key.Value))
			s.opt.UpdateTTL(int64(tmp))
		}
		if strings.HasSuffix(string(key.Key), "timeout") {
			s.opt.UpdateTimeout(string(key.Value))
		}
	}

	go s.keepalive()

	return nil
}

func (s *Server) keepalive() {
	tick := time.Tick(time.Second * 2)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			s.RLock()
			if nil == s.etcd {
				s.RUnlock()
				return
			}
			resp, err := s.etcd.Grant(s.ctx, 3)
			if nil != err {
				s.RUnlock()
				s.log.Error(err)
				continue
			}

			if _, err = s.etcd.Put(s.ctx, fmt.Sprintf("/nodes/%s", s.opt.addr), s.opt.namespace, clientv3.WithLease(resp.ID)); nil != err {
				s.log.Error(err)
			}
			s.RUnlock()
		}
	}

	return
}

func (s *Server) sync() {
	if err := s.initDir(); nil != err {
		s.log.Error(err)
		return
	}

	watch := s.etcd.Watch(s.ctx, fmt.Sprintf("/%s/configs", s.opt.namespace), clientv3.WithPrefix())
	for {
		select {
		case <-s.ctx.Done():
			return
		case resp, ok := <-watch:
			if !ok {
				return
			}

			for _, event := range resp.Events {
				if nil == event {
					continue
				}

				if strings.HasSuffix(string(event.Kv.Key), "timeout") {
					s.opt.UpdateTimeout(string(event.Kv.Value))
				}

				if strings.HasSuffix(string(event.Kv.Key), "stop") {
					s.opt.UpdateStop(string(event.Kv.Value))
				}

				if strings.HasSuffix(string(event.Kv.Key), "ttl") {
					tmp, _ := strconv.Atoi(string(event.Kv.Value))
					s.opt.UpdateTTL(int64(tmp))
				}
			}
		}
	}

	return
}

func (s *Server) notify() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGSTOP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)

	<-sigChan
	signal.Stop(sigChan)

	s.Stop()

	s.log.Infof("pid:%d is closed.", os.Getpid())

	return
}
