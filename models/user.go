package models

import (
	"time"

	"app.com/db"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:100;unique;not null"`
	Password  string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Group struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	AdminID   uint      `gorm:"not null"`                         // Foreign key referencing User.ID
	Admin     User      `gorm:"foreignKey:AdminID;references:ID"` // Association
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type MemberGroup struct {
	ID       uint  `gorm:"primaryKey;autoIncrement"`
	MemberId uint  `gorm:"not null"`
	Member   User  `gorm:"foreignKey:MemberId;references:ID"`
	GroupId  uint  `gorm:"not null"`
	Group    Group `gorm:"foreignKey:GroupId;references:ID"`
}
type Response_Group struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Name    string `gorm:"size:100;not null"`
	AdminID uint   `gorm:"not null"`
}

type Response_user struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:100;unique;not null"`
}

type Expense_db struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	GroupId     uint           `gorm:"not null"`
	Group       Group          `gorm:"foreignKey:GroupId;references:ID"`
	GivenById   uint           `gorm:"not null"`
	GivenBy     User           `gorm:"foreignKey:GivenById;references:ID"`
	Amount      uint           `gorm:"not null"`
	Category    string         `gorm:"size:200;not null"`
	Description string         `gorm:"size:200;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Soft delete field
}


type Expense_Request struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	GroupId     uint           `gorm:"not null"`
	Group       Group          `gorm:"foreignKey:GroupId;references:ID"`
	GivenByEmail string        `gorm:"not null"` // Changed from uint to string, as email should be a string
	GivenBy     User           `gorm:"foreignKey:GivenByEmail;references:Email"` // Corrected reference to Email
	Amount      uint           `gorm:"not null"`
	Category    string         `gorm:"size:200;not null"`
	Description string         `gorm:"size:200;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Soft delete field
}

type DebtTrack struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	ExpenseId   uint           `gorm:"not null"`
	Expense     Expense_db     `gorm:"foreignKey:ExpenseId;references:ID"`
	GivenById   uint           `gorm:"not null"`
	GivenBy     User           `gorm:"foreignKey:GivenById;references:ID"`
	OwnById     uint           `gorm:"not null"`
	OwnBy       User           `gorm:"foreignKey:OwnById;references:ID"` // Corrected reference to OwnById
	Amount      uint           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Soft delete field
}



type ResponseExchange struct {
 ExchangeAmount  int64           `gorm:"not null"`
 Category        string         `gorm:"size:200;not null"`
 Description string         `gorm:"size:200;not null"`
 Expense_id uint           `gorm:"not null"`
 Debit_id    uint          `gorm:"not null"`

}
type OverallResponseExchange struct {
    TotalAmount     int64              `gorm:"not null"`
    Exchanges       []ResponseExchange `gorm:"-"`
}



















func (e User) Save() error {

	result := db.Db.Create(&e)
	if result.Error != nil {
		// log.Fatalf("failed to create user: %v", result.Error)
		return result.Error
	}
	return nil

}

// func GetAllEvent() []User {
// 	return user
// }
