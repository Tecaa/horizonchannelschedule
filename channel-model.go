package horizonchannelschedule

type ChannelModel struct {
	Id                    string
	Active                bool
	Fullname              string
	Logo                  string
	Url                   string
	Thumbnail             string
	Schedule              ChanelScheduleModel
	ScheduleCloudFunction string
}

type ChanelScheduleModel struct {
	Yesterday []ChannelEntry
	Today     []ChannelEntry
	Tommorrow []ChannelEntry
}
