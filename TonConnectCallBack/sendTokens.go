package TonConnectCallback

import (
	"SHUTKANULbot/db/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"gorm.io/gorm"
)

type TonApiEventsResponse struct {
	Events   []Event `json:"events"`
	NextFrom int     `json:"next_from"`
}

type Event struct {
	EventID    string   `json:"event_id"`
	Account    Account  `json:"account"`
	Timestamp  int64    `json:"timestamp"`
	Actions    []Action `json:"actions"`
	IsScam     bool     `json:"is_scam"`
	Lt         int64    `json:"lt"`
	InProgress bool     `json:"in_progress"`
	Extra      int64    `json:"extra"`
}

type Account struct {
	Address  string `json:"address"`
	IsScam   bool   `json:"is_scam"`
	IsWallet bool   `json:"is_wallet"`
}

type Action struct {
	Type             string       `json:"type"`
	TonTransfer      *TonTransfer `json:"TonTransfer,omitempty"`
	BaseTransactions []string     `json:"base_transactions,omitempty"`
}

type TonTransfer struct {
	Sender    Account `json:"sender"`
	Recipient Account `json:"recipient"`
	Amount    uint64  `json:"amount"`
	Comment   string  `json:"comment"`
}

func isTransactionConfirmed(ctx context.Context, ownerWallet, userWallet, uuid string) (bool, error) {
	offset := 0
	limit := 20
	for range 30 {
		var url string
		if offset != 0 {
			url = fmt.Sprintf("https://tonapi.io/v2/accounts/%s/events?limit=%d&before_lt=%d", userWallet, limit, offset)
		} else {
			url = fmt.Sprintf("https://tonapi.io/v2/accounts/%s/events?limit=%d", userWallet, limit)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return false, fmt.Errorf("запрос к тонапи поломался к хуям: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, fmt.Errorf("ответ тонапи не пришёл: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("тонапи отдало странный статус: %d", resp.StatusCode)
		}

		var data TonApiEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return false, fmt.Errorf("нихрена не разобрал JSON: %v", err)
		}

		if len(data.Events) == 0 {
			break
		}

		for _, event := range data.Events {
			if event.Actions != nil && event.Actions[0].TonTransfer != nil && event.Actions[0].TonTransfer.Sender.Address == userWallet && event.Actions[0].TonTransfer.Recipient.Address == ownerWallet && event.Actions[0].TonTransfer.Comment != "" {
				if event.Actions[0].TonTransfer.Comment == uuid {
					if !event.InProgress {
						return true, nil
					} else {
						return false, nil
					}
				}
			}
		}

		offset = data.NextFrom

		time.Sleep(500 * time.Millisecond)
	}

	return false, nil
}

func mintTokens(
	ctx context.Context,
	adminWallet *wallet.Wallet,
	minterAddr string,
	userAddr string,
	tokenAmount uint64,
) error {
	minter, err := address.ParseRawAddr(minterAddr)
	if err != nil {
		return fmt.Errorf("не смог распарсить minterAddr: %v", err)
	}
	user, err := address.ParseRawAddr(userAddr)
	if err != nil {
		return fmt.Errorf("не смог распарсить userAddr: %v", err)
	}

	nullAddr, err := address.ParseRawAddr("0:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return fmt.Errorf("не смог распарсить null адрес: %v", err)
	}
	transferAmount := uint64(20000000)
	forwardTon := uint64(1000000)
	sendAmount := tlb.MustFromTON("0.04")

	internalMsg := cell.BeginCell().
		MustStoreUInt(0x178d4519, 32).
		MustStoreUInt(0, 64).
		MustStoreCoins(tokenAmount).
		MustStoreAddr(nullAddr).
		MustStoreAddr(user).
		MustStoreCoins(forwardTon).
		MustStoreUInt(0, 1).
		EndCell()

	payload := cell.BeginCell().
		MustStoreUInt(21, 32).
		MustStoreUInt(0, 64).
		MustStoreAddr(user).
		MustStoreCoins(transferAmount).
		MustStoreRef(internalMsg).
		EndCell()

	msg := &wallet.Message{
		Mode: 3,
		InternalMessage: &tlb.InternalMessage{
			DstAddr: minter,
			Amount:  sendAmount,
			Body:    payload,
		},
	}

	if err := adminWallet.Send(ctx, msg); err != nil {
		return fmt.Errorf("ошибка отправки mint-транзакции: %v", err)
	}

	log.Println("✅ Mint транзакция отправлена")
	return nil
}

func TransactionWorker(db *gorm.DB, api *ton.APIClient, ctx context.Context, adminWallet *wallet.Wallet, minterAddr, ownerWallet string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("🔍 Проверка неподтвержденных транзакций...")

			var transactions []models.TransactionNet
			if err := db.Where("status = 0").Find(&transactions).Error; err != nil {
				log.Printf("Ошибка запроса транзакций: %v", err)
				continue
			}

			for _, tx := range transactions {
				confirmed, err := isTransactionConfirmed(ctx, ownerWallet, tx.Wallet, tx.UUID)
				if err != nil {
					log.Printf("Ошибка проверки TX %d: %v", tx.ID, err)
					continue
				}

				if !confirmed {
					continue
				}

				if err := mintTokens(ctx, adminWallet, minterAddr, tx.Wallet, tx.Amount); err != nil {
					log.Printf("Ошибка минта для TX %d: %v", tx.ID, err)
					continue
				}

				if err := db.Model(&tx).Update("status", 1).Error; err != nil {
					log.Printf("Ошибка обновления статуса TX %d: %v", tx.ID, err)
				} else {
					log.Printf("✅ TX %d успешно обработан", tx.ID)
				}
			}
		}
	}
}
