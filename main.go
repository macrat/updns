package main

import (
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
)

var (
	ipcheckServer    = kingpin.Flag("ipcheck-server", "http server for checking global IP address.").Default("http://ipcheck.ieserver.net").URL()
	targetDomain     = kingpin.Arg("domain", "the domain for updating.").Required().String()
	masterID         = kingpin.Arg("master-id", "master ID for login to MyDNS.").Required().String()
	password         = kingpin.Arg("password", "password for login to MyDNS.").Required().String()
	statusFile       = kingpin.Flag("status-file", "file for storing status information.").Default("/var/lib/updns.json").String()
	interval         = kingpin.Flag("interval", "update interval when IP address does not updated.").Short('i').Default("24h").Duration()
	prometheusServer = kingpin.Flag("prometheus-push-gateway", "prometheus push gateway server address for sending metrics").Short('p').URL()
)

func main() {
	kingpin.Parse()

	logger := log.WithFields(log.Fields{
		"domain": *targetDomain,
	})

	info, err := LoadOrMakeStatusInfo(*statusFile, *targetDomain)
	if err != nil {
		info.MinorErrorCount++
		logger.WithFields(log.Fields{
			"reason": err.Error(),
		}).Error("failed to load status info")
	}
	info.ExecutedCount++

	metrics := NewMetrics(info)

	logger.WithFields(log.Fields{
		"last_updated": info.LastUpdated,
		"file_path":    *statusFile,
	}).Info("loaded status info")

	defer func() {
		if err = info.Save(); err != nil {
			info.MinorErrorCount++
			logger.WithFields(log.Fields{
				"reason": err.Error(),
			}).Error("failed to save status info")
		}

		if *prometheusServer != nil {
			if err = PushToPrometheus(*prometheusServer, metrics); err != nil {
				logger.WithFields(log.Fields{
					"pushgateway": *prometheusServer,
					"reason":      err.Error(),
				}).Info("failed to reporting into prometheus")
			} else {
				logger.WithFields(log.Fields{
					"pushgateway": *prometheusServer,
				}).Info("reported to prometheus")
			}
		}
	}()

	stime := time.Now()
	currentAddress, err := GetCurrentAddress(*targetDomain)
	etime := time.Now()
	if err != nil {
		info.MinorErrorCount++
		logger.WithFields(log.Fields{
			"taken": etime.Sub(stime),
		}).Info("failed to get current address")
	}
	metrics.CurrentAddressTakenTime = etime.Sub(stime)
	logger.WithFields(log.Fields{
		"taken":           etime.Sub(stime),
		"current_address": currentAddress,
	}).Info("got current address")

	stime = time.Now()
	realAddress, err := GetRealAddress(*ipcheckServer)
	etime = time.Now()
	if realAddress == "unknown" {
		info.FatalErrorCount++
		logger.WithFields(log.Fields{
			"taken": etime.Sub(stime),
		}).Fatal("failed to get real address")
		os.Exit(1)
	}
	metrics.RealAddressTakenTime = etime.Sub(stime)
	logger.WithFields(log.Fields{
		"taken":        etime.Sub(stime),
		"real_address": realAddress,
	}).Info("got real address")

	if currentAddress != realAddress || info.LastUpdated.Add(*interval).Before(time.Now()) {
		var dnsserver DNSServer = NewMyDNSServer(*targetDomain, *masterID, *password)

		stime = time.Now()
		err = dnsserver.Update(realAddress)
		etime = time.Now()
		if err != nil {
			info.FatalErrorCount++
			logger.WithFields(log.Fields{
				"taken":  etime.Sub(stime),
				"reason": err.Error(),
			}).Fatal("failed to update")
			os.Exit(1)
		}
		metrics.UpdateTakenTime = etime.Sub(stime)

		info.Updated()

		logger.WithFields(log.Fields{
			"taken":     etime.Sub(stime),
			"timestamp": info.LastUpdated,
		}).Info("updated")
	}
}
