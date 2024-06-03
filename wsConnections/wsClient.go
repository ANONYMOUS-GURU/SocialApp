package websockets

import (
	"encoding/json"
	"errors"
	"fmt"
	"g_chat/models"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn              *websocket.Conn
	UserId            string
	Egress            chan Event
	ConnectionManager *ConnectionManager
	MessagesChan      chan []models.OutgoingChatPayload
}

func (client *Client) routeIncomingEvent(event Event) error {
	fmt.Println(event)
	if _, ok := client.ConnectionManager.IncomingHandlers[event.Type]; !ok {
		log.Print("Invalid Event type")
		return errors.New("invalid event type")
	}
	return client.ConnectionManager.IncomingHandlers[event.Type](event, client)
}

func (client *Client) ReadMessage() {
	defer func() {
		// Graceful Close the Connection once this
		// function is done
		client.ConnectionManager.RemoveClient(client)
	}()

	// Set Max Size of Messages in Bytes
	client.Conn.SetReadLimit(512)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := client.Conn.SetReadDeadline(time.Now().Add(PONG_INTERVAL * time.Second)); err != nil {
		log.Printf("Error setting Read Deadline %v", err)
		return
	}

	// Handling Pong Messages
	client.Conn.SetPongHandler(client.PongHandler)

	// Loop Forever
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		messageType, payload, err := client.Conn.ReadMessage()

		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			log.Printf("error reading message: %v", err)
			break // Break the loop to close Conn & Cleanup
		}

		// chat Data recieved
		if messageType == websocket.TextMessage {
			var event Event
			if err := json.Unmarshal(payload, &event); err != nil {
				log.Printf("Could not read the payload %v", err)
				client.SendAckToClient(Acknowledge{
					ReceiverID: client.UserId,
					EventType:  Unknown,
					Status:     false,
					Message:    "failed unmarshal to event, check input json",
					AckTime:    time.Now().Unix(),
				}, "")
			} else {
				go client.routeIncomingEvent(event)
			}
		}
	}
}

func (client *Client) WriteMessage() {
	defer func() {
		// Graceful close if this triggers a closing
		client.ConnectionManager.RemoveClient(client)
	}()

	tick := time.NewTicker(PING_INTERVAL)

	for {
		select {
		case message, ok := <-client.Egress:
			// Ok will be false Incase the Egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := client.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				log.Printf("Error marshalling the message %v", err)
			}

			// Write data to the connection
			if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}

		case <-tick.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Error sending ping to client %v", err)
				return
			}
		}

	}
}

func (client *Client) PongHandler(pong string) error {
	if err := client.Conn.SetReadDeadline(time.Now().Add(PONG_INTERVAL * time.Second)); err != nil {
		log.Printf("Error setting Read Deadline %v", err)
		return err
	}
	return nil
}

func (client *Client) SendMessageToClient(messagePayload models.OutgoingChatPayload, retryCount ...uint) {
	payload, err := json.Marshal(messagePayload)

	if err != nil {
		log.Printf("Error marshalling outgoing message payload %v", err)
		return
	}

	var retry uint = 0
	if len(retryCount) > 0 {
		retry = retryCount[0]
	}

	event := Event{
		Type:    EventOutgoingChatMessage,
		Id:      "",
		Payload: payload,
		Retry:   retry,
	}

	client.Egress <- event
}

func (client *Client) SendReadUpdateToClient(readUpdatePayload OutgoingReadUpdate, retryCount ...uint) {
	payload, err := json.Marshal(readUpdatePayload)

	if err != nil {
		log.Printf("Error marshalling outgoing read update payload %v", err)
		return
	}

	var retry uint = 0
	if len(retryCount) > 0 {
		retry = retryCount[0]
	}

	event := Event{
		Type:    EventOutgoingReadUpdate,
		Payload: payload,
		Id:      "",
		Retry:   retry,
	}

	client.Egress <- event
}

func (client *Client) SendDeliveryUpdateToClient(outgoingDeliveryUpdate OutgoingDeliveredUpdate, retryCount ...uint) {
	payload, err := json.Marshal(outgoingDeliveryUpdate)

	if err != nil {
		log.Printf("Error marshalling outgoing delivery update payload %v", err)
		return
	}

	var retry uint = 0
	if len(retryCount) > 0 {
		retry = retryCount[0]
	}

	event := Event{
		Type:    EventOutgoingDeliveredUpdate,
		Payload: payload,
		Id:      "",
		Retry:   retry,
	}

	client.Egress <- event
}

func (client *Client) SendAckToClient(ack Acknowledge, id string) {
	payload, err := json.Marshal(ack)

	if err != nil {
		log.Printf("Error marshalling outgoing message payload %v", err)
		return
	}

	event := Event{
		Type:    AckEvent,
		Payload: payload,
		Id:      id,
		Retry:   0,
	}

	client.Egress <- event
}
