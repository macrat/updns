package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bogdanovich/dns_resolver"
	"github.com/go-kit/kit/log"
)

var (
	ipcheckServer = "http://ipcheck.ieserver.net"
	targetDomain  = "target.domain"
	masterID      = "master ID of MyDNS"
	password      = "password of MyDNS"
)

type Reporter interface {
	CurrentAddress(address string)
	RealAddress(address string)
	FailedToGetRealAddress(message string)
	Updated()
	FailedToUpdate(message string)
}

type LogReporter struct {
	logger log.Logger
}

func NewLogReporter(targetDomain string) LogReporter {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger.Log("level", "info", "target_domain", targetDomain)
	return LogReporter{logger}
}

func (reporter LogReporter) CurrentAddress(address string) {
	reporter.logger.Log("level", "info", "current_address", address)
}

func (reporter LogReporter) RealAddress(address string) {
	reporter.logger.Log("level", "info", "real_address", address)
}

func (reporter LogReporter) FailedToGetRealAddress(reason string) {
	reporter.logger.Log("level", "fatal", "error", "failed to get real IP address", "reason", reason)
}

func (reporter LogReporter) Updated() {
	reporter.logger.Log("level", "info", "message", "updated")
}

func (reporter LogReporter) FailedToUpdate(reason string) {
	reporter.logger.Log("level", "fatal", "error", "failed to update", "reason", reason)
}

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

func main() {
	var reporter Reporter = NewLogReporter(targetDomain)

	currentAddress, _ := GetCurrentAddress(targetDomain)
	reporter.CurrentAddress(currentAddress)

	realAddress, err := GetRealAddress(ipcheckServer)
	if realAddress == "unknown" {
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress)

	if currentAddress != realAddress {
		dnsserver := NewMyDNSServer(targetDomain, masterID, password)
		if err := dnsserver.Update(realAddress); err != nil {
			reporter.FailedToUpdate(err.Error())
		} else {
			reporter.Updated()
		}
	}
}
