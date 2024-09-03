package db

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDB() (*gorm.DB, error) {

	// dsn1 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", "admin", "umang9936", "database-1.c9oe6qq44hu1.ap-south-1.rds.amazonaws.com", "3306", "mydatabase")

	dsn := "user:Umang@9936@tcp(127.0.0.1:3306)/mydatabase?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect o database: %v", err)
		return nil, err
	}
	fmt.Println("successful;ly connected")
	return Db, nil

}
