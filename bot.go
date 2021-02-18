package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Distortions81/rcon"
	"github.com/bwmarrin/discordgo"

	"./cfg"
	"./disc"
	"./glob"
	"./logs"
	"./platform"
)

func err_handler(err error) {
	logs.Log(fmt.Sprintf("Error: `%v`\n", err))
}

func main() {
	t := time.Now()

	//Create our log file name
	glob.BotLogName = fmt.Sprintf("log/bot-%v.log", t.Unix())

	//Make log directory
	os.MkdirAll("log", os.ModePerm)

	//Open log files
	bdesc, errb := os.OpenFile(glob.BotLogName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	//Save descriptors, open/closed elsewhere
	glob.BotLogDesc = bdesc

	//Send stdout and stderr to our logfile, to capture panic errors and discordgo errors
	platform.CaptureErrorOut(bdesc)

	if errb != nil {
		logs.Log(fmt.Sprintf("An error occurred when attempting to create bot log. Details: %s", errb))
		os.Exit(1)
	}

	if !cfg.ReadSettings() {
		logs.Log("No bot settings found, or invalid data.")
		os.Exit(1)
		return
	}
	if !cfg.ReadGCfg() {
		logs.Log("No global server config found, or invalid data")
		os.Exit(1)
		return
	}
	if !cfg.FindAndReadLConfigs() {
		logs.Log("No server configs found, or invalid data.")
		os.Exit(1)
		return
	}

	discord, err := discordgo.New("Bot " + cfg.Settings.Token)
	if err != nil {
		logs.Log("Unable to connect to Discord!")
		os.Exit(1)
	}

	time.Sleep(2 * time.Second)

	discord.AddHandler(IncomingMessage)
	erro := discord.Open()
	if erro != nil {
		logs.Log("Unable to open session.")
		err_handler(erro)
		os.Exit(1)
	}

	glob.DS = discord
	logs.Log("Bot is ready.")
	CMS("Bot online.")

	//Channel Message Send Loop
	CMSLoop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

func IncomingMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("BORK")
	if m.Author.ID == s.State.User.ID {
		//Hello, no this is Patrick!
		return
	}

	if m.Author.Bot {
		//Don't listen to other bots...
		return
	}
	fmt.Println("A MEEP")

	//Right channel?
	if m.ChannelID == cfg.Settings.CWChannelID {
		fmt.Println("MEEP")

		args := strings.Split(m.Content, " ")
		if len(args) > 1 {
			command := strings.Join(args[1:], " ")
			server := strings.ToLower(args[0])

			//Log who, and what command
			logs.Log("~" + m.Author.Username + ": " + m.Content)

			for i, serv := range cfg.Local {
				if strings.EqualFold(serv.ServerCallsign, server) || (strings.EqualFold(server, "all") && serv.ServerCallsign != "") {
					if command == "" {
						_, err := s.ChannelMessageSend(m.ChannelID, "No command specified.")
						err_handler(err)
						return
					}

					//Force to read in order
					serv.Lock.Lock()
					serv.Waiting = true

					SendRCON(i+1, command, s)
				}
			}
			return
		}
	}
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "...(cut, max 2000 chars)"
	}
	return bnoden
}

func SendRCON(i int, command string, s *discordgo.Session) {

	serv := cfg.Local[i]
	portstr := fmt.Sprintf("%v", serv.Port+cfg.Global.RconPortOffset)
	remoteConsole, err := rcon.Dial(cfg.Settings.Host+":"+portstr, cfg.Global.RconPass)
	if err != nil || remoteConsole == nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", serv.Name, err))
		serv.Lock.Unlock()
		return
	}

	defer func() {
		serv.Lock.Unlock()
		remoteConsole.Close()
	}()

	reqID, err := remoteConsole.Write(command)
	if err != nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", serv.Name, err))
		return
	}

	resp, respReqID, err := remoteConsole.Read()
	if err != nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", serv.Name, err))
		return
	}

	if reqID != respReqID {
		log.Println("Invalid response ID.")
		return
	}

	CMS(fmt.Sprintf("**%v:**\n```%v```", serv.Name, resp))
}

func CMS(text string) {

	//Split at newlines, so we can batch neatly
	lines := strings.Split(text, "\n")

	glob.CMSBufferLock.Lock()

	for _, line := range lines {

		if len(line) <= 2000 {
			var item glob.CMSBuf
			item.Text = line

			glob.CMSBuffer = append(glob.CMSBuffer, item)
			logs.Log("~" + line)
		} else {
			logs.Log("CMS: Line too long! Discarding...")
		}
	}

	glob.CMSBufferLock.Unlock()
}

func CMSLoop() {
	//*******************************
	//CMS Output from buffer, batched
	//*******************************
	go func() {
		for {

			if glob.DS != nil {

				//Check if buffer is active
				active := false
				glob.CMSBufferLock.Lock()
				if glob.CMSBuffer != nil {
					active = true
				}
				glob.CMSBufferLock.Unlock()

				//If buffer is active, sleep and wait for it to fill up
				if active {
					time.Sleep(time.Duration(cfg.Settings.CMSRate) * time.Millisecond)

					//Waited for buffer to fill up, grab and clear buffers
					glob.CMSBufferLock.Lock()
					lcopy := glob.CMSBuffer
					glob.CMSBuffer = nil
					glob.CMSBufferLock.Unlock()

					if lcopy != nil {

						var factmsg []string

						for _, msg := range lcopy {
							factmsg = append(factmsg, msg.Text)
						}

						//Send out buffer, split up if needed
						buf := ""
						for _, line := range factmsg {
							oldlen := len(buf) + 1
							addlen := len(line)
							if oldlen+addlen >= 2000 {
								disc.SmartWriteDiscord(cfg.Settings.CWChannelID, buf)
								buf = line
							} else {
								buf = buf + "\n" + line
							}
						}
						if buf != "" {
							disc.SmartWriteDiscord(cfg.Settings.CWChannelID, buf)
						}
					}

					//Don't send any more messages for a while (throttle)
					time.Sleep(time.Duration(cfg.Settings.CMSRestTime) * time.Millisecond)
				}

			}

			//Sleep for a moment before checking buffer again
			time.Sleep(time.Duration(cfg.Settings.CMSPollRate) * time.Millisecond)
		}
	}()
}
