#!/bin/bash

function ontrap() {
	echo "Trapped."
	exit 0
}

trap ontrap SIGINT

pkill -f './rcon-bot-bhmm'

mkdir -p tmp/
[ -f tmp/botlog ] && touch -f tmp/botlog

while true; do
	./rcon-bot-bhmm | tee -a tmp/botlog &> /dev/null
	echo "Bot exited." | tee -a tmp/botlog
	sleep 5
done
