package main

import (
	"os"
	"time"
)

var (
	ipcheckServer = "http://ipcheck.ieserver.net"
	targetDomain  = "target.domain"
	masterID      = "master ID of MyDNS"
	password      = "password of MyDNS"
	statusFile    = "/var/lib/updns.json"
	interval      = time.Hour * 24
)

func main() {
	var reporter Reporter = NewLogReporter(targetDomain)

	info, err := LoadOrMakeStatusInfo(statusFile)
	if err != nil {
		reporter.FailedToLoadStatusInfo(err.Error())
	}

	reporter.LastUpdated(info.LastUpdated)

	currentAddress, _ := GetCurrentAddress(targetDomain)
	reporter.CurrentAddress(currentAddress)

	realAddress, err := GetRealAddress(ipcheckServer)
	if realAddress == "unknown" {
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress)

	if currentAddress != realAddress || info.LastUpdated.Add(interval).Before(time.Now()) {
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
