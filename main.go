package main

import (
	"fmt"
	"log-server/config"
	"log-server/server"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongjialong2006/log"
	"github.com/namsral/flag"
)

func main() {
	var cfg config.Config
	fs := flag.NewFlagSetWithEnvPrefix("log-server", "LOG_SERVER_", flag.ContinueOnError)
	fs.Int64Var(&cfg.Ttl, "ttl", 604800, "every log ttl")
	fs.Int64Var(&cfg.Port, "port", 55505, "local port")
	fs.StringVar(&cfg.Path, "path", "", "log file dir")
	fs.StringVar(&cfg.Store, "store", "127.0.0.1:2379", "etcd db addr")
	fs.StringVar(&cfg.Identity, "identity", "", "model identity")
	fs.StringVar(&cfg.Separator, "separator", "", "log separator")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		return
	}

	if "" == cfg.Store {
		fmt.Println("etcd server addr is empty.")
		return
	}

	logger := log.New("main")
	srv, err := server.New(&cfg)
	if err != nil {
		logger.Fatal(err)
	}

	err = signalNotify(srv)
	if err != nil {
		logger.Fatal(err)
	}

	if err = srv.Serve(); err != nil {
		logger.Warn(err)
	}

	time.Sleep(time.Second)
	return
}

func init() {
	if err := log.InitLocalLogSystem(log.WithLogLevel("debug"), log.WithLogName("./log/local.log")); nil != err {
		fmt.Println(err)
		return
	}
}

func signalNotify(srv *server.Server) error {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigChan
		signal.Stop(sigChan)

		if nil != srv {
			srv.Stop()
		}
	}()
	return nil
}
