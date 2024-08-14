package router

import (
	"app.com/handler"
	"app.com/middleware"
	"github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
	userGroup := router.Group("/user")
	userGroup.Use(middleware.RequiredAuth)
	{
		
		userGroup.GET("/getuser", handler.Getuser)
		userGroup.POST("/creategroup",handler.Creategroup)
		userGroup.GET("/getgroupname",handler.GetGroupName)
		
		userGroup.POST("/addmember", handler.AddMember)
		userGroup.GET("/viewmember",handler.ViewMember)
		userGroup.POST("addmoney",handler.AddMoney)
		userGroup.GET("/exchange",handler.Exchange)




		userGroup.GET("/validate", middleware.RequiredAuth, handler.Validate)
	}



	authGroup:=router.Group("/auth")
{
	authGroup.POST("/createuser", handler.Createuserfunc)
	authGroup.POST("/login", handler.Login)

}
	

}
