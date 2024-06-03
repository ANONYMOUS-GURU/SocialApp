package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"g_chat/database"
	"g_chat/models"
	ws "g_chat/wsConnections"
	"log"
	"time"
)

// TODO validate all input data and auth check for evry handler by using sender_id == client.userId

func handleIncomingChatMessage(event ws.Event, client *ws.Client) error {
	var incomingChatPayload models.IncomingChatPayload
	var ackData ws.Acknowledge

	if err := json.Unmarshal(event.Payload, &incomingChatPayload); err != nil {
		log.Printf("incorrect payload of incoming chat message %v", err)
		ackData = ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the chat message",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	if client.UserId != incomingChatPayload.SenderId {
		log.Print("WTF !!!")
		client.ConnectionManager.RemoveClient(client)
		return errors.New("unauthorized, disconnecting client")
	}

	if incomingChatPayload.IsGroup {
		isUserPartOfConv, err := IsUserPartOfConversation(incomingChatPayload.ConversationId, client.UserId)

		if err != nil {
			ackData = ws.Acknowledge{
				EventType: event.Type,
				Status:    false,
				Message:   "cannot confirm authorization",
				AckTime:   time.Now().UnixNano(),
			}
			client.SendAckToClient(ackData, event.Id)

			return err
		}

		if !isUserPartOfConv {
			ackData = ws.Acknowledge{
				EventType: event.Type,
				Status:    false,
				Message:   "user not part of conversation",
				AckTime:   time.Now().UnixNano(),
			}
			client.SendAckToClient(ackData, event.Id)

			return errors.New("user not part of conversation")
		}
	}

	// Write to DB
	_, err := database.GetChatQueries().WriteIncomingMessageWS(context.Background(), incomingChatPayload) // TODO : this function should be moved to chat controller
	if err != nil {
		ackData = ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed inserting in DB",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)
		return err
	}

	ackData = ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}
	client.SendAckToClient(ackData, event.Id)

	return nil
}

func handleDeliveredUpdateForMessage(event ws.Event, client *ws.Client) error {
	var incomingDeliveredUpdate ws.IncomingDeliveredUpdate
	if err := json.Unmarshal(event.Payload, &incomingDeliveredUpdate); err != nil {
		log.Printf("Error Unmarshalling acknowledge data %v", err)
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "error unmarshalling data",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		return err
	}

	if incomingDeliveredUpdate.SenderId != client.UserId {
		log.Printf("WTF !!")
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "unauthorized, disconnecting",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		client.ConnectionManager.RemoveClient(client)
		return errors.New("unauthorized disconnecting client")
	}

	deletedRows, err := markMessageAsRecievedByUserWS(incomingDeliveredUpdate)
	if err != nil {
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "DB error",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		return err
	}

	if len(deletedRows) == 0 {
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "message already marked as delivered",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		return nil
	}

	client.SendAckToClient(ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "success",
		AckTime:   time.Now().UnixNano(),
	}, event.Id)

	return nil
}

func handleIncomingReadUpdate(event ws.Event, client *ws.Client) error {
	ctx := context.Background()
	var incomingReadUpdate ws.IncomingReadUpdate
	if err := json.Unmarshal(event.Payload, &incomingReadUpdate); err != nil {
		log.Printf("Error Unmarshalling read update data %v", err)
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "error unmarshalling data",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		return err
	}

	if client.UserId != incomingReadUpdate.SenderId {
		log.Printf("unauthorized, closing connection")
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "unauthorized, disconnecting",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		client.ConnectionManager.RemoveClient(client)
		return errors.New("unauthorized, closing connection")
	}

	// check user is part of conversation
	val, err := database.GetChatQueries().IsUserPartOfConversation(ctx, client.UserId, incomingReadUpdate.ConversationId)

	if err != nil {
		return err
	}

	if !val {
		log.Printf("unauthorized, doesn't belong to conversation")
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "unauthorized, disconnecting",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		client.ConnectionManager.RemoveClient(client)
		return errors.New("unauthorized, doesn't belong to conversation")
	}

	if err := database.GetChatQueries().UpdateLastMessageSeenInConversationForUser(ctx, incomingReadUpdate.ConversationId, incomingReadUpdate.SenderId, incomingReadUpdate.Time); err != nil {
		client.SendAckToClient(ws.Acknowledge{
			ReceiverID: client.UserId,
			EventType:  event.Type,
			Status:     false,
			Message:    "DB error",
			AckTime:    time.Now().UnixNano(),
		}, event.Id)
		return err
	}

	client.SendAckToClient(ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "success",
		AckTime:   time.Now().UnixNano(),
	}, event.Id)

	return nil
}

func markMessageAsRecievedByUserWS(incomingDeliveredUpdate ws.IncomingDeliveredUpdate) ([]database.Messageusermap, error) {
	messageIds := []string{incomingDeliveredUpdate.MessageId}
	deletedRows, err := database.GetChatQueries().MarkMessagesAsRecievedByUser(context.Background(), messageIds, incomingDeliveredUpdate.SenderId)
	if err != nil {
		return nil, err
	}
	return deletedRows, nil
}

func getMessageFromId(msgId string, receiverId string) (models.OutgoingChatPayload, error) {
	message, err := database.GetChatQueries().GetMessageByIdWS(context.Background(), msgId)

	if err != nil {
		log.Printf("Error retreving message %v", err)
		return models.OutgoingChatPayload{}, err
	}

	return models.OutgoingChatPayload{
		ID:                message.ID,
		MessageBody:       message.Body,
		Sender:            message.SenderID,
		ConversationId:    message.ConversationID,
		SentAt:            message.SentAt,
		ServerRecieveTime: message.CreatedAt,
		ReceiverId:        receiverId,
	}, nil
}

func RegisterWSHandlers() {
	var handlers = make(map[ws.EventType]ws.EventHandler)

	handlers[ws.EventIncomingDeliveredUpdate] = handleDeliveredUpdateForMessage
	handlers[ws.EventIncomingChatMessage] = handleIncomingChatMessage
	handlers[ws.EventIncomingReadUpdate] = handleIncomingReadUpdate

	// handlers[ws.EventIncomingSocialRequest] = handleIncomingSocialRequest
	// handlers[ws.EventIncomingSocialRequestStatusChange] = handleIncomingSocialRequestStatusChange

	// handlers[ws.EventIncomingCalendarRequest] = handleIncomingCalendarRequest
	// handlers[ws.EventIncomingCalendarRequestStatusChange] = handleIncomingCalendarRequestStatusChange

	ws.GetConnectionManager().SetupIncomingEventHandlers(handlers)
}
