package server

import (
	"fmt"
	"log-server/util/etcd"
	"net/url"
	//"path"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

func TestETED(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ETCD Suite")
}

var _ = Describe("ETCD", func() {
	It("etcd", func() {
		ctx := context.Background()
		client, err := etcd.NewClient(ctx, url.URL{
			Scheme: "etcd",
			Host:   "127.0.0.1:2379",
			// Path:   path.Join("nodes-config", "10.95.128.71-55505"),
		})

		Expect(err).Should(BeNil())
		time.Sleep(time.Second * 2)

		_, err = client.Put(ctx, "/dcf11/d1", "xxxxxxxxxx")
		Expect(err).Should(BeNil())

		_, err = client.Put(ctx, "/dcf11/d2", "xxxxxxxxxx")
		Expect(err).Should(BeNil())

		_, err = client.Put(ctx, "/dcf11/d3", "xxxxxxxxxx")
		Expect(err).Should(BeNil())

		tmp, err := client.Grant(ctx, 30)
		Expect(err).Should(BeNil())

		_, err = client.Put(ctx, fmt.Sprintf("/dcf11/configs/keepalive"), "true", clientv3.WithLease(tmp.ID))
		Expect(err).Should(BeNil())

		resp, err := client.Get(ctx, "/dcf11", clientv3.WithPrefix())
		Expect(err).Should(BeNil())
		for _, value := range resp.Kvs {
			if nil == value {
				continue
			}
			fmt.Println(value)
		}

		time.Sleep(time.Second * 5)
	})
})
