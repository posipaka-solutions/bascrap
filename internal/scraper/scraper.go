package scraper

import (
	"errors"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/zelenin/go-tdlib/client"
	"strings"
)

const (
	binanceChannelId = -1001146915409
	//testChannelId    = -1001146915409
)

type ScrapHandler struct {
	latestAnnounce announcement.Details
	tdlibClient    *client.Client
}

func New(tclient *client.Client) ScrapHandler {
	handler := ScrapHandler{
		tdlibClient: tclient,
	}

	var err error
	handler.latestAnnounce, err = handler.GetLatestAnnounce()
	if err != nil {
		panic(err.Error())
	}

	return handler
}

func (handler *ScrapHandler) GetLatestAnnounce() (announcement.Details, error) {
	messagesReq := client.GetChatHistoryRequest{
		ChatId:        binanceChannelId,
		FromMessageId: 0,
		Offset:        0,
		Limit:         20,
		OnlyLocal:     false,
	}
	messages, err := handler.tdlibClient.GetChatHistory(&messagesReq)
	if err != nil {
		return announcement.Details{}, err
	}

	content, isOkay := messages.Messages[0].Content.(*client.MessageVideo)
	if !isOkay {
		return announcement.Details{}, errors.New("[scraper] -> Message content casting failed")
	}

	announcedDetails := announcement.Details{
		Header: content.Caption.Text,
	}
	linkIdx := strings.Index(content.Caption.Text, "https://")
	if linkIdx != -1 {
		announcedDetails.Header = content.Caption.Text[:linkIdx-1]
		announcedDetails.Link = content.Caption.Text[linkIdx:]
	}

	if handler.latestAnnounce.Equal(announcedDetails) {
		return announcement.Details{}, &NoNewsUpdate{}
	}

	handler.latestAnnounce = announcedDetails
	return announcedDetails, nil
}
