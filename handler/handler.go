package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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
		var x models.Response_user;
		x.Email=user.Email;
		x.ID=user.ID;
		x.Name=user.Name;

		users = append(users, x);
	}


	c.JSON(http.StatusOK, gin.H{"group_users": users})
}














func AddMember(c *gin.Context) {
	GroupId := c.Query("id")

	groupID, err := strconv.Atoi(GroupId)
	if err != nil {
		// Handle the error if the conversion fails
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid group ID: %s", GroupId),
		})
		return
	}
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
var  Already_member models.MemberGroup
db.Db.Where("group_id = ? AND member_id = ?", groupID, user.ID).First(&Already_member);

if Already_member.ID!=0{
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
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Successfully created group",
		"member_id": user.ID,
		"group_id" :groupID,
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

	for _, member_group := range member_groups {
		
			var debit models.DebtTrack
		debit.ExpenseId = expense_db.ID
		debit.GivenById = expense_db.GivenById
		debit.Amount = (expense_db.Amount) / uint(len(member_groups))
		debit.OwnById = member_group.MemberId
        if (uint(debit.GivenById)!=uint(debit.OwnById)){
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
    GroupId:= c.Query("group_id")

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
	response.TotalAmount=0;
    
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

        var x int64 = -1
		responseArray.Expense_id=expenseTable.ID
		responseArray.Debit_id=debitTable.ID

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

    // Send the response
    c.JSON(http.StatusOK, gin.H{
        "message": "Successfully fetched settlement",
        "data":    response,
    })
}
