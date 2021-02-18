package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Distortions81/rcon"
	"github.com/bwmarrin/discordgo"

	"./cfg"
	"./constants"
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

	if !cfg.FindAndReadConfigs() {
		logs.Log("No server configs found.")
		return
	}

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

	discord, err := discordgo.New("Bot " + glob.ServerList.Token)
	if err != nil {
		os.Exit(1)
	}

	discord.AddHandler(IncomingMessage)
	erro := discord.Open()
	if erro != nil {
		err_handler(erro)
		os.Exit(1)
	}

	glob.DS = discord
	logs.Log("Bot is ready.")

	//Channel Message Send Loop
	CMSLoop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

func IncomingMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		//Hello, no this is Patrick!
		return
	}

	if m.Author.Bot {
		//Don't listen to other bots...
		return
	}

	//Right channel?
	if m.ChannelID == glob.ServerList.ChannelID {

		args := strings.Split(m.Content, " ")
		if len(args) > 1 {
			command := strings.Join(args[1:], " ")
			server := strings.ToLower(args[0])

			//Log who, and what command
			logs.Log("~" + m.Author.Username + ": " + m.Content)

			for i := 0; i < glob.NumServers; i++ {
				if strings.EqualFold(glob.ServerList.Servers[i].CmdName, server) || (strings.EqualFold(server, "all") && glob.ServerList.Servers[i].CmdName != "") {
					if command == "" {
						_, err := s.ChannelMessageSend(m.ChannelID, "No command specified.")
						err_handler(err)
						return
					}
					/*
						time.Sleep(100 * time.Millisecond)
						go func(pos int) {
							glob.ServerList.Servers[pos].Lock.Lock()
							glob.ServerList.Servers[pos].Waiting = true

							SendRCON(pos, command, s)
						}(i)*/

					//Force to read in order
					glob.ServerList.Servers[i].Lock.Lock()
					glob.ServerList.Servers[i].Waiting = true

					SendRCON(i, command, s)
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

	remoteConsole, err := rcon.Dial(glob.ServerList.Servers[i].Host+":"+glob.ServerList.Servers[i].Port, glob.ServerList.Servers[i].Pass)
	if err != nil || remoteConsole == nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", glob.ServerList.Servers[i].Name, err))
		glob.ServerList.Servers[i].Lock.Unlock()
		return
	}

	defer func() {
		glob.ServerList.Servers[i].Lock.Unlock()
		remoteConsole.Close()
	}()

	reqID, err := remoteConsole.Write(command)
	if err != nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", glob.ServerList.Servers[i].Name, err))
		return
	}

	resp, respReqID, err := remoteConsole.Read()
	if err != nil {
		err_handler(err)
		CMS(fmt.Sprintf("%v: Error: `%v`", glob.ServerList.Servers[i].Name, err))
		return
	}

	if reqID != respReqID {
		log.Println("Invalid response ID.")
		return
	}

	CMS(fmt.Sprintf("**%v:**\n```%v```", glob.ServerList.Servers[i].Name, resp))
}

func ReadConfig() bool {

	_, err := os.Stat(constants.DATA_DIR + constants.CONFIG_FILE)
	notfound := os.IsNotExist(err)

	if notfound {
		err_handler(err)
		log.Println("Config file not found!")
		return false

	} else {

		file, err := ioutil.ReadFile(constants.DATA_DIR + constants.CONFIG_FILE)

		if file != nil && err == nil {
			err := json.Unmarshal([]byte(file), &glob.ServerList)
			if err != nil {
				err_handler(err)
			}

			log.Println("Config loaded.")
			return true
		} else {
			err_handler(err)
			return false
		}
	}
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
					time.Sleep(constants.CMSRate)

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
								disc.SmartWriteDiscord(glob.ServerList.ChannelID, buf)
								buf = line
							} else {
								buf = buf + "\n" + line
							}
						}
						if buf != "" {
							disc.SmartWriteDiscord(glob.ServerList.ChannelID, buf)
						}
					}

					//Don't send any more messages for a while (throttle)
					time.Sleep(constants.CMSRestTime)
				}

			}

			//Sleep for a moment before checking buffer again
			time.Sleep(constants.CMSPollRate)
		}
	}()
}
