package constants

import "time"

const Version = "0.202-111820200641a"
const Unknown = "Unknown"

//Throttle to about 5 every 6 seconds
const CMSRate = 500 * time.Millisecond
const CMSRestTime = 6000 * time.Millisecond
const CMSPollRate = 100 * time.Millisecond
