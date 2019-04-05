package horizonchannelschedule

type HorizonChannelPeriod struct {
	EntryCount      int
	TotalResults    int
	Updated         int64
	Expires         int64
	Title           string
	Periods         int
	PeriodStartTime int64
	PeriodEndTime   int64
	Entries         []HorizonChannelEntry
}

type HorizonChannelEntry struct {
	O string
	L []HorizonProgram
}

type HorizonProgram struct {
	T string // Program Name
	E int64  //End Hour
	S int64  //Start Hour
	A bool
	I string
	R bool
}
