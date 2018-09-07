package main

import (
	"fmt"
	"net/http"
	"net/url"
)

type DNSServer interface {
	Update(address string) error
}

type BasicDNSServer struct {
	id       string
	password string
	endpoint *url.URL
}

func (dns BasicDNSServer) Update(address string) error {
	req, err := http.NewRequest("GET", dns.endpoint.String(), nil)

	if err != nil {
		return err
	}

	req.SetBasicAuth(dns.id, dns.password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to authorization")
	} else {
		return nil
	}
}

type MyDNSServer BasicDNSServer

func NewMyDNSServer(masterID, password string) DNSServer {
	return BasicDNSServer{masterID, password, &url.URL{
		Scheme:  "https",
		Host:    "www.mydns.jp",
		Path:    "/login.html",
		RawPath: "/login.html",
	}}
}
