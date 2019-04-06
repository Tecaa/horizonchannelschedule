package horizonchannelschedule

type ChannelModel struct {
	Id                    string
	Active                bool
	Fullname              string
	Logo                  string
	Url                   string
	Thumbnail             string
	Schedule              ChannelScheduleModel
	ScheduleCloudFunction string
}

type ChannelScheduleModel struct {
	Yesterday []ChannelEntry `json:"yesterday"`
	Today     []ChannelEntry `json:"today"`
	Tomorrow  []ChannelEntry `json:"tomorrow"`
}
