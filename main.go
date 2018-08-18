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

	info, err := LoadOrMakeStatusInfo(*statusFile, *targetDomain)
	if err != nil {
		info.MinorErrorCount++
		reporter.FailedToLoadStatusInfo(err.Error())
	}
	info.ExecutedCount++

	defer func() {
		if err = info.Save(); err != nil {
			info.MinorErrorCount++
			reporter.FailedToSaveStatusInfo(err.Error())
		}
		reporter.Done(info)
	}()

	reporter.LastUpdated(info.LastUpdated)

	stime := time.Now()
	currentAddress, err := GetCurrentAddress(*targetDomain)
	if err != nil {
		info.MinorErrorCount++
	}
	reporter.CurrentAddress(currentAddress, time.Now().Sub(stime))

	stime = time.Now()
	realAddress, err := GetRealAddress(*ipcheckServer)
	etime := time.Now()
	if realAddress == "unknown" {
		info.FatalErrorCount++
		reporter.FailedToGetRealAddress(err.Error())
		os.Exit(1)
	}

	reporter.RealAddress(realAddress, etime.Sub(stime))

	if currentAddress != realAddress || info.LastUpdated.Add(*interval).Before(time.Now()) {
		var dnsserver DNSServer = NewMyDNSServer(*targetDomain, *masterID, *password)

		stime = time.Now()
		err = dnsserver.Update(realAddress)
		etime = time.Now()
		if err != nil {
			info.FatalErrorCount++
			reporter.FailedToUpdate(err.Error())
			os.Exit(1)
		}

		info.Updated()

		reporter.Updated(info.LastUpdated, etime.Sub(stime))
	}
}
