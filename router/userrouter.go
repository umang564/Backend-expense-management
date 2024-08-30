package router

import (
	"app.com/handler"
	"app.com/middleware"
	"github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
	userGroup := router.Group("/user")
	router.GET("/health", handler.HealthCheck)
	userGroup.Use(middleware.RequiredAuth)
	{

		userGroup.GET("/getuser", handler.Getuser)
		userGroup.POST("/creategroup", handler.Creategroup)
		userGroup.GET("/getgroupname", handler.GetGroupName)

		userGroup.POST("/addmember", handler.AddMember)
		userGroup.GET("/viewmember", handler.ViewMember)
		userGroup.POST("addmoney", handler.AddMoney)
		userGroup.GET("/exchange", handler.Exchange)
		userGroup.DELETE("/settleddebit", handler.Debit)
		userGroup.POST("/notify", handler.Notify)
		userGroup.DELETE("/deleteGroup", handler.DeleteGroup)
		userGroup.GET("/groupDetails", handler.GroupDetails)
		userGroup.DELETE("/allsettle", handler.AllSettle)
		userGroup.GET("/csvfile", handler.CsvFile)
		userGroup.GET("/totalamount",handler.GetTotalAmount)

		userGroup.GET("/validate", middleware.RequiredAuth, handler.Validate)
	}

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/createuser", handler.Createuserfunc)
		authGroup.POST("/login", handler.Login)

	}

}
