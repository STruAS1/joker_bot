package models

type TransactionNet struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"index"`
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Amount uint64
	UUID   string
	Status uint
	Wallet string
}
