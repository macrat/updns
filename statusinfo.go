package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type StatusInfo struct {
	path            string
	domain          string
	LastUpdated     time.Time `json:"last_updated"`
	MinorErrorCount int
	FatalErrorCount int
	ExecutedCount   int
}

type StatusInfoFile map[string]StatusInfo

func LoadStatusInfoFile(path string) (info StatusInfoFile, err error) {
	bytes, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(bytes, &info)
	} else if os.IsNotExist(err) {
		err = nil
	}

	return info, err
}

func LoadOrMakeStatusInfo(path, domain string) (info *StatusInfo, err error) {
	raw, err := LoadStatusInfoFile(path)
	if err != nil {
		return &StatusInfo{path: path, domain: domain}, err
	}

	i := raw[domain]
	info = &i
	info.path = path
	info.domain = domain

	return info, nil
}

func (info *StatusInfo) Save() error {
	data, err := LoadStatusInfoFile(info.path)
	if err != nil {
		return err
	}
	if data == nil {
		data = make(StatusInfoFile)
	}

	data[info.domain] = *info

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(info.path, bytes, 0644)
}

func (info *StatusInfo) Updated() {
	info.LastUpdated = time.Now()
}
