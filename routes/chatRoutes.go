package routes

import (
	"g_chat/controllers"
	"g_chat/middleware"

	"github.com/gin-gonic/gin"
)

// TODO : implement all routes for chat

func CreateChatRoutes(baseRouter *gin.RouterGroup) {
	baseRouter.Use(middleware.ValidateUserToken())

	baseRouter.POST("/createConversationWithUser", controllers.CreateUser)
	baseRouter.POST("/createGroupConversation", controllers.CreateUser)

	baseRouter.POST("/sendMessageInGroup", controllers.UpdateUser)
	baseRouter.POST("/sendMessageToUser", controllers.DeleteUser)

	baseRouter.GET("/getAllMessages", controllers.CreateUser)
}
