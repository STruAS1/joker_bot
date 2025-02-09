package models

import (
	"SHUTKANULbot/blockchain"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	TotalCoins *Coins
	once       sync.Once
)

type User struct {
	ID             uint   `gorm:"primaryKey"`
	TelegramID     int64  `gorm:"uniqueIndex"`
	Balance        uint64 `gorm:"index;not null;default:0"`
	Username       string `gorm:"size:100"`
	FirstName      string `gorm:"size:100"`
	LastName       string `gorm:"size:100"`
	AuthorUserName string
	AnonymsMode    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Transactions   []Transaction `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "Users"
}

type Transaction struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Amount    uint64 `gorm:"not null"`
	Type      string `gorm:"size:50;not null"`
	CreatedAt time.Time
}

func (Transaction) TableName() string {
	return "Transactions"
}

type Coins struct {
	DBCoins    uint64
	BlockCoins uint64
}

func InitTotalCoins(db *gorm.DB) {
	once.Do(func() {
		var DBCoins float64
		if err := db.Model(&User{}).Select("SUM(balance)").Scan(&DBCoins).Error; err != nil {
			DBCoins = 0
		}
		TotalCoins = &Coins{
			DBCoins:    uint64(DBCoins),
			BlockCoins: blockchain.GetTotalSupply(),
		}
	})
}
func (u *User) SetAnonymsMode(db *gorm.DB) error {
	if u.AnonymsMode {
		u.AnonymsMode = false
	} else {
		u.AnonymsMode = true
	}
	return db.Save(u).Error
}
func (u *User) DepositBalance(db *gorm.DB, amount uint64) error {
	u.Balance += amount
	if err := db.Save(u).Error; err != nil {
		return err
	}
	TotalCoins.DBCoins += amount
	TotalCoins.BlockCoins -= amount
	transaction := Transaction{
		UserID: u.ID,
		Amount: uint64(amount),
		Type:   "deposit",
	}
	return db.Create(&transaction).Error
}
func (u *User) WithdrawBalance(db *gorm.DB, amount uint64) error {
	InitTotalCoins(db)

	if u.Balance < amount {
		return gorm.ErrInvalidData
	}

	TotalCoins.DBCoins -= amount
	TotalCoins.BlockCoins += amount
	u.Balance -= amount
	if err := db.Save(u).Error; err != nil {
		return err
	}

	transaction := Transaction{
		UserID: u.ID,
		Amount: uint64(amount),
		Type:   "withdraw",
	}
	return db.Create(&transaction).Error
}

func (u *User) AddTokenForEvaluationAuthor(db *gorm.DB, evaluation uint64) error {
	InitTotalCoins(db)

	GlobalTotalCoins := TotalCoins.DBCoins + TotalCoins.BlockCoins

	var power uint64 = 0
	if GlobalTotalCoins >= 1_000_000_000_000 {
		power = (GlobalTotalCoins / 1_000_000_000_000) - 1
	}

	difficulty := uint64(1)
	if power > 0 {
		difficulty = 1 << power
	}

	gift := ((evaluation * 1_000_000_000) / 5) / difficulty

	fmt.Print(gift)
	if u.Balance+gift < u.Balance {
		return fmt.Errorf("переполнение баланса")
	}

	u.Balance += gift
	if err := db.Save(u).Error; err != nil {
		return err
	}

	transaction := Transaction{
		UserID: u.ID,
		Amount: gift,
		Type:   "deposit",
	}
	return db.Create(&transaction).Error
}

func (u *User) AddTokenForEvaluation(db *gorm.DB) error {
	InitTotalCoins(db)

	GlobalTotalCoins := TotalCoins.DBCoins + TotalCoins.BlockCoins

	var power uint64 = 0
	if GlobalTotalCoins >= 1_000_000_000_000 {
		power = (GlobalTotalCoins / 1_000_000_000_000) - 1
	}

	difficulty := uint64(1)
	if power > 0 {
		difficulty = 1 << power
	}

	gift := ((1 * 1_000_000_00) / 5) / difficulty

	fmt.Print(gift)
	if u.Balance+gift < u.Balance {
		return fmt.Errorf("переполнение баланса")
	}

	u.Balance += gift
	if err := db.Save(u).Error; err != nil {
		return err
	}

	transaction := Transaction{
		UserID: u.ID,
		Amount: gift,
		Type:   "deposit",
	}
	return db.Create(&transaction).Error
}
