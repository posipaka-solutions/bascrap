package telegram

import (
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/zelenin/go-tdlib/client"
	"path/filepath"
)

const (
	apiId             = 8061033
	apiHash           = "5665589a975a637135402402542dd520"
	posipakaChannelId = -1001577983722
)

type Telegram struct {
	Client *client.Client
}

func SendMessageToChannel(message string, tdlibClient *client.Client) {

	test := client.InputMessageText{
		Text: &client.FormattedText{
			Text:     message,
			Entities: nil,
		},
		DisableWebPagePreview: false,
		ClearDraft:            false,
	}
	messageReq := client.SendMessageRequest{
		ChatId:              posipakaChannelId,
		MessageThreadId:     0,
		ReplyToMessageId:    0,
		Options:             nil,
		ReplyMarkup:         nil,
		InputMessageContent: &test,
	}
	_, err := tdlibClient.SendMessage(&messageReq)
	if err != nil {
		cmn.LogInfo.Print("Failed in sending message to channel")
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
		cmn.LogError.Print("TDLib client creation failed. Error: ", err.Error())
		return nil
	}

	_, err = tdlibClient.GetChats(&client.GetChatsRequest{
		ChatList:     nil,
		OffsetOrder:  9223372036854775807,
		OffsetChatId: 0,
		Limit:        10,
	})
	if err != nil {
		cmn.LogError.Print("Chat list request failed. Error: ", err.Error())
		return nil
	}

	return tdlibClient
}
