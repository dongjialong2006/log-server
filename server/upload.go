package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	br := bufio.NewReader(conn)
	for {
		data, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}

		data = strings.Trim(data, " ")
		data = strings.Trim(data, "\n")
		if "" == data {
			continue
		}

		if "" == data || len(data) < 3 {
			continue
		}

		if err = s.check(data); nil != err {
			s.response(err.Error(), conn)
			continue
		}

		if err = s.do(data); nil != err {
			s.response(err.Error(), conn)
			continue
		}

		s.response("ok", conn)
	}
}

func (s *Server) check(data string) error {
	if !strings.HasPrefix(data, "{") || !strings.HasSuffix(data, "}") {
		return fmt.Errorf("received data format error.")
	}

	return nil
}

func (s *Server) response(resp string, conn net.Conn) {
	conn.Write([]byte(resp))
}

func (s *Server) do(data string) error {
	data = strings.ToLower(data)

	if !strings.Contains(data, "time") {
		return fmt.Errorf("received log entry format error, not found time key.")
	}

	if "" == s.opt.GetIdentity() {
		if !strings.Contains(data, "identity") {
			return fmt.Errorf("received log entry format error, not found identity key.")
		}
	}

	var tmp map[string]interface{} = make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &tmp)
	if nil != err {
		return err
	}

	key := tmp["time"]
	delete(tmp, "time")

	identity := s.opt.GetIdentity()
	temp, ok := tmp["identity"]
	if ok {
		identity = temp.(string)
		delete(tmp, "identity")
	}

	tran, _ := json.Marshal(tmp)

	_, err = s.etcd.Put(s.ctx, fmt.Sprintf("/%s/%s/%s", s.opt.namespace, identity, key), string(tran), clientv3.WithLease(s.grant()))

	return err
}

func (s *Server) watch(dir string, identity string, separator string) {
	if "" == dir {
		s.log.Errorf("log dir is empty.")
		return
	}

	if "" == identity {
		s.log.Errorf("identity is empty.")
		return
	}

	logs := make(map[string]*attr)
	tick := time.Tick(time.Second)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				s.log.Error(err)
				continue
			}

			if 0 == len(files) {
				continue
			}

			for _, f := range files {
				if f.IsDir() {
					continue
				}

				name := path.Join(dir, f.Name())
				tmp, ok := logs[name]
				if ok {
					if 1 == atomic.LoadInt32(&tmp.end) {
						continue
					}
					if f.Size() > tmp.size {
						tmp.size = f.Size()
						atomic.AddInt32(&tmp.end, 1)
						go s.load(name, identity, separator, tmp)
					}
					continue
				}

				tmp = &attr{
					size: f.Size(),
				}
				logs[name] = tmp
				atomic.AddInt32(&tmp.end, 1)
				go s.load(name, identity, separator, tmp)
			}
		}
	}

	return
}

func (s *Server) load(name string, identity string, separator string, info *attr) {
	defer atomic.AddInt32(&info.end, -1)
	file, err := os.Open(name)
	if nil != err {
		if os.IsNotExist(err) {
			s.log.WithField("name", name).Error("file is not exist.")
		} else {
			s.log.WithField("name", name).Error(err)
		}
		return
	}
	defer file.Close()

	if "" == separator {
		separator = " "
	}

	var num int64 = 0
	var key string = ""
	var old string = ""
	var first bool = true
	var header bool = false
	var commit bool = false
	var reader = bufio.NewReader(file)
	for {
		s.RLock()
		if nil == s.etcd {
			s.RUnlock()
			return
		}
		s.RUnlock()

		data, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		num++

		if num <= info.getNum() {
			continue
		}

		tmp := string(data)

		tmp = strings.Trim(tmp, " ")
		if "" == tmp || tmp == "EOF" {
			continue
		}

		key, header = s.key(tmp, separator)
		if "" == key {
			continue
		}
		if 0 == num {
			first = header
		}

		if first && !header {
			old += tmp
			commit = true
			continue
		}

		if err = s.put(identity, key, tmp); nil != err {
			s.log.WithField("name", name).Warn(err)
			continue
		}

		if commit {
			key, header = s.key(old, separator)
			if err = s.put(identity, key, old); nil != err {
				s.log.WithField("name", name).Warn(err)
				continue
			}
			commit = false
		}

		old = tmp
	}
	info.update(num)

	return
}

func (s *Server) put(identity, key, value string) error {
	s.RLock()
	defer s.RUnlock()
	if nil == s.etcd {
		return fmt.Errorf("etcd is closed.")
	}

	_, err := s.etcd.Put(s.ctx, fmt.Sprintf("/%s/%s/%s", s.opt.namespace, identity, key), value, clientv3.WithLease(s.grant()))

	return err
}

func (s *Server) grant() clientv3.LeaseID {
	resp, err := s.etcd.Grant(s.ctx, s.opt.GetTTL())
	if nil != err {
		return clientv3.LeaseID(60480)
	}

	return resp.ID
}

func (s *Server) key(data string, separator string) (string, bool) {
	data = strings.ToLower(data)
	pos := strings.Index(data, "time=")
	if -1 == pos {
		return time.Now().Format("2006-01-02 15:04:05.00000000"), false
	}

	data = data[pos+5:]

	if '"' == data[0] {
		data = data[1:]
		pos = strings.Index(data, "\"")
		return data[:pos], true
	}

	pos = strings.Index(data, separator)
	if -1 == pos {
		s.log.Errorf("data:%s format error.", data)
		return "", true
	}

	return data[:pos], true
}
