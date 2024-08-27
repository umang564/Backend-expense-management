package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"app.com/db"
	"app.com/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
)

func sendEmail(to string, subject string, body string) error {
	from := mail.NewEmail("Umang split app", "umangkumar9936@gmail.com")
	toEmail := mail.NewEmail("Recipient", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)
	sendgridAPIKey := os.Getenv("SENDGRID_API_KEY")
	client := sendgrid.NewSendClient(sendgridAPIKey)

	response, err := client.Send(message)
	if err != nil {
		return err
	}

	log.Printf("Response Status Code: %d", response.StatusCode)
	log.Printf("Response Body: %s", response.Body)
	log.Printf("Response Headers: %v", response.Headers)

	return nil
}

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

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "hello umang"})
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

	err = sendEmail(user.Email, "Registration on umang  split app", "you are successfully registerd in umang split app")
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
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

	err = sendEmail(body.Email, "login in umang split app", "you are successfully login in umang split app")
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   tokenString,
		"id":      user.ID, "name": user.Name, "email": user.Email, // You can still include it in the response body if needed
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

func ViewMember(c *gin.Context) {

	GroupId := c.Query("id")

	groupID, err := strconv.Atoi(GroupId)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	var groupUsers []models.MemberGroup
	result := db.Db.Where("group_id = ?", groupID).Find(&groupUsers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error querying database",
			"error":   result.Error.Error(),
		})
		return
	}

	var users []models.Response_user
	for _, member := range groupUsers {
		var user models.User
		result := db.Db.Where("id = ?", member.MemberId).Find(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error querying group member",
				"error":   result.Error.Error(),
			})
			return
		}
		var x models.Response_user
		x.Email = user.Email
		x.ID = user.ID
		x.Name = user.Name

		users = append(users, x)

	}

	c.JSON(http.StatusOK, gin.H{"group_users": users})
}

func AddMember(c *gin.Context) {
	x, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found"})
		return
	}

	userModel, ok := x.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "user type assertion failed"})
		return
	}
	GroupId := c.Query("id")

	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		// Handle the error if the conversion fails
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}
	var group models.Group
	db.Db.Where("id=?", groupID).Find(&group)
	var body struct {
		Email string
	}
	err = c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse the  add member data"})
		return
	}

	var user models.User
	db.Db.Where("Email = ?", body.Email).First(&user)

	// Check if user exists
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "email id is not exist"})
		return
	}
	var Already_member models.MemberGroup
	db.Db.Where("group_id = ? AND member_id = ?", groupID, user.ID).First(&Already_member)

	if Already_member.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user is already added in the group"})
		return
	}

	var member models.MemberGroup
	member.MemberId = user.ID
	member.GroupId = uint(groupID)
	result := db.Db.Create(&member)
	if result.Error != nil {

		c.JSON(http.StatusBadRequest, gin.H{"message": "addition of member is failed"})
		return

	}

	err = sendEmail(body.Email, "you are added in the group", fmt.Sprintf("you are added in the group %s \n created by  =%s \n whose email id ", group.Name, userModel.Name, userModel.Email))
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Successfully created group",
		"member_id": user.ID,
		"group_id":  groupID,
	})

}

func AddMoney(c *gin.Context) {

	GroupId := c.Query("id")

	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	var expense models.Expense_Request
	err = c.ShouldBindJSON(&expense)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse the data"})
		return
	}
	fmt.Println(expense.MemberIDs)
	fmt.Println(expense.Amount)
	fmt.Println(expense.Category)
	var user models.User
	db.Db.Where("email = ?", expense.GivenByEmail).Find(&user)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "The user who is giving amount is not registered"})
		return
	}

	var expense_db models.Expense_db
	expense_db.Amount = expense.Amount
	expense_db.GivenById = user.ID
	expense_db.GroupId = uint(groupID)
	expense_db.Description = expense.Description
	expense_db.Category = expense.Category

	tx := db.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not start transaction"})
		return
	}

	if err := tx.Create(&expense_db).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not save data to the database"})
		return
	}

	var member_groups []models.MemberGroup
	if err := tx.Where("group_id = ?", groupID).Find(&member_groups).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not retrieve members"})
		return
	}

	if len(member_groups) == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"message": "No members found in the group"})
		return
	}

	for _, member_group := range expense.MemberIDs {

		var debit models.DebtTrack
		debit.ExpenseId = expense_db.ID
		debit.GivenById = expense_db.GivenById
		debit.Amount = (expense_db.Amount) / uint(len(expense.MemberIDs))
		debit.OwnById = member_group
		if uint(debit.GivenById) != uint(debit.OwnById) {
			if err := tx.Create(&debit).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"message": "could not save data to the database"})
				return
			}
		}

	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction completed successfully"})
}

func Exchange(c *gin.Context) {
	// Parse the query parameters
	memberIDStr := c.Query("member_id")
	GroupId := c.Query("group_id")

	// Convert member_id and group_id to integers
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", memberIDStr),
		})
		return
	}

	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	// Get the user from the context
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

	var expenseTables []models.Expense_db

	// Fetch expenses where the user is either the giver or receiver within the group
	result := db.Db.Where("(given_by_id = ? OR given_by_id = ?) AND group_id = ?", memberID, userModel.ID, groupID).Find(&expenseTables)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "problem in fetching expense data"})
		return
	}

	var response models.OverallResponseExchange
	response.TotalAmount = 0

	// Iterate through each expense and calculate the corresponding exchange data
	for _, expenseTable := range expenseTables {
		var responseArray models.ResponseExchange
		var debitTable models.DebtTrack

		// Fetch the debit entry for the expense between the two users
		result = db.Db.Where("((given_by_id = ? AND own_by_id = ?) OR (given_by_id = ? AND own_by_id = ?)) AND expense_id = ?",
			memberID, userModel.ID,
			userModel.ID, memberID,
			expenseTable.ID).Find(&debitTable)

		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "problem in fetching debit data"})
			return
		}

		if debitTable.ID != 0 {
			var x int64 = -1
			responseArray.Expense_id = expenseTable.ID
			responseArray.Debit_id = debitTable.ID

			// Populate the response array with the category, description, and calculated amount
			responseArray.Category = expenseTable.Category
			responseArray.Description = expenseTable.Description

			if debitTable.GivenById == userModel.ID {
				responseArray.ExchangeAmount = int64(debitTable.Amount)
				response.TotalAmount += responseArray.ExchangeAmount
			} else {
				responseArray.ExchangeAmount = int64(debitTable.Amount) * x
				response.TotalAmount += responseArray.ExchangeAmount
			}

			// Append the response array to the overall response
			response.Exchanges = append(response.Exchanges, responseArray)
		}
	}

	// Send the response
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched settlement",
		"data":    response,
	})
}

func Debit(c *gin.Context) {

	var res models.DebitRequest

	// Bind the JSON payload to the res struct
	if err := c.ShouldBindJSON(&res); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse the data sent while clicking on settled delete button"})
		return
	}
	var user models.User
	db.Db.Where("id=? ", res.MemberID).Find(&user)
	var debitx models.DebtTrack
	db.Db.Where("id=?", res.DebitId).Find(&debitx)

	var debit models.DebtTrack
	var amount int = int(debitx.Amount)

	var expense models.Expense_db
	db.Db.Where("id=?", debitx.ExpenseId).Find(&expense)
	// Perform the soft delete based on the DebitId
	if err := db.Db.Where("id = ?", res.DebitId).Delete(&debit).Error; err != nil {
		// Handle error
		log.Println("Error while performing soft delete:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while deleting record"})
		return
	}
	err := sendEmail(user.Email, "expense cleared on particular  settlement", fmt.Sprintf("expense cleared on the  category %s \n of amount =%d", expense.Category, amount))
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}

	log.Println("Record soft deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Record soft deleted successfully"})
}

func Notify(c *gin.Context) {
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

	var notify models.Notify
	if err := c.ShouldBindJSON(&notify); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse the data sent while clicking on settled delete button"})
		return
	}

	var member models.User
	db.Db.Where("id=?", notify.Member_id).Find(&member)
	err := sendEmail(
		member.Email,
		"Notification for the balance",
		fmt.Sprintf("You need to pay balance = %d\nWith username = %s\nWith email = %s",
			notify.Total_amount, userModel.Name, userModel.Email),
	)

	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notified successfully"})

}

func DeleteGroup(c *gin.Context) {
	GroupId := c.Query("groupid")
	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	// Start a transaction
	tx := db.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to start transaction"})
		return
	}

	// Perform the delete operations
	if err := tx.Where("group_id = ?", groupID).Delete(&models.MemberGroup{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete MemberGroup records"})
		return
	}

	if err := tx.Where("group_id = ?", groupID).Delete(&models.Expense_db{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete Expense_db records"})
		return
	}

	if err := tx.Where("id = ?", groupID).Delete(&models.Group{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete Group record"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func GroupDetails(c *gin.Context) {

	GroupId := c.Query("groupid")
	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}
	var expense_dbs []models.Expense_db
	if err := db.Db.Where("group_id=?", groupID).Find(&expense_dbs).Error; err != nil {
		// Handle error
		log.Println("problem in fetching expense array", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while fetching expense"})
		return
	}

	var ress []models.Group_Details
	for _, expense_db := range expense_dbs {
		var res models.Group_Details

		var user models.User
		db.Db.Where("id=?", expense_db.GivenById).Find(&user)

		res.Amount = expense_db.Amount
		res.Category = expense_db.Category
		res.GivenByName = user.Name
		res.Description = expense_db.Description

		ress = append(ress, res)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched settlement",
		"data":    ress,
	})

}

func AllSettle(c *gin.Context) {
	GroupId := c.Query("groupid")
	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	MemberId := c.Query("memberid")
	memberID, err := strconv.Atoi(MemberId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid member ID: %s", MemberId),
		})
		return
	}

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

	var expense_dbs []models.Expense_db
	db.Db.Where("(given_by_id = ? OR given_by_id = ?) AND group_id = ?", memberID, userModel.ID, groupID).Find(&expense_dbs)

	// Start the transaction
	tx := db.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to start transaction"})
		return
	}

	for _, expense_db := range expense_dbs {
		var x models.DebtTrack
		err := tx.Where("((given_by_id = ? AND own_by_id = ?) OR (given_by_id = ? AND own_by_id = ?)) AND expense_id = ?",
			memberID, userModel.ID,
			userModel.ID, memberID,
			expense_db.ID).Delete(&x).Error

		if err != nil {
			tx.Rollback() // Rollback the transaction in case of error
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete debt track"})
			return
		}
	}

	// Commit the transaction if all delete operations are successful
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to commit transaction"})
		return
	}

	var member models.User
	db.Db.Where("id = ?", memberID).Find(&member)
	err = sendEmail(
		member.Email,
		"Notification for the balance",
		"The amount between us has been settled down",
	)

	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully deleted",
	})
}

const (
	bucketName = "bucket-umang"
	region     = "ap-south-1" // e.g., "us-west-2"
)

func CsvFile(c *gin.Context) {
	GroupId := c.Query("groupid")
	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}

	var expense_dbs []models.Expense_db
	if err := db.Db.Where("group_id=?", groupID).Find(&expense_dbs).Error; err != nil {
		log.Println("problem in fetching expense array", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while fetching expense"})
		return
	}

	var ress []models.Group_Details
	for _, expense_db := range expense_dbs {
		var res models.Group_Details

		var user models.User
		db.Db.Where("id=?", expense_db.GivenById).Find(&user)

		res.Amount = expense_db.Amount
		res.Category = expense_db.Category
		res.GivenByName = user.Name
		res.Description = expense_db.Description

		ress = append(ress, res)
	}

	// Generate CSV content
	var csvBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuffer)

	// Write CSV header
	header := []string{"Amount", "Category", "GivenByName", "Description"}
	if err := csvWriter.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error writing CSV header"})
		return
	}

	// Write CSV rows
	for _, item := range ress {
		record := []string{
			fmt.Sprintf("%f", item.Amount),
			item.Category,
			item.GivenByName,
			item.Description,
		}
		if err := csvWriter.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error writing CSV record"})
			return
		}
	}
	csvWriter.Flush()

	// Create a session with AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Println("Failed to create AWS session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create AWS session"})
		return
	}

	s3Client := s3.New(sess)

	// Upload CSV file to S3
	objectKey := "expenses.csv"
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(csvBuffer.Bytes()),
		ContentType: aws.String("text/csv"),
	})
	if err != nil {
		log.Println("Failed to upload CSV to S3", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload CSV to S3"})
		return
	}

	// Generate S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, objectKey)

	// Return the URL
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully uploaded to S3",
		"url":     s3URL,
	})
}
