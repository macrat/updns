package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bogdanovich/dns_resolver"
)

func GetCurrentAddress(domain string) (address string, err error) {
	resolver := dns_resolver.New([]string{"8.8.8.8", "4.4.4.4"})
	if addrs, err := resolver.LookupHost(domain); err == nil {
		return addrs[0].String(), nil
	} else {
		return "unknown", err
	}
}

func GetRealAddress(ipcheckServer string) (address string, err error) {
	resp, err := http.Get(ipcheckServer)
	if err != nil {
		return "unknown", err
	}
	defer resp.Body.Close()

	if addr, err := ioutil.ReadAll(resp.Body); err != nil {
		return "unknown", err
	} else {
		return strings.TrimSpace(string(addr)), nil
	}
}
