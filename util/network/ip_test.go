package network

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Suite")
}

var _ = Describe("Network", func() {
	Specify("external ip test", func() {
		ip, err := ExternalIP()
		Expect(err).Should(BeNil())
		Expect(ip).ShouldNot(BeEmpty())
	})

	Specify("mac addr test", func() {
		mac, err := GetMacAddr()
		Expect(err).Should(BeNil())
		Expect(mac).ShouldNot(BeEmpty())
	})

	Specify("local ip test", func() {
		ip, err := LocalIP("ens33")
		Expect(err).ShouldNot(BeNil())

		inters, err := net.Interfaces()
		Expect(err).Should(BeNil())

		for _, inter := range inters {
			ip, err = LocalIP(inter.Name)
			if nil == err {
				break
			}
		}
		Expect(ip).ShouldNot(BeEmpty())
	})
})
