package glob

import (
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CMSBuf struct {
	Added time.Time
	Text  string
}

var CMSBuffer []CMSBuf
var CMSBufferLock sync.Mutex

var DS *discordgo.Session
var BotLogName = ""
var BotLogDesc *os.File
