package models

import (
	"time"

	"app.com/db"
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
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	AdminID   uint      `gorm:"not null"`  
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
