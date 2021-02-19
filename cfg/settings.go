package cfg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"../constants"
	"../logs"
)

var Settings settings

type settings struct {
	Version     string
	Token       string
	CWChannelID string

	CMSRate     int
	CMSRestTime int
	CMSPollRate int

	CWGlobalConfig string
	CWLocalConfig  string

	Host string
}

func CreateSettings() settings {
	cfg := settings{Version: "0.0.1"}
	return cfg
}

func WriteSettings() bool {
	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(Settings); err != nil {
		logs.Log("WriteSettings: enc.Encode failure")
		return false
	}

	_, err := os.Create(constants.SettingsName)

	if err != nil {
		logs.Log("WriteSettings: os.Create failure")
		return false
	}

	err = ioutil.WriteFile(constants.SettingsName, []byte(outbuf.String()), 0644)

	if err != nil {
		logs.Log("WriteSettings: WriteFile failure")
	}

	return true
}

func ReadSettings() bool {

	_, err := os.Stat(constants.SettingsName)
	notfound := os.IsNotExist(err)

	if notfound {
		logs.Log("ReadSettings: os.Stat failed")
		return false

	} else {

		file, err := ioutil.ReadFile(constants.SettingsName)

		if file != nil && err == nil {
			cfg := CreateSettings()

			err := json.Unmarshal([]byte(file), &cfg)
			if err != nil {
				logs.Log("ReadSettings: Unmashal failure")
				logs.Log(err.Error())
				os.Exit(1)
			}

			Settings = cfg

			return true
		} else {
			logs.Log("ReadSettings: ReadFile failure")
			return false
		}
	}
}
