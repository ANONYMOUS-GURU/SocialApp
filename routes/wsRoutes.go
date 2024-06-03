package routes

import (
	"g_chat/middleware"
	ws "g_chat/wsConnections"

	"github.com/gin-gonic/gin"
)

func CreateWSRoutes(baseRouter *gin.RouterGroup) {
	// use auth middleware
	ticketRoute := baseRouter.Group("/ticketing")
	ticketRoute.Use(middleware.ValidateUserToken())
	ticketRoute.GET("/createTicket", ws.GetConnectionManager().CreateNewTicket)

	// Dont use middleware
	baseRouter.GET("/connect", ws.GetConnectionManager().ServeWS)
}
