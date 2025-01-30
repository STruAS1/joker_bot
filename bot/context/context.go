package context

import (
	"SHUTKANULbot/config"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserState struct {
	mu        sync.Mutex
	Name      string
	Level     int
	Data      map[string]interface{}
	MessageID int
	LastSeen  time.Time
}

type Context struct {
	BotAPI     *tgbotapi.BotAPI
	UserStates sync.Map
	Config     *config.Config
}

type BotContext struct {
	Ctx           *Context
	UserID        int64
	Message       *tgbotapi.Message
	CallbackQuery *tgbotapi.CallbackQuery
}

func NewContext(botAPI *tgbotapi.BotAPI, cfg *config.Config) *Context {
	ctx := &Context{
		BotAPI: botAPI,
		Config: cfg,
	}
	go ctx.CleanupOldUsers()
	return ctx
}

func GetUserState(botCtx *BotContext) *UserState {
	if state, exists := botCtx.Ctx.UserStates.Load(botCtx.UserID); exists {
		userState := state.(*UserState)
		userState.mu.Lock()
		userState.LastSeen = time.Now()
		userState.mu.Unlock()
		return userState
	}

	newState := &UserState{
		Name:      "start",
		Level:     0,
		Data:      make(map[string]interface{}),
		MessageID: 0,
		LastSeen:  time.Now(),
	}

	botCtx.Ctx.UserStates.Store(botCtx.UserID, newState)
	return newState
}

func UpdateUserLevel(botCtx *BotContext, newLevel int) {
	state := GetUserState(botCtx)
	state.mu.Lock()
	state.Level = newLevel
	state.mu.Unlock()
}

func UpdateUserName(botCtx *BotContext, newName string) {
	if len(newName) > 50 {
		newName = newName[:50]
	}
	state := GetUserState(botCtx)
	state.mu.Lock()
	state.Name = newName
	state.Level = 0
	state.mu.Unlock()
}

func ClearAllUserData(botCtx *BotContext) {
	state := GetUserState(botCtx)
	state.mu.Lock()
	state.Data = make(map[string]interface{})
	state.mu.Unlock()
}

func SaveMessageID(botCtx *BotContext, messageID int) {
	state := GetUserState(botCtx)
	state.mu.Lock()
	state.MessageID = messageID
	state.mu.Unlock()
}

func (botCtx *BotContext) SendMessage(msg tgbotapi.MessageConfig) (int, error) {
	sentMessage, err := botCtx.Ctx.BotAPI.Send(msg)
	if err != nil {
		return 0, err
	}

	SaveMessageID(botCtx, sentMessage.MessageID)
	return sentMessage.MessageID, nil
}

func (ctx *Context) CleanupOldUsers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		var toDelete []int64

		ctx.UserStates.Range(func(key, value interface{}) bool {
			userID := key.(int64)
			state := value.(*UserState)

			state.mu.Lock()
			inactive := now.Sub(state.LastSeen) > 30*time.Minute
			state.mu.Unlock()

			if inactive {
				toDelete = append(toDelete, userID)
			}
			return true
		})

		log.Printf("Удаляем %d неактивных пользователей", len(toDelete))

		for _, userID := range toDelete {
			ctx.UserStates.Delete(userID)
		}
	}
}
