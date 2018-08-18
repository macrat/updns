package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type StatusInfo struct {
	path        string
	LastUpdated time.Time `json:"last_updated"`
}

func LoadOrMakeStatusInfo(path string) (info *StatusInfo, err error) {
	info = new(StatusInfo)
	info.path = path

	bytes, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(bytes, &info)
	} else if os.IsNotExist(err) {
		err = nil
	}

	return
}

func (info *StatusInfo) Save() error {
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(info.path, bytes, 0644)
}

func (info *StatusInfo) Updated() {
	info.LastUpdated = time.Now()
}
