package horizonchannelschedule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ScrapChannel(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	idInt, _ := strconv.Atoi(id)

	tvnScrapper := new(MiTvScrapper)
	tvnScrapper.Init(idInt)
	tvnChannel := tvnScrapper.Scrap()

	_ = json.NewEncoder(w).Encode(tvnChannel)
	return
}

func HorizonChannelSchedule(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(r.URL.Path, "/")
	if len(p)-1 < 2 {
		fmt.Fprint(w, "Not enough params")
		return
	}
	channelIdParam, _ := strconv.Atoi(p[1])
	dateParam := p[2]

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprint(w, "Error in internal call. Message -> ", r)
		}
	}()
	channel := StealChannel(channelIdParam, dateParam)

	b, _ := json.Marshal(channel)

	fmt.Fprint(w, string(b))
	return
}

func StealChannel(indexChannel int, date string) *Channel {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	responseChannels, _ := netClient.Get("https://web-api-sugar.horizon.tv/oesp/v3/CL/spa/web/channels?byLocationId=65535&includeInvisible=true&personalised=false&sort=channelNumber")

	if responseChannels.StatusCode != 200 {
		panic(fmt.Sprintf("%v", responseChannels.StatusCode))
	}

	horizonGuide := new(HorizonGuide)
	json.NewDecoder(responseChannels.Body).Decode(horizonGuide)

	channel := new(Channel)
	channel.Name = horizonGuide.Channels[indexChannel].Title
	channel.Schedule = make([]ChannelEntry, 0)

	channelId := horizonGuide.Channels[indexChannel].StationSchedules[0].Station.Id

	channel.Schedule = AddSchedule(channel.Schedule, channelId, date)

	return channel
}

func AddSchedule(schedule []ChannelEntry, channelId string, dateString string) []ChannelEntry {
	const BASE_URL = "https://web-api-sugar.horizon.tv/oesp/v3/CL/spa/web/programschedules/%s/%d"
	//currentTime := time.Now()
	//todayString := fmt.Sprintf("%04d%02d%02d", currentTime.Year(), currentTime.Month(), currentTime.Day())

	urlPeriod1 := fmt.Sprintf(BASE_URL, dateString, 1)
	urlPeriod2 := fmt.Sprintf(BASE_URL, dateString, 2)
	urlPeriod3 := fmt.Sprintf(BASE_URL, dateString, 3)
	urlPeriod4 := fmt.Sprintf(BASE_URL, dateString, 4)

	schedule = AppendPeriod(schedule, channelId, urlPeriod1)
	schedule = AppendPeriod(schedule, channelId, urlPeriod2)
	schedule = AppendPeriod(schedule, channelId, urlPeriod3)
	schedule = AppendPeriod(schedule, channelId, urlPeriod4)

	return schedule
}

func AppendPeriod(schedule []ChannelEntry, channelId string, url string) []ChannelEntry {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	responsePeriod, _ := netClient.Get(url)

	if responsePeriod.StatusCode != 200 {
		panic(fmt.Sprintf("%v", responsePeriod.StatusCode))
	}

	horizonChannelPeriod := new(HorizonChannelPeriod)
	json.NewDecoder(responsePeriod.Body).Decode(horizonChannelPeriod)

	indexFound := -1
	for index, element := range horizonChannelPeriod.Entries {
		if element.O == channelId {
			indexFound = index
			break
		}
	}

	if indexFound == -1 {
		panic("No se encontró programación para el canal en esta fecha.")
	}

	for index, element := range horizonChannelPeriod.Entries[indexFound].L {
		if len(schedule) != 0 && index == 0 {
			continue
		}

		channelEntry := ChannelEntry{
			ProgramName: element.T,
			Start:       time.Unix(0, element.S*int64(time.Millisecond)),
			End:         time.Unix(0, element.E*int64(time.Millisecond)),
		}
		schedule = append(schedule, channelEntry)
	}
	return schedule
}
