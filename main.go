package main

import (
	"os"
)

var (
	ipcheckServer = "http://ipcheck.ieserver.net"
	targetDomain  = "target.domain"
	masterID      = "master ID of MyDNS"
	password      = "password of MyDNS"
	statusFile    = "/var/lib/updns.json"
)

func main() {
	var reporter Reporter = NewLogReporter(targetDomain)

	info, err := LoadOrMakeStatusInfo(statusFile)
	if err != nil {
		reporter.FailedToLoadStatusInfo(err.Error())
	}

	currentAddress, _ := GetCurrentAddress(targetDomain)
	reporter.CurrentAddress(currentAddress)

	realAddress, err := GetRealAddress(ipcheckServer)
	if realAddress == "unknown" {
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress)

	if currentAddress != realAddress {
		var dnsserver DNSServer = NewMyDNSServer(targetDomain, masterID, password)
		if err = dnsserver.Update(realAddress); err != nil {
			reporter.FailedToUpdate(err.Error())
			os.Exit(1)
		}

		info.Updated()

		if err = info.Save(); err != nil {
			reporter.FailedToSaveStatusInfo(err.Error())
		}

		reporter.Updated(info.LastUpdated)
	}
}
