package telegram

import (
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"github.com/zelenin/go-tdlib/client"
	"path/filepath"
)

const (
	apiId             = 8061033
	apiHash           = "5665589a975a637135402402542dd520"
	posipakaChannelId = -1001577983722
)

func SendMessageToChannel(message string, tdlibClient *client.Client) {

	text := client.InputMessageText{
		Text: &client.FormattedText{
			Text: message,
		},
	}
	messageReq := client.SendMessageRequest{
		ChatId:              posipakaChannelId,
		MessageThreadId:     0,
		ReplyToMessageId:    0,
		InputMessageContent: &text,
	}
	_, err := tdlibClient.SendMessage(&messageReq)
	if err != nil {
		log.Info.Print("Failed in sending message to channel")
	}
}

func NewTDLibClient() *client.Client {
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &client.TdlibParameters{
		DatabaseDirectory:      filepath.Join(".tdlib", "database"),
		FilesDirectory:         filepath.Join(".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		ApiId:                  apiId,
		ApiHash:                apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "PosipakaServer",
		ApplicationVersion:     "0.9-development",
		EnableStorageOptimizer: true,
	}

	logVerbosity := client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 0,
	})

	tdlibClient, err := client.NewClient(authorizer, logVerbosity)
	if err != nil {
		log.Error.Print("TDLib client creation failed. Error: ", err.Error())
		return nil
	}

	_, err = tdlibClient.GetChats(&client.GetChatsRequest{
		OffsetOrder:  9223372036854775807,
		OffsetChatId: 0,
		Limit:        10,
	})
	if err != nil {
		log.Error.Print("Chat list request failed. Error: ", err.Error())
		return nil
	}

	return tdlibClient
}
