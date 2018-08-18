package main

import (
	"os"
	"time"

	"github.com/alecthomas/kingpin"
)

var (
	ipcheckServer = kingpin.Flag("ipcheck-server", "http server for checking global IP address.").Default("http://ipcheck.ieserver.net").URL()
	targetDomain  = kingpin.Arg("domain", "the domain for updating.").Required().String()
	masterID      = kingpin.Arg("master-id", "master ID for login to MyDNS.").Required().String()
	password      = kingpin.Arg("password", "password for login to MyDNS.").Required().String()
	statusFile    = kingpin.Flag("status-file", "file for storing status information.").Default("/var/lib/updns.json").String()
	interval      = kingpin.Flag("interval", "update interval when IP address does not updated.").Short('i').Default("24h").Duration()
)

func main() {
	kingpin.Parse()

	var reporter Reporter = NewLogReporter(*targetDomain)

	info, err := LoadOrMakeStatusInfo(*statusFile)
	if err != nil {
		reporter.FailedToLoadStatusInfo(err.Error())
	}

	reporter.LastUpdated(info.LastUpdated)

	currentAddress, _ := GetCurrentAddress(*targetDomain)
	reporter.CurrentAddress(currentAddress)

	realAddress, err := GetRealAddress(*ipcheckServer)
	if realAddress == "unknown" {
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress)

	if currentAddress != realAddress || info.LastUpdated.Add(*interval).Before(time.Now()) {
		var dnsserver DNSServer = NewMyDNSServer(*targetDomain, *masterID, *password)
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
