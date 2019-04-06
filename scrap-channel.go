package horizonchannelschedule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
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

func SetChannelsSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	//Local
	//opt := option.WithCredentialsFile("tv-chile-flutter-0474bef9a427.json")
	//firebaseApp, err := firebase.NewApp(ctx, nil, opt)

	//Firebase
	conf := &firebase.Config{ProjectID: "tv-chile-flutter"}
	firebaseApp, err := firebase.NewApp(ctx, conf)

	if err != nil {
		log.Fatal(err)
		panic(fmt.Errorf("error initializing app: %v", err))
	}

	firestoreClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
		panic(fmt.Errorf("Failed to create a new firestore client: %v", err))
	}

	iter := firestoreClient.Collection("channels").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(fmt.Errorf("Failed %v", err))
		}
		var channelModel ChannelModel
		doc.DataTo(&channelModel)
		channelModel.Id = doc.Ref.ID

		if channelModel.ScheduleCloudFunction == "" {
			continue
		}

		currentTime := time.Now()
		channelModel.Schedule.Today = GetScheduleFromTime(channelModel.ScheduleCloudFunction, currentTime)

		tomorrowTime := currentTime.Add(time.Hour * 24)
		channelModel.Schedule.Tomorrow = GetScheduleFromTime(channelModel.ScheduleCloudFunction, tomorrowTime)

		yesterdayTime := currentTime.Add(time.Hour * -24)
		channelModel.Schedule.Yesterday = GetScheduleFromTime(channelModel.ScheduleCloudFunction, yesterdayTime)

		_, err = firestoreClient.Collection("channels").Doc(channelModel.Id).Update(ctx, []firestore.Update{{Path: "schedule", Value: channelModel.Schedule}})
		if err != nil {
			panic(fmt.Errorf("Failed to create a new firestore client: %v", err))
		}
	}
	firestoreClient.Close()
}

func GetScheduleFromTime(scheduleCloudFunctionEndpoint string, day time.Time) []ChannelEntry {
	dateString := fmt.Sprintf("%04d%02d%02d", day.Year(), day.Month(), day.Day())
	dateEndpoint := strings.Replace(scheduleCloudFunctionEndpoint, "{date}", dateString, -1)

	channelSchedule := GetSchedule(dateEndpoint)
	return channelSchedule.Schedule
}

func GetSchedule(cloudFunctionEndpoint string) Channel {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	cloudFunctionResponse, _ := netClient.Get(cloudFunctionEndpoint)

	if cloudFunctionResponse.StatusCode != 200 {
		panic(fmt.Sprintf("cloud function endpoint response %v", cloudFunctionResponse.StatusCode))
	}

	var channel Channel
	json.NewDecoder(cloudFunctionResponse.Body).Decode(&channel)

	return channel
}

func HorizonChannelSchedule(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(r.URL.Path, "/")
	if len(p)-1 < 2 {
		fmt.Fprint(w, "Not enough params")
		return
	}

	channelIdParam := p[1]
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

func StealChannel(channelIdParam string, date string) *Channel {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	responseChannels, _ := netClient.Get("https://web-api-sugar.horizon.tv/oesp/v3/CL/spa/web/channels?byLocationId=65535&includeInvisible=true&personalised=false&sort=channelNumber")

	if responseChannels.StatusCode != 200 {
		panic(fmt.Sprintf("%v", responseChannels.StatusCode))
	}

	var horizonGuide HorizonGuide
	json.NewDecoder(responseChannels.Body).Decode(&horizonGuide)

	indexChannel := FindIndexChannel(channelIdParam, horizonGuide.Channels)
	if indexChannel == -1 {
		panic(fmt.Sprintf("Channel id not found"))
	}

	channel := new(Channel)
	channel.Name = horizonGuide.Channels[indexChannel].Title
	channel.Schedule = make([]ChannelEntry, 0)

	stationId := horizonGuide.Channels[indexChannel].StationSchedules[0].Station.Id

	channel.Schedule = AddSchedule(channel.Schedule, stationId, date)

	return channel
}

func FindIndexChannel(channelIdParam string, channels []HorizonChannelInfo) int {
	indexFound := -1
	for index, element := range channels {
		if element.Id == channelIdParam {
			indexFound = index
			break
		}
	}
	return indexFound
}

func AddSchedule(schedule []ChannelEntry, stationId string, dateString string) []ChannelEntry {
	const BASE_URL = "https://web-api-sugar.horizon.tv/oesp/v3/CL/spa/web/programschedules/%s/%d"

	urlPeriod1 := fmt.Sprintf(BASE_URL, dateString, 1)
	urlPeriod2 := fmt.Sprintf(BASE_URL, dateString, 2)
	urlPeriod3 := fmt.Sprintf(BASE_URL, dateString, 3)
	urlPeriod4 := fmt.Sprintf(BASE_URL, dateString, 4)

	schedule = AppendPeriod(schedule, stationId, urlPeriod1)
	schedule = AppendPeriod(schedule, stationId, urlPeriod2)
	schedule = AppendPeriod(schedule, stationId, urlPeriod3)
	schedule = AppendPeriod(schedule, stationId, urlPeriod4)

	return schedule
}

func AppendPeriod(schedule []ChannelEntry, stationId string, url string) []ChannelEntry {
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
		if element.O == stationId {
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
