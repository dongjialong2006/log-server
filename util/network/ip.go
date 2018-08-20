package network

import (
	"fmt"
	"net"
)

func LocalIP(ifName string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return nil, err
	}

	return getAddr(*iface)
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		ip, err := getAddr(iface)
		if nil != err {
			continue
		}

		if "" != ip.String() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("get addr error.")
}

func getAddr(iface net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil || ip.IsLoopback() {
			continue
		}

		ip = ip.To4()
		if len(ip) == 0 {
			continue
		}

		return ip, nil
	}

	return nil, fmt.Errorf("ip not found.")
}

func GetMacAddr() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr.String()
		if "" != mac {
			return mac, nil
		}
	}

	return "", fmt.Errorf("mac not found.")
}
