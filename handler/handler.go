package handler

import (
	"log"
	"net/http"
	"time"

	"app.com/db"
	"app.com/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func Getuser(c *gin.Context) {

	var users []models.User
	result := db.Db.Find(&users)
	if result.Error != nil {
		log.Fatalf("failed to retrieve users: %v", result.Error)
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully created user",
		"user":    users,
	})

}

func Createuserfunc(c *gin.Context) {
	var user models.User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse the data"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "unable  to hashed the password"})
	}

	user.Password = string(hash)

	err = user.Save()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully created user",
		"user":    user,
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	// Bind request body to the struct
	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read the body"})
		return
	}

	// Query the user from the database
	var user models.User
	db.Db.Where("Email = ?", body.Email).First(&user)

	// Check if user exists
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid email and password"})
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "password is wrong"})
		return
	}

	// Create a new JWT token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // Token expiration
	})
	tokenString, err := token.SignedString([]byte("umang")) // Signing the token
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to create token"})
		return
	}

	// Send the token in the response header
	c.Header("Authorization", "Bearer "+tokenString)
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   tokenString, // You can still include it in the response body if needed
	})
}

func Validate(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user type assertion failed"})
		return
	}

	// Now you can use userModel in your handler
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome " + userModel.Email,
		"user":    userModel,
	})

}













func Creategroup(c *gin.Context) {
    // Retrieve the user from the context
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found"})
        return
    }

    // Assert the user to the User model
    userModel, ok := user.(models.User)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "user type assertion failed"})
        return
    }

    // Bind the incoming JSON to the Group model
    var group models.Group
    if err := c.ShouldBindJSON(&group); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse the data"})
        return
    }

    // Start a transaction
    tx := db.Db.Begin()
    if tx.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "could not start transaction"})
        return
    }

    // Set the AdminID of the group to the user's ID
    group.AdminID = userModel.ID

    // Create the group in the database
    if err := tx.Create(&group).Error; err != nil {
        tx.Rollback() // Rollback the transaction in case of an error
        c.JSON(http.StatusBadRequest, gin.H{"message": "could not save data to the database"})
        return
    }

    // Insert the member into the MemberGroup table
    memberGroup := models.MemberGroup{
        MemberId: userModel.ID,
        GroupId:  group.ID,
    }
    if err := tx.Create(&memberGroup).Error; err != nil {
        tx.Rollback() // Rollback the transaction in case of an error
        c.JSON(http.StatusBadRequest, gin.H{"message": "could not save data to the database"})
        return
    }

    // Commit the transaction
    if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "could not commit transaction"})
        return
    }

    // Respond with a success message
    c.JSON(http.StatusCreated, gin.H{
        "message":  "Successfully created group",
        "group_ID": group.ID,
    })
}



















func GetGroupName(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user type assertion failed"})
		return
	}

	var group_user []models.MemberGroup
	result := db.Db.Where("member_id = ?", userModel.ID).Find(&group_user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error querying database", "error": result.Error.Error()})
		return
	}

	var groups []models.Response_Group
	for _, groupUser := range group_user {
		var group models.Group
		var group2 models.Response_Group
		result := db.Db.Where("id = ?", groupUser.GroupId).Find(&group)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error querying group", "error": result.Error.Error()})
			return
		}
		group2.AdminID = group.AdminID
		group2.ID = group.ID
		group2.Name = group.Name
		groups = append(groups, group2)
	}

	c.JSON(http.StatusOK, gin.H{"group_users": groups})
}
