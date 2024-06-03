package routes

import (
	"g_chat/controllers"
	"g_chat/middleware"

	"github.com/gin-gonic/gin"
)

// TODO : implement all routes for user

func CreateUserRoutes(baseRouter *gin.RouterGroup) {
	baseRouter.Use(middleware.ValidateUserToken())

	baseRouter.POST("/create", controllers.CreateUser)
	baseRouter.PATCH("/update", controllers.UpdateUser)
	baseRouter.DELETE("/delete/:id", controllers.DeleteUser)
	baseRouter.GET("/getCurrentUser", controllers.GetCurrentUserDetails)
	baseRouter.GET("/get", controllers.GetUserFromId)
}
