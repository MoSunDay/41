package utils

import (
	"net"
	"strings"
)

var confLogger = GetLogger("network")

func GetLocalIpV4(card string) string {
	inters, err := net.Interfaces()
	if err != nil {
		confLogger.Panic(err)
	}
	ip := ""
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && strings.HasPrefix(inter.Name, card) {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if net, ok := addr.(*net.IPNet); ok {
					if net.IP.To4() != nil {
						ip = net.IP.String()
					}
				}
			}
		}
	}
	if ip == "" {
		confLogger.Panic("can't found local ip")
	}
	return ip
}
