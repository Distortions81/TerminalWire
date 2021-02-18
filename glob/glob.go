package glob

import (
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var CMSBuffer []CMSBuf
var CMSBufferLock sync.Mutex

var DS *discordgo.Session
var BotLogName = ""
var BotLogDesc *os.File
