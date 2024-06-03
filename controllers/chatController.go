package controllers

import (
	"context"
	"errors"
	"g_chat/database"
	"g_chat/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMostRecentConversationsForUser(ctx *gin.Context) {
	time, err := strconv.ParseInt(ctx.Query("lastTimestamp"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "insufficient or wrong params",
		})
		return
	}
	queryCount, err := strconv.Atoi(ctx.Query("queryCount"))
	if err != nil || queryCount <= 0 {
		return
	}

	if queryCount > 20 {
		queryCount = 20
	}

	conversations, err := database.GetChatQueries().GetMostRecentConversationsForUser(ctx, ctx.Keys["userId"].(string), time, uint(queryCount))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unknown",
		})
		return
	}

	ctx.JSON(http.StatusOK, conversations)
}

func GetAllMessagesForConversation(ctx *gin.Context) {
	time, err := strconv.ParseInt(ctx.Query("lastTimestamp"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "insufficient or wrong params",
		})
		return
	}
	queryCount, err := strconv.Atoi(ctx.Query("queryCount"))
	if err != nil || queryCount <= 0 {
		return
	}

	if queryCount > 20 {
		queryCount = 20
	}

	convId := ctx.Query("conversationId")

	val, err := IsUserPartOfConversation(ctx.Keys["userId"].(string), convId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching data from DB",
		})
		return
	}

	if !val {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Error fetching data from DB",
		})
		return
	}

	messages, err := database.GetChatQueries().GetAllMessagesForConversation(ctx, convId, time, uint(queryCount))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching data from DB",
		})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

func GetAllUsersInConversation(ctx *gin.Context) {
	convId := ctx.Query("conversationId")

	val, err := IsUserPartOfConversation(ctx.Keys["userId"].(string), convId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching data from DB",
		})
		return
	}

	if !val {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Error fetching data from DB",
		})
		return
	}

	users, err := database.GetChatQueries().GetAllUsersInConversation(ctx, convId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unknown",
		})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func GetAllUsersInConversationWS(userId string, conversationId string) ([]string, error) {

	val, err := IsUserPartOfConversation(userId, conversationId)

	if err != nil {
		log.Print("error getting user conversation from DB")
		return nil, err
	}

	if !val {
		log.Print("user not part of conversation")
		return nil, errors.New("unauthorized")
	}

	users, err := database.GetChatQueries().GetAllUsersInConversationWS(context.Background(), conversationId)

	if err != nil {
		log.Print("error retreiving DB")
		return nil, err
	}

	return users, nil
}

func IsUserPartOfConversation(userId string, convId string) (bool, error) {
	val, err := database.GetChatQueries().IsUserPartOfConversation(context.Background(), userId, convId)

	if err != nil {
		return false, err
	}

	return val, nil
}

func getUnsentRequestsQueryParam(ctx *gin.Context) (int64, uint, error) {
	time, err := strconv.ParseInt(ctx.Query("lastTimestamp"), 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid last timestamp param")
	}

	queryCount, err := strconv.Atoi(ctx.Query("queryCount"))
	if err != nil || queryCount <= 0 {
		return 0, 0, errors.New("invalid query count param")
	}

	if queryCount > 20 {
		queryCount = 20
	}

	return time, uint(queryCount), nil
}

func GetUnsentMessages(ctx *gin.Context) {
	lastTimeStamp, queryCount, err := getUnsentRequestsQueryParam(ctx)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Params" + err.Error(),
		})
	}

	conversationId := ctx.Query("conversationId")

	unsentMessages, isEnd, err := database.GetChatQueries().GetUnsentMessages(ctx, lastTimeStamp, queryCount, conversationId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Unknown error while retreiving data",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"lastMessage": isEnd,
		"response":    unsentMessages,
	})
}

func GetAllMessages(ctx *gin.Context) {
	lastTimeStamp, queryCount, err := getUnsentRequestsQueryParam(ctx)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Params" + err.Error(),
		})
	}

	messages, err := database.GetChatQueries().GetAllMessagesAfterGivenTime(ctx, ctx.Keys["userId"].(string), lastTimeStamp, queryCount)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Unknown error while retreiving data",
		})
	}

	isEnd := false
	if len(messages) < int(queryCount) {
		isEnd = true
	}

	ctx.JSON(http.StatusOK, gin.H{
		"lastMessage": isEnd,
		"response":    messages,
	})
}

func MarkMessagesAsRecievedByUser(ctx *gin.Context) {
	var messageIds []string
	if err := ctx.BindJSON(&messageIds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "request body invalid",
		})
		return
	}

	// TODO : confirm user is the recipient of all these messages

	if _, err := database.GetChatQueries().MarkMessagesAsRecievedByUser(ctx, messageIds, ctx.Keys["userId"].(string)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "error writing to DB",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func MarkMessageAsSeenByUser(ctx *gin.Context) {
	var message models.IncomingChatPayload
	if err := ctx.BindJSON(&message); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "request body invalid",
		})
		return
	}

	// TODO : confirm user is the recipient of all these messages

	if err := database.GetChatQueries().UpdateLastMessageSeenInConversationForUser(ctx, message.ConversationId, ctx.Keys["userId"].(string), message.SentAt); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "error writing to DB",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

// func GetUnsentCalendarRequests(ctx *gin.Context) {
// 	lastTimeStamp, queryCount, err := GetUnsentRequestsQueryParam(ctx)

// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid Params" + err.Error(),
// 		})
// 	}

// 	unsentCalendarRequests, isEnd, err := database.GetUnsentCalendarRequests(ctx, lastTimeStamp, queryCount)

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Unknown error while retreiving data",
// 		})
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"lastMessage": isEnd,
// 		"response":    unsentCalendarRequests,
// 	})
// }

// func GetUnsentFriendRequests(ctx *gin.Context) {
// 	lastTimeStamp, queryCount, err := GetUnsentRequestsQueryParam(ctx)

// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid Params" + err.Error(),
// 		})
// 	}

// 	unsentFriendRequests, isEnd, err := database.GetUnsentFriendRequests(ctx, lastTimeStamp, queryCount)

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Unknown error while retreiving data",
// 		})
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"lastMessage": isEnd,
// 		"response":    unsentFriendRequests,
// 	})
// }

// func GetAllFriendRequests(ctx *gin.Context) {
// 	lastTimeStamp, queryCount, err := GetUnsentRequestsQueryParam(ctx)

// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid Params" + err.Error(),
// 		})
// 	}

// 	friendRequests, isEnd, err := database.GetAllFriendRequests(ctx, lastTimeStamp, queryCount)

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Unknown error while retreiving data",
// 		})
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"lastMessage": isEnd,
// 		"response":    friendRequests,
// 	})
// }

// func GetAllCalendarRequests(ctx *gin.Context) {
// 	lastTimeStamp, queryCount, err := GetUnsentRequestsQueryParam(ctx)

// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid Params" + err.Error(),
// 		})
// 	}

// 	calendarRequests, isEnd, err := database.GetAllCalendarRequests(ctx, lastTimeStamp, queryCount)

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "Unknown error while retreiving data",
// 		})
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"lastMessage": isEnd,
// 		"response":    calendarRequests,
// 	})
// }
