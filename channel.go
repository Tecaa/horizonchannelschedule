package horizonchannelschedule

import "time"

type Channel struct {
	Name     string
	Schedule []ChannelEntry
}

type ChannelEntry struct {
	Start       time.Time
	End         time.Time
	ProgramName string
}
