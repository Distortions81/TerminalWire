package cfg

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"../constants"
	"../logs"
)

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
			fmt.Println(buf)
		}
	}

	if servFound != nil {
		ReadConfigs(servFound)
	} else {
		logs.Log("No servers found!")
		return false
	}
	return true
}

func ReadConfigs(serversFound []string) bool {

	found := false

	for _, s := range serversFound {
		_, err := os.Stat(constants.ServersRoot + s + "/cw-local-config.json")
		if err == nil {
			buf := fmt.Sprintf("Server config file found: %v", s)
			found = true
			fmt.Println(buf)
		}
	}

	if !found {
		fmt.Println("No servers configs found!!!")
		return false
	} else {
		return true
	}
}
