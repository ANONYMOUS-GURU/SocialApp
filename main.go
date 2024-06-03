package main

import (
	"context"
	"fmt"
	"g_chat/controllers"
	"g_chat/database"
	"g_chat/firebase"
	"g_chat/routes"
	ws "g_chat/wsConnections"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	if err := database.CreateDatabaseInstance(); err != nil {
		log.Fatalf("error creating database instance %v", err)
	}

	err = firebase.CreateFirebaseApp()
	if err != nil {
		log.Fatalf("Failed Firebase connection %v", err)
	}

	ws.CreateConnectionManager(context.Background())

	if err := database.InitializeListener(ws.GetConnectionManager()); err != nil {
		log.Fatalf("error creating database notification on ws : error - %v", err)
	}

	fmt.Println("Initialized Firebase and Database and DB notifiers")
}

func main() {
	server := gin.Default()

	server.Use(CORSMiddleware())

	userServer := server.Group("/api/v1/users")
	wsGroup := server.Group("/api/v1/ws")
	chatGroup := server.Group("/api/v1/chat")

	routes.CreateUserRoutes(userServer)
	routes.CreateWSRoutes(wsGroup)
	routes.CreateChatRoutes(chatGroup)

	controllers.RegisterWSHandlers()

	server.Run()
}

// TODO Cors

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
