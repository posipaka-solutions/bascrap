package scraper

import (
	"errors"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/zelenin/go-tdlib/client"
	"strings"
)

const (
	binanceChannelId = -1001146915409
	//binanceChannelId    = -1001390705243 // testChannel
	//binanceChannelId = -1001501948186 //Oleg
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
		ChatId:        binanceChannelId,
		FromMessageId: 0,
		Offset:        0,
		Limit:         20,
		OnlyLocal:     false,
	})
	if err != nil {
		return announcement.Details{}, err
	}

	announcedDetails, err := parseNewMessage(messages.Messages[0])
	if err != nil {
		return announcement.Details{}, err
	}

	if handler.latestAnnounce.Equal(announcedDetails) {
		return announcedDetails, &NoNewsUpdate{}
	}

	handler.latestAnnounce = announcedDetails
	return announcedDetails, nil
}

func parseNewMessage(message *client.Message) (announcement.Details, error) {
	var messageText string
	switch message.Content.MessageContentType() {
	case client.TypeMessageVideo:
		content, isOkay := message.Content.(*client.MessageVideo)
		if !isOkay {
			return announcement.Details{}, errors.New("[scraper] -> Casting of video message failed")
		}
		messageText = content.Caption.Text
		break
	case client.TypeMessagePhoto:
		content, isOkay := message.Content.(*client.MessagePhoto)
		if !isOkay {
			return announcement.Details{}, errors.New("[scraper] -> Casting of photo message failed")
		}
		messageText = content.Caption.Text
		break
	case client.TypeMessageText:
		content, isOkay := message.Content.(*client.MessageText)
		if !isOkay {
			return announcement.Details{}, errors.New("[scraper] -> Casting of text message failed")
		}
		messageText = content.Text.Text
		break
	default:
		return announcement.Details{}, errors.New("[scraper] -> latest message in channel is neither video " +
			"nor photo nor video")
	}

	announcedDetails := announcement.Details{Header: messageText}
	linkIdx := strings.Index(messageText, "https://")
	if linkIdx != -1 {
		announcedDetails.Header = messageText[:linkIdx-1]
		announcedDetails.Link = messageText[linkIdx:]
	}

	return announcedDetails, nil
}
