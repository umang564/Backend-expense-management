package models

import (
	"time"

	"app.com/db"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"size:100;unique;not null"`
	Password  string         `gorm:"size:255;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete field
}

type SuggestionList struct{

	Name      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"size:100;unique;not null"`


}
type SettledExpense struct{
OwnByName	string         `gorm:"size:100;not null"`
GivenByName string         `gorm:"size:100;not null"`
Category    string         `gorm:"size:200;not null"`
Amount   uint           `gorm:"not null"`
Time  gorm.DeletedAt `gorm:"index"` 
}

type Group struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	AdminID   uint           `gorm:"not null"`                                                     // Foreign key referencing User.ID
	Admin     User           `gorm:"foreignKey:AdminID;references:ID;constraint:OnDelete:CASCADE"` // Association with cascade delete
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete field
}

type MemberGroup struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	MemberId  uint           `gorm:"not null"`
	Member    User           `gorm:"foreignKey:MemberId;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for Member
	GroupId   uint           `gorm:"not null"`
	Group     Group          `gorm:"foreignKey:GroupId;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for Group
	DeletedAt gorm.DeletedAt `gorm:"index"`                                                        // Soft delete field
}

type Expense_db struct {
	ID          uint           `gorm:"primaryKey;autoIncrement"`
	GroupId     uint           `gorm:"not null"`
	Group       Group          `gorm:"foreignKey:GroupId;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for Group
	GivenById   uint           `gorm:"not null"`
	GivenBy     User           `gorm:"foreignKey:GivenById;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for User (GivenBy)
	Amount      uint           `gorm:"not null"`
	Category    string         `gorm:"size:200;not null"`
	Description string         `gorm:"size:200;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Soft delete field
}

type DebtTrack struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	ExpenseId uint           `gorm:"not null"`
	Expense   Expense_db     `gorm:"foreignKey:ExpenseId;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for Expense
	GivenById uint           `gorm:"not null"`
	GivenBy   User           `gorm:"foreignKey:GivenById;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for User (GivenBy)
	OwnById   uint           `gorm:"not null"`
	OwnBy     User           `gorm:"foreignKey:OwnById;references:ID;constraint:OnDelete:CASCADE"` // Cascade delete for User (OwnBy)
	Amount    uint           `gorm:"not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete field
}

type Group_Details struct {
	GivenByName string `gorm:"size:100;not null"`
	Amount      uint   `gorm:"not null"`
	Category    string `gorm:"size:200;not null"`
	Description string `gorm:"size:200;not null"`
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

type Expense_Request struct {
	ID           uint     `gorm:"primaryKey;autoIncrement"`
	GroupId      uint     `gorm:"not null"`
	Group        Group    `gorm:"foreignKey:GroupId;references:ID"`
	GivenByEmail string   `gorm:"not null"`                                 // Changed from uint to string, as email should be a string
	GivenBy      User     `gorm:"foreignKey:GivenByEmail;references:Email"` // Corrected reference to Email
	Amount       uint     `gorm:"not null"`
	Category     string   `gorm:"size:200;not null"`
	Description  string   `gorm:"size:200;not null"`
	MemberIDs    []uint  `gorm:"type:integer[]"`
}

type ResponseExchange struct {
	ExchangeAmount int64  `gorm:"not null"`
	Category       string `gorm:"size:200;not null"`
	Description    string `gorm:"size:200;not null"`
	Expense_id     uint   `gorm:"not null"`
	Debit_id       uint   `gorm:"not null"`
}
type OverallResponseExchange struct {
	TotalAmount int64              `gorm:"not null"`
	Exchanges   []ResponseExchange `gorm:"-"`
}

type DebitRequest struct {
	DebitId  uint `gorm:"not null"`
	MemberID uint `gorm:"not null"`
}

type Notify struct {
	Member_id    uint `gorm:"not null"`
	Total_amount uint `gorm:"not null"`
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
