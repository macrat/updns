package main

import (
	"fmt"
	"net/http"
)

type DNSServer interface {
	Update(address string) error
}

type MyDNSServer struct {
	domain   string
	masterID string
	password string
}

func NewMyDNSServer(domain, masterID, password string) DNSServer {
	return MyDNSServer{domain, masterID, password}
}

func (mydns MyDNSServer) Update(address string) error {
	req, err := http.NewRequest("GET", "https://www.mydns.jp/login.html", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(mydns.masterID, mydns.password)

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
