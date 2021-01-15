#!/bin/bash

function ontrap() {
	echo "Trapped."
	exit 0
}

trap ontrap SIGINT

pkill -f './TerminalWire'

mkdir -p tmp/
[ -f tmp/botlog ] && touch -f tmp/botlog

while true; do
	./TerminalWire &> /dev/null
	echo "Bot exited." 
	sleep 5
done
