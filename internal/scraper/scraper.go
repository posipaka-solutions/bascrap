package scraper

import (
	"errors"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/zelenin/go-tdlib/client"
	"strings"
)

const (
	//binanceChannelId = -1001146915409
	//binanceChannelId    = -1001390705243 // testChannel
	//binanceChannelId = -1001501948186 //Oleg
	madNewsChannelId = -1001501948186 //-1001219306781
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
		if _, isOkay := err.(*NoNewsUpdate); !isOkay {
			panic("First call of latest news getter exited with error: " + err.Error())
		}
	}
	return handler
}

func (handler *ScrapHandler) GetLatestAnnounce() (announcement.Details, error) {
	messages, err := handler.tdlibClient.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:        madNewsChannelId,
		FromMessageId: 0,
		Offset:        0,
		Limit:         20,
		OnlyLocal:     false,
	})
	if err != nil {
		return announcement.Details{}, err
	}

	announcedDetails, err := parseMadNewsMessage(messages.Messages[0])
	if err != nil {
		return announcement.Details{}, err
	}

	if handler.latestAnnounce.Equal(announcedDetails) {
		return announcedDetails, &NoNewsUpdate{}
	}

	handler.latestAnnounce = announcedDetails
	return announcedDetails, nil
}

func parseMadNewsMessage(message *client.Message) (announcement.Details, error) {
	var messageText string
	content, isOkay := message.Content.(*client.MessageText)
	if !isOkay {
		return announcement.Details{}, errors.New("[scraper] -> Casting of text message failed")
	}
	messageText = content.Text.Text
	messageText = messageText[:strings.Index(messageText, "\n")-1]

	announcedDetails := announcement.Details{Header: messageText}

	return announcedDetails, nil

}
