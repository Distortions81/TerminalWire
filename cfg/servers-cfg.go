package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"../constants"
	"../logs"
)

var Local []config
var Global gconfig
var ConfigLock sync.Mutex

type config struct {
	Version string

	ServerCallsign string
	Name           string
	Port           int

	MapPreset    string
	MapGenPreset string

	AutoStart         bool
	AutoUpdate        bool
	UpdateFactExp     bool
	ResetScheduleText string
	WriteStatsDisc    bool
	ResetPingString   string

	ChannelData    ChannelDataStruct
	SlowConnect    SlowConnectStruct
	SoftModOptions SoftModOptionsStruct
}

type gconfig struct {
	Version string

	RconPortOffset int
	RconPass       string

	DiscordData    DiscordDataStruct
	AdminData      AdminData
	RoleData       RoleDataStruct
	PathData       PathDataStruct
	MapPreviewData MapPreviewDataStruct

	DiscordCommandPrefix string
	ResetPingString      string
}

type AdminData struct {
	IDs   []string
	Names []string
}

//Global
//bor = based on root
//boh = based on home
//ap = absolute path
type PathDataStruct struct {
	FactorioServersRoot string //root of factorio server
	FactorioHomePrefix  string //per-server
	ChatWireHomePrefix  string //per-server
	FactorioBinary      string

	RecordPlayersFilename string //boh
	SaveFilePath          string //boh

	ScriptInserterPath string //bor
	DBFileName         string //bor
	LogCompScriptPath  string //bor
	FactUpdaterPath    string //bor
	FactUpdateCache    string //bor
	MapGenPath         string //bor

	MapPreviewPath   string //ap
	MapArchivePath   string //ap
	ImageMagickPath  string //ap
	ShellPath        string //ap
	FactUpdaterShell string //ap
	ZipBinaryPath    string //ap
	MapPreviewURL    string
	ArchiveURL       string
}

type DiscordDataStruct struct {
	Token   string
	GuildID string

	StatTotalChannelID    string
	StatMemberChannelID   string
	StatRegularsChannelID string

	ReportChannelID   string
	AnnounceChannelID string
}

type RoleDataStruct struct {
	Admins   string
	Regulars string
	Members  string
}

type MapPreviewDataStruct struct {
	Args       string
	Res        string
	Scale      string
	JPGQuality string
	JPGScale   string
}

//Local
type ChannelDataStruct struct {
	Pos    int
	ChatID string
	LogID  string
}

type SlowConnectStruct struct {
	SlowConnect  bool
	DefaultSpeed float32
	ConnectSpeed float32
}

type SoftModOptionsStruct struct {
	DoWhitelist    bool
	RestrictMode   bool
	FriendlyFire   bool
	CleanMapOnBoot bool
}

func ReadGCfg() bool {

	_, err := os.Stat(Settings.CWGlobalConfig)
	notfound := os.IsNotExist(err)

	if notfound {
		logs.Log("ReadGCfg: os.Stat failed")
		return false

	} else {

		file, err := ioutil.ReadFile(Settings.CWGlobalConfig)

		if file != nil && err == nil {
			cfg := CreateGCfg()

			err := json.Unmarshal([]byte(file), &cfg)
			if err != nil {
				logs.Log("ReadGCfg: Unmashal failure")
				logs.Log(err.Error())
				os.Exit(1)
			}

			Global = cfg

			return true
		} else {
			logs.Log("ReadGCfg: ReadFile failure")
			return false
		}
	}
}

func CreateGCfg() gconfig {
	cfg := gconfig{Version: "0.0.1"}
	return cfg
}

func CreateLCfg() config {
	cfg := config{Version: "0.0.1"}
	return cfg
}

func FindAndReadLConfigs() bool {
	var servFound []string

	files, err := ioutil.ReadDir(Global.PathData.FactorioServersRoot)

	if err != nil {

		logs.Log(err.Error())
		return false
	}

	for _, f := range files {

		if f.IsDir() && strings.Contains(f.Name(), Global.PathData.ChatWireHomePrefix) {
			servFound = append(servFound, f.Name())
			if constants.Debug {
				buf := fmt.Sprintf("Possible server found: %v", f.Name())
				logs.Log(buf)
			}
		}
	}

	if servFound != nil {
		ReadLConfigs(servFound)
	} else {
		logs.Log("No servers found!")
		return false
	}
	return true
}

func ReadLConfigs(serversFound []string) bool {

	var cfglist []string

	for _, s := range serversFound {
		path := fmt.Sprintf("%v%v/%v", Global.PathData.FactorioServersRoot, s, Settings.CWLocalConfig)
		_, err := os.Stat(path)
		if err == nil {
			cfglist = append(cfglist, path)
			if constants.Debug {
				buf := fmt.Sprintf("Server config file found: %v", path)
				logs.Log(buf)
			}
		} else {
			buf := fmt.Sprintf("Server with no config: %v", path)
			logs.Log(buf)
		}
	}

	//Read server config
	var servlist []config
	var cfgread = 0
	for _, s := range cfglist {
		file, err := ioutil.ReadFile(s)
		if file != nil && err == nil {
			cfg := CreateLCfg()

			err := json.Unmarshal([]byte(file), &cfg)
			if err != nil {
				buf := fmt.Sprintf("readConfigs: Unmashal failure for file %v", s)
				logs.Log(buf)
				logs.Log(err.Error())
				return false
			}

			cfgread = cfgread + 1
			servlist = append(servlist, cfg)
			buf := fmt.Sprintf("Read Config: %v", s)
			logs.Log(buf)

		} else {
			buf := fmt.Sprintf("readConfigs: ReadFile failure for file: %v", s)
			logs.Log(buf)
			logs.Log("readConfigs: ReadFile failure")
			return false
		}
	}
	if cfgread > 0 {
		buf := fmt.Sprintf("%v config files read.", cfgread)
		logs.Log(buf)
		Local = servlist

		return true
	} else {
		logs.Log("Unable to find or read any server config files!")
	}
	return true
}
