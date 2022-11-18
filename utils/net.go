package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

var privateIPNets = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"100.64.0.0/10",
	"fd00::/8",
}

func IsPrivateIP(ip net.IP) bool {
	for _, cidr := range privateIPNets {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func GetClientIP(r *http.Request) string {
	reqIP := r.Header.Get("X-Real-IP")
	if reqIP == "" {
		h, _, _ := net.SplitHostPort(r.RemoteAddr)
		reqIP = h
	}
	return reqIP
}

func GetLocationByIP(ip string) string {
	if IsPrivateIP(net.ParseIP(ip)) {
		return "LAN"
	}
	resp, err := http.Get(fmt.Sprintf("http://whois.pconline.com.cn/ipJson.jsp?json=true&ip=%s", ip))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	info := map[string]string{}
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s %s", info["pro"], info["city"])
}
