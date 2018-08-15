package etcd

import (
	"context"
	"fmt"

	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/coreos/etcd/clientv3"
	"github.com/juju/errors"
)

func NewClient(ctx context.Context, addr string, user string, pwd string) (*clientv3.Client, error) {
	var endpoints []string = nil
	if strings.HasPrefix(addr, "http://") {
		endpoints = append(endpoints, addr)
	} else {
		endpoints = append(endpoints, fmt.Sprintf("http://%s", addr))
	}

	cfg := clientv3.Config{
		Context:          ctx,
		Endpoints:        endpoints,
		DialTimeout:      15 * time.Second,
		AutoSyncInterval: 30 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    5 * time.Minute,
				Timeout: 10 * time.Second,
			}),
		},
	}

	if "" != user {
		cfg.Username = user
		cfg.Password = pwd
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return c, nil
}
