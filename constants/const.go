package constants

import "time"

const Version = "0.203-02-17-2021-1004p"
const Unknown = "Unknown"

//Throttle to about 5 every 6 seconds
const CMSRate = 500 * time.Millisecond
const CMSRestTime = 6000 * time.Millisecond
const CMSPollRate = 100 * time.Millisecond

//Config files
const ServersRoot = "/home/fact/"
const ServersPrefix = "fact-"
const CWGlobalConfig = "cw-global-config.json"
const CWLocalConfig = "cw-local-config.json"
