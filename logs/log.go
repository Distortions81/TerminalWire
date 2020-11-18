package logs

import (
	"fmt"
	"time"

	"../glob"
)

func Log(text string) {

	t := time.Now()
	date := fmt.Sprintf("%02d-%02d-%04d_%02d-%02d-%02d", t.Month(), t.Day(), t.Year(), t.Hour(), t.Minute(), t.Second())

	buf := fmt.Sprintf("%s %s", date, text)
	glob.BotLogDesc.WriteString(buf + "\n")
}
