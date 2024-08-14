package main

import (
	"log"

	"app.com/db"
	"app.com/models"
	"app.com/router"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	server := gin.Default()
	_, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
    db.Db.AutoMigrate(&models.DebtTrack{});
	// Automigrate the schema

	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)

	}
	router.Routes(server)

	// server.GET("/users", func(c *gin.Context) {
	// 	var users []models.User
	// 	result :=Db.Find(&users)
	// 	if result.Error != nil {
	// 		log.Fatalf("failed to retrieve users: %v", result.Error)
	// 	}
	// 	c.JSON(http.StatusCreated, gin.H{
	// 		"message": "Successfully created user",
	// 		"user":    users,
	// 	})

	// })

	// server.POST("/users", func(c *gin.Context) {
	// 	var user models.User

	// 	err := c.ShouldBindJSON(&user)
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse the data"})
	// 		return
	// 	}
	// 	user.Save()
	// 	c.JSON(http.StatusCreated, gin.H{
	// 		"message": "Successfully created user",
	// 		"user":    user,
	// 	})
	// })

	server.Run(":8080")
}
