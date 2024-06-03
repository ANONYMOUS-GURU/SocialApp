package database

import (
	"context"
	"database/sql"
	"errors"
	"g_chat/models"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type ChatQueries struct {
	*Queries
}

// TODO use context with timeout and deadlines for creating new goroutines for DB and other async ops.
// TODO use context with deadline for input request context.
// TODO batch updates/writes/deletes using pgx - LATER

func GetChatQueries() *ChatQueries {
	queries := getQueries()
	return &ChatQueries{queries}
}

// Returns if THE user is part of the conversation where THE message was sent
func (db *ChatQueries) FMessageIsInUsersConversation(ctx context.Context, userId string, messageId string) (bool, error) {
	val, err := db.Queries.fMessageIsInUsersConversation(ctx, fMessageIsInUsersConversationParams{
		ID:     messageId,
		UserID: userId,
	})

	if err != nil {
		log.Printf("error in querying DB")
		return false, err
	}

	return val, nil
}

// Get a single message given Id for THE message
func (db *ChatQueries) GetMessageByID(ctx context.Context, messageId string) (Message, error) {
	message, err := db.Queries.getMessageByID(ctx, messageId)
	if err != nil {
		log.Printf("error querying DB %v", err)
		return Message{}, err
	}
	return message, nil
}

// Get a single message given Id for THE message
func (db *ChatQueries) GetMessageByIdWS(ctx context.Context, messageId string) (Message, error) {
	message, err := db.Queries.getMessageByID(ctx, messageId)
	if err != nil {
		log.Printf("error querying DB %v", err)
		return Message{}, err
	}
	return message, nil
}

func (db *ChatQueries) GetAllMessagesAfterGivenTime(ctx context.Context, userId string, timestamp int64, numRows uint) ([]Message, error) {
	rows, err := db.Queries.getAllMessagesAfterGivenTime(ctx, getAllMessagesAfterGivenTimeParams{
		UserID:    userId,
		CreatedAt: timestamp,
		Limit:     int32(numRows),
	})

	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(rows))
	for i, row := range rows {
		messages[i].ID = row.ID
		messages[i].Body = row.Body
		messages[i].ConversationID = row.ConversationID
		messages[i].SenderID = row.SenderID
		messages[i].SentAt = row.SentAt
		messages[i].CreatedAt = row.CreatedAt
	}

	return messages, nil
}

// TODO: NO need maybe
func (db *ChatQueries) GetUnsentMessages(ctx context.Context, timestmap int64, numrows uint, convId string) ([]models.OutgoingChatPayload, bool, error) {
	return nil, false, nil
}

// Gets all Messages after a given TIME for a given conversation. only numRows are returned
func (db *ChatQueries) GetMostRecentMessagesForUserInConversation(ctx context.Context, time int64, numrows uint, convId string, userId string) ([]Message, error) {
	val, err := db.Queries.isUserPartOfConversation(ctx, isUserPartOfConversationParams{
		UserID:         userId,
		ConversationID: convId,
	})

	if err != nil {
		log.Print("error getting info from DB")
		return nil, err
	}

	if !val {
		return nil, errors.New("unauthorized")
	}

	messages, err := db.Queries.getMostRecentMessagesForUserInConversation(ctx, getMostRecentMessagesForUserInConversationParams{
		ConversationID: convId,
		CreatedAt:      time,
		Limit:          int32(numrows),
	})

	if err != nil {
		log.Print("error getting info from DB")
		return nil, err
	}

	return messages, nil
}

// Writes incoming message to Message table and Unsent message table
func (db *ChatQueries) WriteIncomingMessageWS(ctx context.Context, incomingChatPayload models.IncomingChatPayload) (Message, error) {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	insertedMessage, err := writeIncomingMessageHelper(ctx, qtx, &incomingChatPayload)

	if err != nil {
		log.Printf("DB error : error writing message : f(WriteIncomingMessageWS) : error : %v", err)
		return Message{}, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : commiting transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return Message{}, err
	}

	return insertedMessage, nil
}

func writeIncomingMessageHelper(ctx context.Context, qtx *Queries, incomingChatPayload *models.IncomingChatPayload) (Message, error) {
	var convId string

	if incomingChatPayload.IsGroup {
		convId = incomingChatPayload.ConversationId
	} else {
		if incomingChatPayload.ReceiverId < incomingChatPayload.SenderId {
			convId = incomingChatPayload.ReceiverId + incomingChatPayload.SenderId
		} else {
			convId = incomingChatPayload.SenderId + incomingChatPayload.ReceiverId
		}
	}

	conversation, err := qtx.getConversationByID(ctx, convId)
	if err != nil || reflect.DeepEqual(conversation, Conversation{}) {
		if err != nil {
			log.Printf("DB error : error getting conversation info : f(WriteIncomingMessageWS) : error : %v", err)
			return Message{}, err
		} else if conversation.IsGroup {
			log.Printf("Value not found in DB : conversation does not exist : f(WriteIncomingMessageWS)")
			return Message{}, errors.New("value not found in DB : conversation does not exist : f(writeincomingmessagews)")
		} else {
			log.Printf("conversation not created for 1o1 message, creating it ... ")
			if _, err := createConversationWithUserHelperWithTransaction(ctx, incomingChatPayload.SenderId, incomingChatPayload.ReceiverId, qtx); err != nil {
				return Message{}, err
			}
		}
	}

	msgId := uuid.NewString()

	receivers, err := qtx.getAllUsersInConversation(ctx, incomingChatPayload.ConversationId)

	if err != nil {
		return Message{}, err
	}

	insertedMessage, err := qtx.createMessage(ctx, createMessageParams{
		ID:             msgId,
		Body:           incomingChatPayload.MessageBody,
		ConversationID: convId,
		SentAt:         incomingChatPayload.SentAt,
		SenderID:       incomingChatPayload.SenderId,
		CreatedAt:      time.Now().UnixNano(),
		DeliveredCount: 1,
		SeenCount:      1,
		SentToCount:    int32(len(receivers)),
	})

	if err != nil {
		return Message{}, err
	}

	for _, receiver := range receivers {
		err := qtx.createMessageUserMap(ctx, createMessageUserMapParams{
			MessageID:  msgId,
			ReceiverID: receiver.UserID,
		})

		if err != nil {
			return Message{}, err
		}
	}

	return insertedMessage, nil
}

// Writes incoming message to Message table and Unsent message table
func (db *ChatQueries) WriteIncomingMessages(ctx context.Context, incomingChatPayloads []models.IncomingChatPayload) error {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	for _, incomingChatPayload := range incomingChatPayloads {
		if _, err := writeIncomingMessageHelper(ctx, qtx, &incomingChatPayload); err != nil {
			log.Printf("DB error : writing message to DB failed : f(WriteIncomingMessages) : error : %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : commiting transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return err
	}

	return nil
}

// For a list of messages it removes the message from the UnsentMessages table
func (db *ChatQueries) MarkMessagesAsRecievedByUser(ctx context.Context, messageIds []string, userId string) ([]Messageusermap, error) {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		log.Printf("DB error : creating transaction failed : f(MarkMessagesAsRecievedByUser) : error : %v", err)
		return nil, err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	// TODO : this can cause error if multiple clients open at the same time. As messages will be deleted for another client which current client will also try to delete.
	deletedRows, err := qtx.markMessageAsReceivedByUser(ctx, markMessageAsReceivedByUserParams{
		Column1:    messageIds,
		ReceiverID: userId,
	})

	if err != nil {
		log.Printf("DB error : error deleting message : f(MarkMessagesAsRecievedByUser) : error : %v", err)
		return nil, err
	}

	// TODO: do this using trigger ie for every deleted message from messageusermap increment that messages count by 1
	err = qtx.updateDeliveredCountForMessages(ctx, messageIds)

	if err != nil {
		log.Printf("DB error : error updating delivery count in message : f(MarkMessagesAsRecievedByUser) : error : %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : commiting transaction failed : f(MarkMessagesAsRecievedByUser) : error : %v", err)
		return nil, err
	}

	return deletedRows, nil
}

// For a given message it updates the conversation participant table to mark last_message_seen_at column with the timestamp of the message
func (db *ChatQueries) UpdateLastMessageSeenInConversationForUser(ctx context.Context, conversationId string, userId string, time int64) error {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		log.Printf("DB error : creating transaction failed : f(UpdateLastMessageSeenInConversationForUser) : error : %v", err)
		return err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	// get last message seen at for conversation and user
	lastMessageSeenTime, err := qtx.getLastMessageSeenTimeByUserInConversation(ctx, getLastMessageSeenTimeByUserInConversationParams{
		UserID:         userId,
		ConversationID: conversationId,
	})

	if err != nil {
		log.Printf("DB error : error getting last message seen at : f(UpdateLastMessageSeenInConversationForUser) : error -> %v", err)
		return err
	}

	// check if provided timestamp > current in DB
	if lastMessageSeenTime > time {
		log.Printf("incorrect params for time : f(UpdateLastMessageSeenInConversationForUser)")
		return errors.New("wrong time provided")
	}

	// update last message seen at
	err = qtx.updateLastMessageSeenAtInConversationParticipant(ctx, updateLastMessageSeenAtInConversationParticipantParams{
		ConversationID:    conversationId,
		UserID:            userId,
		LastMessageSeenAt: time,
	})

	if err != nil {
		log.Printf("DB error : error updating last message seen at : f(UpdateLastMessageSeenInConversationForUser) : error -> %v", err)
		return err
	}

	// update seen count for all messages between timestamps
	err = qtx.updateSeenCountForMessagesBetweenTimestamps(ctx, updateSeenCountForMessagesBetweenTimestampsParams{
		CreatedAt:   lastMessageSeenTime,
		CreatedAt_2: time,
	})

	if err != nil {
		log.Printf("DB error : error updating seen count of messages : f(UpdateLastMessageSeenInConversationForUser) : error -> %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : error committing transaction : f(UpdateLastMessageSeenInConversationForUser) : error -> %v", err)
		return err
	}

	return nil
}

// Gets all user in a conversation
func (db *ChatQueries) GetAllUsersInConversation(ctx context.Context, conversationId string) ([]string, error) {
	conversationParticipant, err := db.Queries.getAllUsersInConversation(ctx, conversationId)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(conversationParticipant))
	for i, participant := range conversationParticipant {
		ids[i] = participant.UserID
	}
	return ids, nil
}

// Gets all user in a conversation
func (db *ChatQueries) GetAllUsersInConversationWS(ctx context.Context, conversationId string) ([]string, error) {
	conversationParticipant, err := db.Queries.getAllUsersInConversation(ctx, conversationId)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(conversationParticipant))
	for i, participant := range conversationParticipant {
		ids[i] = participant.UserID
	}
	return ids, nil
}

// Gets most recent conversation for a user after the given timestamp.
func (db *ChatQueries) GetMostRecentConversationsForUser(ctx context.Context, userId string, timestamp int64, numRows uint) ([]Conversation, error) {
	db.Queries.getMostRecentConversationsForUser(ctx, getMostRecentConversationsForUserParams{
		UserID:        userId,
		LastMessageAt: timestamp,
	})
	return nil, nil
}

// Gets all Messages after a given TIME for a given conversation. only numRows are returned
func (db *ChatQueries) GetAllMessagesForConversation(ctx context.Context, conversationId string, time int64, numRows uint) ([]Conversation, error) {
	db.Queries.getAllMessagesForConversation(ctx, getAllMessagesForConversationParams{
		ConversationID: conversationId,
		CreatedAt:      time,
		Limit:          int32(numRows),
	})
	return nil, nil
}

// Get all messages for a conversation after a timestamp
func (db *ChatQueries) IsUserPartOfConversation(ctx context.Context, userId string, convId string) (bool, error) {
	val, err := db.Queries.isUserPartOfConversation(ctx, isUserPartOfConversationParams{
		UserID:         userId,
		ConversationID: convId,
	})

	if err != nil {
		log.Print("error getting info from DB")
		return false, err
	}

	return val, nil
}

func (db *ChatQueries) CreateNewConversationWithUser(ctx context.Context, userId1 string, userId2 string) (string, error) {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		log.Printf("DB error : creating transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	conversation, err := createConversationWithUserHelperWithTransaction(ctx, userId1, userId2, qtx)

	if err != nil {
		log.Printf("DB error : creating conversation failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : commiting transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}

	return conversation.ID, nil
}

func createConversationWithUserHelperWithTransaction(ctx context.Context, userId1, userId2 string, qtx *Queries) (Conversation, error) {
	var convId string
	if userId1 < userId2 {
		convId = userId1 + "," + userId2
	} else {
		convId = userId2 + "," + userId1
	}

	currentTime := time.Now().Unix()

	conversation, err := qtx.createConversation(ctx, createConversationParams{
		ID:      convId,
		IsGroup: false,
		Name: sql.NullString{
			String: "",
			Valid:  false,
		},
		Description: sql.NullString{
			String: "",
			Valid:  false,
		},
		ImageUrl: sql.NullString{
			String: "",
			Valid:  false,
		},
		CreatedAt: currentTime,
		OwnerID: sql.NullString{
			String: "",
			Valid:  false,
		},
	})

	if err != nil {
		log.Printf("DB error : creating conversation failed : f(CreateNewConversationWithUser) : error : %v", err)
		return Conversation{}, err
	}

	err = qtx.createConversationParticipant(ctx, createConversationParticipantParams{
		ConversationID:    conversation.ID,
		UserID:            userId1,
		JoinedAt:          currentTime,
		LastMessageSeenAt: 0,
	})

	if err != nil {
		log.Printf("DB error : creating conversation participant 1 failed : f(CreateNewConversationWithUser) : error : %v", err)
		return Conversation{}, err
	}

	err = qtx.createConversationParticipant(ctx, createConversationParticipantParams{
		ConversationID:    conversation.ID,
		UserID:            userId2,
		JoinedAt:          currentTime,
		LastMessageSeenAt: 0,
	})

	if err != nil {
		log.Printf("DB error : creating conversation participant 2 failed : f(CreateNewConversationWithUser) : error : %v", err)
		return Conversation{}, err
	}

	return conversation, nil
}

func (db *ChatQueries) CreateNewGroupConversation(ctx context.Context, creatorId string, participants []string) (string, error) {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		log.Printf("DB error : creating transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	currentTime := time.Now().Unix()

	conversation, err := qtx.createConversation(ctx, createConversationParams{
		ID:      uuid.NewString(),
		IsGroup: false,
		Name: sql.NullString{
			String: "",
			Valid:  false,
		},
		Description: sql.NullString{
			String: "",
			Valid:  false,
		},
		ImageUrl: sql.NullString{
			String: "",
			Valid:  false,
		},
		CreatedAt: currentTime,
		OwnerID: sql.NullString{
			String: creatorId,
			Valid:  true,
		},
	})

	if err != nil {
		log.Printf("DB error : creating conversation failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}

	var isOwner bool
	insertedRequestor := false
	for i, participant := range participants {
		if participant == creatorId {
			isOwner = true
			insertedRequestor = true
		} else {
			isOwner = false
		}

		err = qtx.createConversationParticipant(ctx, createConversationParticipantParams{
			ConversationID:    conversation.ID,
			UserID:            participant,
			JoinedAt:          currentTime,
			IsOwner:           isOwner,
			LastMessageSeenAt: 0,
		})

		if err != nil {
			log.Printf("DB error : creating conversation participant %v failed : f(CreateNewConversationWithUser) : error : %v", i, err)
			return "", err
		}
	}

	if !insertedRequestor {
		err = qtx.createConversationParticipant(ctx, createConversationParticipantParams{
			ConversationID:    conversation.ID,
			UserID:            creatorId,
			JoinedAt:          currentTime,
			IsOwner:           true,
			LastMessageSeenAt: 0,
		})

		if err != nil {
			log.Printf("DB error : creating conversation participant creator failed : f(CreateNewConversationWithUser) : error : %v", err)
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("DB error : commiting transaction failed : f(CreateNewConversationWithUser) : error : %v", err)
		return "", err
	}

	return conversation.ID, nil
}
