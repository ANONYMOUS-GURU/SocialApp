package controllers

import (
	"encoding/json"
	"errors"
	"g_chat/database"
	ws "g_chat/wsConnections"
	"log"
	"time"
)

// TODO : can seen_count, delivered_count be > sent_to_count.

func chatWrittenHandler(channel string, payload string) error {
	if channel != "chat_written" {
		log.Printf("wrong channel")
		return errors.New("wrong handler channel")
	}

	var msgUser database.Messageusermap

	if err := json.Unmarshal([]byte(payload), &msgUser); err != nil {
		log.Printf("error unmarshalling payload, err : %v", err)
		return err
	}

	message, err := getMessageFromId(msgUser.MessageID, msgUser.ReceiverID)

	if err != nil {
		log.Print("error getting message from DB : err %V", err)
		return err
	}

	ws.GetConnectionManager().PerformSendMessageToUserWS(message)

	return nil
}

func chatReceivedHandler(channel string, payload string) error {
	if channel != "chat_received" {
		log.Printf("wrong channel")
		return errors.New("wrong handler channel")
	}

	var message database.Message

	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		log.Printf("error unmarshalling payload, err : %v", err)
		return err
	}

	deliveryUpdate := ws.OutgoingDeliveredUpdate{
		MessageId:  message.ID,
		ReceiverID: message.SenderID,
		Time:       time.Now().UnixNano(),
	}

	ws.GetConnectionManager().PerformOutgoingDeliveredUpdateWS(deliveryUpdate)

	return nil
}

func chatReadHandler(channel string, payload string) error {
	if channel != "chat_received" {
		log.Printf("wrong channel")
		return errors.New("wrong handler channel")
	}

	var message database.Message

	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		log.Printf("error unmarshalling payload, err : %v", err)
		return err
	}

	readUpdate := ws.OutgoingReadUpdate{
		MessageId:  message.ID,
		ReceiverID: message.SenderID,
		Time:       time.Now().UnixNano(),
	}

	ws.GetConnectionManager().PerformOutgoingReadUpdateWS(readUpdate)

	return nil
}

func RegisterDBNotifyHandlers() {
	var dbNotifyHandler = make(map[string]func(channel string, payload string) error)

	dbNotifyHandler["chat_written"] = chatWrittenHandler
	dbNotifyHandler["chat_received"] = chatReceivedHandler
	dbNotifyHandler["chat_read"] = chatReadHandler

	ws.GetConnectionManager().SetupOutgoingEventHandlers(dbNotifyHandler)
}
