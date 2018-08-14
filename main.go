package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bogdanovich/dns_resolver"
	"github.com/go-kit/kit/log"
)

var (
	targetDomain  = "blanktar.jp"
	ipcheckServer = "http://ipcheck.ieserver.net"
)

type Reporter interface {
	CurrentAddress(address string)
	RealAddress(address string)
	FailedToGetRealAddress(message string)
}

type LogReporter struct {
	logger log.Logger
}

func NewLogReporter() LogReporter {
	return LogReporter{log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))}
}

func (reporter LogReporter) CurrentAddress(address string) {
	reporter.logger.Log("level", "info", "current_address", address)
}

func (reporter LogReporter) RealAddress(address string) {
	reporter.logger.Log("level", "info", "real_address", address)
}

func (reporter LogReporter) FailedToGetRealAddress(message string) {
	reporter.logger.Log("level", "fatal", "message", message)
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
	var reporter Reporter = NewLogReporter()

	currentAddress, _ := GetCurrentAddress(targetDomain)
	reporter.CurrentAddress(currentAddress)

	realAddress, err := GetRealAddress(ipcheckServer)
	if realAddress == "unknown" {
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress)
}
