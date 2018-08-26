package main

import (
	"net"
	"strconv"
	"time"

	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
)

func NewLogger(driver, server string, level logrus.Level, fields logrus.Fields) *logrus.Logger {
	logger := logrus.New()

	logger.SetLevel(level)

	logger.WithFields(fields)

	if driver != "stdout" {
		if hook, err := Drivers[driver](server, level); err != nil {
			logger.WithFields(logrus.Fields{
				"reason":         err.Error(),
				"logging_driver": driver,
				"logging_server": server,
			}).Fatal("failed to prepare logging driver")
		} else {
			logger.AddHook(hook)
		}
	}

	logger.WithFields(logrus.Fields{
		"driver": driver,
		"server": server,
	}).Debug("prepared logging driver")

	return logger
}

var (
	Drivers = map[string]Driver{
		"fluent": fluentDriver,
	}
)

func parseAddress(server string) (string, int, error) {
	if host, port, err := net.SplitHostPort(server); err != nil {
		return "", 0, err
	} else if p, err := strconv.Atoi(port); err != nil {
		return "", 0, err
	} else {
		return host, p, nil
	}
}

func makeLevels(level logrus.Level) (levels []logrus.Level) {
	switch level {
	case logrus.FatalLevel:
		levels = append(levels, logrus.FatalLevel)
		fallthrough
	case logrus.ErrorLevel:
		levels = append(levels, logrus.ErrorLevel)
		fallthrough
	case logrus.WarnLevel:
		levels = append(levels, logrus.WarnLevel)
		fallthrough
	case logrus.InfoLevel:
		levels = append(levels, logrus.InfoLevel)
		fallthrough
	case logrus.DebugLevel:
		levels = append(levels, logrus.DebugLevel)
	}

	return levels
}

type Driver func(server string, level logrus.Level) (logrus.Hook, error)

func timeFormat(raw interface{}) interface{} {
	if t, ok := raw.(time.Time); ok {
		return t.Format("2006-01-02T15:04:05-0700")
	}
	return raw
}

func durationFormat(raw interface{}) interface{} {
	if d, ok := raw.(time.Duration); ok {
		return float64(int64(d)) / 1000.0 / 1000.0 / 1000.0
	}
	return raw
}

func fluentDriver(server string, level logrus.Level) (logrus.Hook, error) {
	host, port, err := parseAddress(server)
	if err != nil {
		return nil, err
	}

	logger, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Host:                host,
		Port:                port,
		LogLevels:           makeLevels(level),
		DefaultTag:          "updns",
		DefaultMessageField: "message",
	})
	if err != nil {
		return nil, err
	}

	logger.AddFilter("last_updated", timeFormat)
	logger.AddFilter("timestamp", timeFormat)
	logger.AddFilter("taken", durationFormat)

	return logger, nil
}
