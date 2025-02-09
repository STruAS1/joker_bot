package Utilities

import (
	"SHUTKANULbot/bot/context"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var photoCache = make(map[string]string)

func GetPhotoId(botCtx *context.BotContext, filename string) string {
	fileDir := fmt.Sprintf("./photos/%s", filename)
	if fileID, exists := photoCache[filename]; exists {
		return fileID
	}

	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		return ""
	}

	photo := tgbotapi.NewPhoto(botCtx.Ctx.Config.Bot.AdminId, tgbotapi.FilePath(fileDir))
	msg, err := botCtx.Ctx.BotAPI.Send(photo)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fileId := msg.Photo[len(msg.Photo)-1].FileID

	photoCache[filename] = fileId
	return fileId
}
