package main

import (
	"os"
	"time"

	"github.com/go-kit/kit/log"
)

type Reporter interface {
	FailedToLoadStatusInfo(reason string)
	LastUpdated(timestamp time.Time)
	CurrentAddress(address string, takenTime time.Duration)
	RealAddress(address string, takenTime time.Duration)
	FailedToGetRealAddress(reason string)
	Updated(timestamp time.Time, takenTime time.Duration)
	FailedToUpdate(reason string)
	FailedToSaveStatusInfo(reason string)
}

type LogReporter struct {
	logger log.Logger
}

func NewLogReporter(targetDomain string) LogReporter {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger.Log("level", "info", "target_domain", targetDomain)
	return LogReporter{logger}
}

func (reporter LogReporter) FailedToLoadStatusInfo(reason string) {
	reporter.logger.Log("level", "error", "error", "failed to load status info", "reason", reason)
}

func (reporter LogReporter) LastUpdated(timestamp time.Time) {
	reporter.logger.Log("level", "info", "last_updated", timestamp, "now", time.Now())
}

func (reporter LogReporter) CurrentAddress(address string, takenTime time.Duration) {
	reporter.logger.Log("level", "info", "current_address", address, "taken_time", takenTime)
}

func (reporter LogReporter) RealAddress(address string, takenTime time.Duration) {
	reporter.logger.Log("level", "info", "real_address", address, "taken_time", takenTime)
}

func (reporter LogReporter) FailedToGetRealAddress(reason string) {
	reporter.logger.Log("level", "fatal", "error", "failed to get real IP address", "reason", reason)
}

func (reporter LogReporter) Updated(timestamp time.Time, takenTime time.Duration) {
	reporter.logger.Log("level", "info", "message", "updated", "time", timestamp, "taken_time", takenTime)
}

func (reporter LogReporter) FailedToUpdate(reason string) {
	reporter.logger.Log("level", "fatal", "error", "failed to update", "reason", reason)
}

func (reporter LogReporter) FailedToSaveStatusInfo(reason string) {
	reporter.logger.Log("level", "error", "error", "failed to save status info", "reason", reason)
}
