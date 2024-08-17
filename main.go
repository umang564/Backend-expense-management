package main

import (
	"log"

	"app.com/db"
	"app.com/models"
	"app.com/router"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	
)
func sendEmail(to string, subject string, body string) error {
	from := mail.NewEmail("Example User", "umangkumar9936@gmail.com")
	toEmail := mail.NewEmail("Recipient", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)
	client := sendgrid.NewSendClient("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")

	response, err := client.Send(message)
	if err != nil {
		return err
	}

	log.Printf("Response Status Code: %d", response.StatusCode)
	log.Printf("Response Body: %s", response.Body)
	log.Printf("Response Headers: %v", response.Headers)

	return nil
}



func main() {
	
	

	db.InitDB()
	server := gin.Default()
	_, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	db.Db.AutoMigrate(&models.User{});
	db.Db.AutoMigrate(&models.Group{});
	db.Db.AutoMigrate(&models.MemberGroup{});
	db.Db.AutoMigrate(&models.Expense_db{});
	db.Db.AutoMigrate(&models.DebtTrack{})

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
