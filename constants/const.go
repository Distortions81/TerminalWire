package constants

import "time"

const Version = "0.203-02-17-2021-1004p"
const Unknown = "Unknown"

//Throttle to about 5 every 6 seconds
const CMSRate = 500 * time.Millisecond
const CMSRestTime = 6000 * time.Millisecond
const CMSPollRate = 100 * time.Millisecond

//Move to new CFG when done
const ServersRoot = "/home/dist/Desktop/fact/"
const ServersPrefix = "cw-"
const CWGlobalConfig = "cw-global-config.json"
const CWLocalConfig = "cw-local-config.json"
const CWChannelID = "811840750722744360"
const HostIP = "m45sci.xyz"
