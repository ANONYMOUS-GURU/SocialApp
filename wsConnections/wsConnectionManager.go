package websockets

import (
	"context"
	"encoding/json"
	"g_chat/models"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// checkOrigin will check origin and return true if its allowed
func checkOrigin(r *http.Request) bool {

	// Grab the request origin
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return true
	}
}

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024, // might use a pool
	}
)

const (
	PONG_INTERVAL = 30
	PING_INTERVAL = 35
)

type Event struct {
	// Type is the message type sent
	Type EventType `json:"type"`
	// a uuid sent from client to manage ack client side
	Id string `json:"id"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
	// Retry Count
	Retry uint `json:"retry"`
}

type ConnectionManager struct {
	ConnectionMap    map[string][]*Client
	IncomingHandlers map[EventType]EventHandler
	OutgoingHandlers map[string]func(channel string, payload string) error
	TicketManager    *Tickets
	sync.RWMutex
}

func (manager *ConnectionManager) SetupIncomingEventHandlers(handlers map[EventType]EventHandler) {
	manager.IncomingHandlers = handlers
}

func (manager *ConnectionManager) addClient(userId string, client *Client) {
	manager.Lock()
	defer manager.Unlock()
	manager.ConnectionMap[userId] = append(manager.ConnectionMap[userId], client)
}

func (manager *ConnectionManager) RemoveClient(client *Client) {
	manager.Lock()
	defer manager.Unlock()
	indexToDelete := -1
	clients := manager.ConnectionMap[client.UserId]
	for i, c := range clients {
		if c == client {
			indexToDelete = i
			break
		}
	}

	if indexToDelete >= 0 {
		client.Conn.Close()
		clients = append(clients[:indexToDelete], clients[indexToDelete+1:]...)
		if len(clients) == 0 {
			delete(manager.ConnectionMap, client.UserId)
		} else {
			manager.ConnectionMap[client.UserId] = clients
		}
	} else {
		log.Println("client does not exist")
	}
}

var (
	myWSConnectionManager *ConnectionManager
)

func GetConnectionManager() *ConnectionManager {
	return myWSConnectionManager
}

// TODO - one client per user only (implement requests for disconnecting)
func CreateConnectionManager(ctx context.Context) {
	myWSConnectionManager = &ConnectionManager{
		ConnectionMap:    make(map[string][]*Client),
		IncomingHandlers: make(map[EventType]EventHandler),
		OutgoingHandlers: make(map[string]func(channel string, payload string) error),
		TicketManager:    CreateNewTicketsMap(ctx, time.Second*30),
	}
}

func (manager *ConnectionManager) CreateNewTicket(ctx *gin.Context) {
	manager.TicketManager.generateTicket(ctx)
}

func (manager *ConnectionManager) ServeWS(ctx *gin.Context) {
	userId, valid := manager.TicketManager.validateTicket(ctx)
	if !valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized or request timeout",
		})
		return
	}

	Conn, err := websocketUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		Conn:              Conn,
		UserId:            userId,
		Egress:            make(chan Event),
		ConnectionManager: manager,
	}

	manager.addClient(userId, client)

	go client.ReadMessage()
	go client.WriteMessage()
}

func (manager *ConnectionManager) PerformOutgoingReadUpdateWS(outgoingReadUpdate OutgoingReadUpdate) {
	userId := outgoingReadUpdate.ReceiverID

	if _, ok := manager.ConnectionMap[userId]; !ok {
		log.Printf("user not connected %v", userId)
		return
	}

	for _, client := range manager.ConnectionMap[userId] {
		go client.SendReadUpdateToClient(outgoingReadUpdate)
	}
}

func (manager *ConnectionManager) PerformOutgoingDeliveredUpdateWS(outgoingdeliveryUpdate OutgoingDeliveredUpdate) {
	userId := outgoingdeliveryUpdate.ReceiverID

	if _, ok := manager.ConnectionMap[userId]; !ok {
		log.Printf("user not connected %v", userId)
		return
	}

	for _, client := range manager.ConnectionMap[userId] {
		go client.SendDeliveryUpdateToClient(outgoingdeliveryUpdate)
	}
}

func (manager *ConnectionManager) PerformSendMessageToUserWS(messagePayload models.OutgoingChatPayload) {
	if _, ok := manager.ConnectionMap[messagePayload.ReceiverId]; !ok {
		log.Printf("user not connected %v", messagePayload.ReceiverId)
		return
	}

	for _, client := range manager.ConnectionMap[messagePayload.ReceiverId] {
		go client.SendMessageToClient(messagePayload)
	}
}

func (manager *ConnectionManager) SetupOutgoingEventHandlers(handlers map[string]func(channel string, payload string) error) {
	manager.OutgoingHandlers = handlers
}
