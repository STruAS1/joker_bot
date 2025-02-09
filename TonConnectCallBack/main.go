package TonConnectCallback

import (
	contextBot "SHUTKANULbot/bot/context"
	"SHUTKANULbot/config"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cameo-engineering/tonconnect"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/exp/maps"

	"github.com/go-redis/redis/v8"
)

type UserSession struct {
	Session    *tonconnect.Session `json:"session"`
	WalletAddr string              `json:"walletAddr"`
}

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func getRedisClient() *redis.Client {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	})
	return redisClient
}

func getSession(userID int64) (string, *tonconnect.Session, bool) {
	key := fmt.Sprintf("session_TON:%d", userID)
	data, err := getRedisClient().Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil, false
		}
		return "", nil, false
	}

	var us UserSession
	if err := json.Unmarshal([]byte(data), &us); err != nil {
		return "", nil, false
	}

	return us.WalletAddr, us.Session, true
}

func saveSession(userID int64, session *tonconnect.Session, walletAddr string) {
	key := fmt.Sprintf("session_TON:%d", userID)
	us := UserSession{
		Session:    session,
		WalletAddr: walletAddr,
	}
	data, err := json.Marshal(us)
	if err != nil {
		return
	}
	if err := getRedisClient().Set(context.Background(), key, data, 0).Err(); err != nil {
		return
	}
}

func Disconnect(userID int64) {
	_, Session, exist := getSession(userID)
	if exist {
		key := fmt.Sprintf("session_TON:%d", userID)
		getRedisClient().Del(context.Background(), key)
		go Session.Disconnect(context.Background())

	}
}
func GenerateWalletLinks(BotCtx *contextBot.BotContext, ctx context.Context) (map[string]string, string, error) {
	log.Println("Создаю новую сессию для пользователя:", BotCtx.UserID)
	s, err := tonconnect.NewSession()
	if err != nil {
		log.Println("Ошибка создания сессии:", err)
		return nil, "", err
	}

	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		log.Println("Ошибка генерации случайных данных:", err)
		return nil, "", err
	}
	cfg := config.LoadConfig()
	connreq, err := tonconnect.NewConnectRequest(
		fmt.Sprintf("https://%s/tonconnect-manifest.json", cfg.Domines),
		tonconnect.WithProofRequest(base32.StdEncoding.EncodeToString(data)),
	)
	if err != nil {
		log.Println("Ошибка создания запроса на подключение:", err)
		return nil, "", err
	}

	deeplink, err := s.GenerateDeeplink(*connreq, tonconnect.WithBackReturnStrategy())
	if err != nil {
		log.Println("Ошибка генерации deeplink:", err)
		return nil, "", err
	}

	walletLinks := map[string]string{}
	for _, wallet := range tonconnect.Wallets {
		link, err := s.GenerateUniversalLink(wallet, *connreq)
		if err != nil {
			log.Println("Ошибка генерации ссылки для", wallet.Name, err)
			return nil, "", err
		}
		walletLinks[wallet.Name] = link
	}

	go WaitForConnection(s, BotCtx, ctx)
	return walletLinks, deeplink, nil
}

func WaitForConnection(s *tonconnect.Session, BotCtx *contextBot.BotContext, ctx context.Context) error {
	state := contextBot.GetUserState(BotCtx)
	cancel, exist := state.Data["ChancelConnect"].(context.CancelFunc)

	if exist {
		defer func() {
			cancel()
			delete(state.Data, "ChancelConnect")
		}()
	}
	res, err := s.Connect(ctx, (maps.Values(tonconnect.Wallets))...)
	if err != nil {
		_, ex := state.Data["SendErr"].(bool)
		if ex {
			delete(state.Data, "SendErr")
			return err
		}
		deleteCfg := tgbotapi.DeleteMessageConfig{
			ChatID:    BotCtx.UserID,
			MessageID: state.MessageID,
		}
		var rows [][]tgbotapi.InlineKeyboardButton
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Назад", "back"),
		))
		BotCtx.Ctx.BotAPI.Send(deleteCfg)
		Text := "❌ Ошибка подключения!"
		msg := tgbotapi.NewMessage(BotCtx.UserID, Text)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

		if _, err := BotCtx.SendMessage(msg); err != nil {
			fmt.Print(err)
		}

		return err
	}

	var addr string
	for _, item := range res.Items {
		if item.Name == "ton_addr" {
			addr = item.Address
		}
	}
	saveSession(BotCtx.UserID, s, addr)
	deleteCfg := tgbotapi.DeleteMessageConfig{
		ChatID:    BotCtx.UserID,
		MessageID: state.MessageID,
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Назад", "back"),
	))
	BotCtx.Ctx.BotAPI.Send(deleteCfg)
	Text := "✅ Кошелёк успешно подключен!"
	msg := tgbotapi.NewMessage(BotCtx.UserID, Text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := BotCtx.SendMessage(msg); err != nil {
		fmt.Print(err)
	}
	return nil
}

func IsUserConnected(userID int64) (bool, string) {
	wallet, _, exist := getSession(userID)
	if exist {
		WalStruc, _ := address.ParseRawAddr(wallet)
		wallet = WalStruc.Bounce(false).String()
	}
	return exist, wallet
}

func SendTokensViaTonConnect(botCtx *contextBot.BotContext, amount uint64, ctx context.Context) error {
	userID := botCtx.UserID
	state := contextBot.GetUserState(botCtx)
	adr, s, exists := getSession(userID)
	if !exists || s == nil {
		return errors.New("сессия не найдена, пользователь не подключен")
	}
	var user models.User
	result := db.DB.Where(&models.User{TelegramID: userID}).First(&user)
	if result.Error != nil {
		return result.Error
	}
	contractAddress := "UQDTkytrQsT-S08PhPn-WjgzcQz-2BBwJ8FLE58FM99E21VZ"
	newUUID := uuid.New()
	uuidString := newUUID.String()

	log.Printf("Сгенерированный UUID v4: %s\n", uuidString)
	var rows [][]tgbotapi.InlineKeyboardButton
	payloadCell := cell.BeginCell()
	payloadCell.MustStoreUInt(0, 32)
	payloadCell.MustStoreStringSnake(uuidString)
	payloadBytes := payloadCell.EndCell().ToBOC()

	msg, err := tonconnect.NewMessage(contractAddress, "200000000", tonconnect.WithPayload(payloadBytes))
	if err != nil {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
		))
		msgT := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, "❌ Ошибка отправки транзакции.", tgbotapi.NewInlineKeyboardMarkup(rows...))
		msgT.DisableWebPagePreview = true
		msgT.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msgT)
		log.Print(err)
		return err
	}

	tx, err := tonconnect.NewTransaction(
		tonconnect.WithTimeout(5*time.Minute),
		tonconnect.WithMessage(*msg),
		tonconnect.WithMainnet(),
		tonconnect.WithFrom(adr),
	)
	if err != nil {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
		))
		msgT := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, "❌ Ошибка отправки транзакции.", tgbotapi.NewInlineKeyboardMarkup(rows...))
		msgT.DisableWebPagePreview = true
		msgT.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msgT)
		log.Print(err)
		return err
	}

	cancel, exist := state.Data["ChancelTransaction"].(context.CancelFunc)

	if exist {
		defer func() {
			cancel()
			delete(state.Data, "ChancelTransaction")
		}()
	}

	boc, err := s.SendTransaction(ctx, *tx)
	if err != nil {
		_, ex := state.Data["SendErr"].(bool)
		if ex {
			delete(state.Data, "SendErr")
			return err
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
		))
		msgT := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, "❌ Ошибка отправки транзакции.", tgbotapi.NewInlineKeyboardMarkup(rows...))
		msgT.DisableWebPagePreview = true
		msgT.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msgT)
		log.Print(err)
		return err
	}
	bocBase64 := base64.StdEncoding.EncodeToString(boc)
	log.Printf("boc: %x\n", bocBase64)
	newTransaction := models.TransactionNet{
		UserID: user.ID,
		UUID:   uuidString,
		Amount: amount,
		Status: 0,
		Wallet: adr,
	}

	user.WithdrawBalance(db.DB, amount)
	fmt.Println(adr)
	db.DB.Create(&newTransaction)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("К настойкам", "back"),
	))
	msgT := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, "✅ Ожидайте поступления на кошелёк", tgbotapi.NewInlineKeyboardMarkup(rows...))
	msgT.DisableWebPagePreview = true
	msgT.ParseMode = "HTML"
	botCtx.Ctx.BotAPI.Send(msgT)
	return nil
}
