package horizonchannelschedule

import "time"

type Channel struct {
	Name     string
	Schedule []ChannelEntry
}

type ChannelEntry struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	ProgramName string    `json:"programName"`
}
