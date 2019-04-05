package horizonchannelschedule

type HorizonGuide struct {
	EntryCount   int
	TotalResults int
	Updated      int64
	Expires      int64
	Title        string
	Channels     []HorizonChannelInfo
}

type HorizonChannelInfo struct {
	Id               string
	Title            string
	LocationId       string
	HasLiveStream    bool
	StationSchedules []StationSchedule
}

type StationSchedule struct {
	Station Station
}

type Station struct {
	Id                 string
	LocationId         string
	Title              string
	IsHd               bool
	ServiceId          string
	ConcurrencyLimit   int
	IsOutOfHomeEnabled bool
}
