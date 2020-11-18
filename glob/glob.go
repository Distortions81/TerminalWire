package glob

import (
	"os"
	"sync"
	"time"

	"../def"
	"github.com/bwmarrin/discordgo"
)

type server struct {
	Lock sync.RWMutex `json:"-"`

	CmdName string `json:",omitempty"`
	Name    string `json:",omitempty"`
	Host    string `json:",omitempty"`
	Port    string `json:",omitempty"`
	Pass    string `json:",omitempty"`

	Response bool
	Waiting  bool `json:"-"`
}

type servers struct {
	Token     string
	ChannelID string
	Servers   [def.MaxServers]server `json:",omitempty"`
}

var ServerList servers
var NumServers = 0

type CMSBuf struct {
	Added time.Time
	Text  string
}

var CMSBuffer []CMSBuf
var CMSBufferLock sync.Mutex

var DS *discordgo.Session
var BotLogName = ""
var BotLogDesc *os.File
