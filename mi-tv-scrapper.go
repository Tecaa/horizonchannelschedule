package horizonchannelschedule

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

type MiTvScrapper struct {
	IdParameter int
}

func (e *MiTvScrapper) Init(idParameter int) {
	e.IdParameter = idParameter
}

func (miTvScrapper MiTvScrapper) Scrap() *Channel {
	c := colly.NewCollector()
	channel := new(Channel)

	idParameterStr := strconv.Itoa(miTvScrapper.IdParameter)
	c.OnHTML("div.channel:nth-of-type("+idParameterStr+")", func(channelHtml *colly.HTMLElement) {
		*channel = miTvScrapper.ProcessHtml(channelHtml)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://mi.tv/cl/async/guide/all/-180")
	return channel
}

func (miTvScrapper MiTvScrapper) ProcessHtml(e *colly.HTMLElement) Channel {
	channelName := e.ChildText("span.info a h3")

	scheduleList := make([]ChannelEntry, 0)

	fn := func(a int, h *colly.HTMLElement) {
		programTime := h.ChildText("a span.time")
		programName := h.ChildText("a p span.title")

		hour, _ := time.Parse("15:04", programTime)

		channelEntry := ChannelEntry{
			ProgramName: programName,
			Start:       hour,
		}
		scheduleList = append(scheduleList, channelEntry)
	}

	e.ForEach("ul.broadcasts li", fn)

	channel := Channel{
		Name:     channelName,
		Schedule: scheduleList,
	}
	return channel
}
