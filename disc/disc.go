package disc

import (
	"fmt"
	"time"

	"../glob"
	"../logs"
)

func SmartWriteDiscord(ch string, text string) {

	if glob.DS != nil {
		_, err := glob.DS.ChannelMessageSend(ch, text)

		if err != nil {

			//time.Sleep(time.Second)
			//SmartWriteDiscord(ch, text)
			logs.Log(fmt.Sprintf("SmartWriteDiscord: ERROR: %v", err))
		}
	} else {

		time.Sleep(5 * time.Second)
		SmartWriteDiscord(ch, text)
	}
}
