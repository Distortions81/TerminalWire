package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"../constants"
	"../logs"
)

var ServerCfgs []servdata

func FindAndReadConfigs() bool {
	var servFound []string

	files, err := ioutil.ReadDir(constants.ServersRoot)

	if err != nil {

		logs.Log(err.Error())
		return false
	}

	for _, f := range files {

		if f.IsDir() && strings.Contains(f.Name(), constants.ServersPrefix) {
			servFound = append(servFound, f.Name())
			buf := fmt.Sprintf("Possible server found: %v", f.Name())
			logs.Log(buf)
		}
	}

	if servFound != nil {
		readConfigs(servFound)
	} else {
		logs.Log("No servers found!")
		return false
	}
	return true
}

type servdata struct {
	letter string
	name   string
	port   int
}

func createServData() servdata {
	data := servdata{}
	return data
}

func readConfigs(serversFound []string) bool {

	var cfglist []string

	for _, s := range serversFound {
		path := fmt.Sprintf("%v/%v%v", constants.ServersRoot, s, "/cw-local-config.json")
		_, err := os.Stat(path)
		if err == nil {
			cfglist = append(cfglist, path)
			buf := fmt.Sprintf("Server config file found: %v", path)
			logs.Log(buf)
		} else {
			buf := fmt.Sprintf("Server with no config: %v", path)
			logs.Log(buf)
		}
	}

	//Read server config
	var servlist []servdata
	var cfgread = 0
	for _, s := range cfglist {
		file, err := ioutil.ReadFile(s)
		if file != nil && err == nil {
			cfg := createServData()

			err := json.Unmarshal([]byte(file), &cfg)
			if err != nil {
				logs.Log("readConfigs: Unmashal failure")
				logs.Log(err.Error())
				return false
			}

			cfgread = cfgread + 1
			servlist = append(servlist, cfg)
			buf := fmt.Sprintf("Read Config: %v", s)
			logs.Log(buf)

		} else {
			logs.Log("readConfigs: ReadFile failure")
			return false
		}
	}
	if cfgread > 0 {
		buf := fmt.Sprintf("%v config files read.", cfgread)
		ServerCfgs = servlist

		return true
		logs.Log(buf)
	} else {
		logs.Log("Unable to find or read any server config files!")
	}
	return false
}
