package bot

import (
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers"
	"SHUTKANULbot/config"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(cfg *config.Config) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	ctx := context.NewContext(botAPI, cfg)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)
	updateQueue := make(chan tgbotapi.Update, 100)

	workerCount := 50

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for update := range updateQueue {
				processUpdate(ctx, update)
			}
		}()
	}

	for update := range updates {
		select {
		case updateQueue <- update:
		default:
			log.Println("Очередь обновлений переполнена, пропуск сообщения")
		}
	}

	close(updateQueue)
	wg.Wait()
}

func processUpdate(ctx *context.Context, update tgbotapi.Update) {
	var userID int64
	var message *tgbotapi.Message
	var callbackQuery *tgbotapi.CallbackQuery
	var lastAction string

	if update.Message != nil {
		if update.Message.Chat.Type != "private" {
			return
		}
		userID = update.Message.Chat.ID
		message = update.Message
		lastAction = "message"
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Chat.Type != "private" {
			return
		}
		userID = update.CallbackQuery.Message.Chat.ID
		callbackQuery = update.CallbackQuery
		lastAction = "callback"
	} else {
		return
	}

	botCtx := &context.BotContext{
		Ctx:           ctx,
		UserID:        userID,
		Message:       message,
		CallbackQuery: callbackQuery,
	}

	state := context.GetUserState(botCtx)
	state.Data["LastAction"] = lastAction

	handlers.HandleUpdate(botCtx)
}
