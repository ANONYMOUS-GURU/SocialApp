package controllers

import (
	"g_chat/database"
	"g_chat/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

//  1. create a new event
// 	2. delete an event
// 	3. request for join
// 	4. get my list of recent events joined or created
//  5. get events created by user
// 	6. check if a event coincides with another event
// 	7. accept/decline an event request
// 	8. get participants for an event
// 	9. ping all participants of an event
// 10. get all active requests for event
// 12. leave an event for a joinee
// 13. update event
// 14. remove participants for event

// TODO : add conversation
// TODO : add ability to add friends in events from start
// TODO : implementing invites

func CreateNewCalendarEvent(ctx *gin.Context) {
	var calendarEvent models.NewCalendarEvent
	if err := ctx.Bind(&calendarEvent); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect params, unable to parse",
		})
		return
	}

	// check if time is correct
	if calendarEvent.FromTime >= calendarEvent.ToTime {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect params, from time >= to time",
		})
		return
	}

	// authorized
	if calendarEvent.UserID != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	// TODO : game id not valid - return

	// return the created calendar event
	createdCalendarEvent, err := database.GetCalendarQueries().CreateNewCalendarEvent(ctx.Request.Context(), calendarEvent)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error creating calendar request",
		})
		return
	}

	ctx.JSON(http.StatusOK, createdCalendarEvent)
}

func UpdateCalendarRequestStatus(ctx *gin.Context) {
	var updateCalendarRequest models.UpdateRequestStatus
	if err := ctx.Bind(&updateCalendarRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect params, unable to parse",
		})
		return
	}

	if updateCalendarRequest.UpdaterId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetCalendarQueries().IsUserOrganizerOfEvent(ctx.Request.Context(), updateCalendarRequest.UpdaterId, updateCalendarRequest.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if updateCalendarRequest.ApprovalStatus == "ACCEPT" {
		val, err = database.GetCalendarQueries().AreUserAlreadyAParticipantOfCalendarEvent(ctx.Request.Context(), []string{updateCalendarRequest.RequestorId}, updateCalendarRequest.EventId)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error creating calendar request",
			})
			return
		}

		if !val {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "user already part of Event",
			})
			return
		}
	}

	_, err = database.GetCalendarQueries().UpdateCalendarRequestStatusChange(ctx.Request.Context(), updateCalendarRequest.RequestId, updateCalendarRequest.ApprovalStatus)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error creating calendar request",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}

func DeleteExisitngCalendarEvent(ctx *gin.Context) {
	var deleteCalendarEvent models.DeleteCalendarEvent
	if err := ctx.Bind(&deleteCalendarEvent); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect params, unable to parse",
		})
		return
	}

	if deleteCalendarEvent.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetCalendarQueries().IsUserOrganizerOfEvent(ctx.Request.Context(), deleteCalendarEvent.UserId, deleteCalendarEvent.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error creating calendar request",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "user not organizer of event",
		})
		return
	}

	err = database.GetCalendarQueries().DeleteCalendarEvent(ctx, deleteCalendarEvent.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error creating calendar request",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}

func RequestJoinForCalendarEvent(ctx *gin.Context) {
	var calendarRequest models.NewCalendarEventRequest
	if err := ctx.BindJSON(&calendarRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	// check if userId same as requestingUserId
	if ctx.Keys["userId"] != calendarRequest.UserId {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "wrong time params",
		})
		return
	}

	// check if eventId exists and is not a past event
	val, err := database.GetCalendarQueries().IsCalendarEventExistsInFuture(ctx.Request.Context(), calendarRequest.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error checking DB",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "event id not allowed",
		})
		return
	}

	// already active request
	val, err = database.GetCalendarQueries().IsUserRequestOnEventExists(ctx.Request.Context(), calendarRequest.UserId, calendarRequest.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error writing to DB",
		})
		return
	}

	if val {
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "request already exists",
		})
	}

	// already a participant
	val, err = database.GetCalendarQueries().AreUserAlreadyAParticipantOfCalendarEvent(ctx.Request.Context(), []string{calendarRequest.UserId}, calendarRequest.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error writing to DB",
		})
		return
	}

	if val {
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "already a participant",
		})
	}

	_, err = database.GetCalendarQueries().CreateNewCalendarRequest(ctx.Request.Context(), calendarRequest.UserId, calendarRequest.EventId, calendarRequest.Message)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "error writing to DB",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

func GetRecentEventsJoinedOrCreatedForUser(ctx *gin.Context) {
	var recentEventsRequestBody models.GetCalendarEventsRequestParam

	if err := ctx.Bind(&recentEventsRequestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if ctx.Keys["userId"] != recentEventsRequestBody.UserId {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if recentEventsRequestBody.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "row count must be > 0",
		})
		return
	}

	events, err := database.GetCalendarQueries().GetCalendarEventsScheduledForUser(ctx.Request.Context(), recentEventsRequestBody.UserId, recentEventsRequestBody.RowCount, recentEventsRequestBody.Time)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, events)
}

func GetRecentEventsCreatedByUser(ctx *gin.Context) {
	var recentEventsRequestBody models.GetCalendarEventsRequestParam

	if err := ctx.Bind(&recentEventsRequestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if ctx.Keys["userId"] != recentEventsRequestBody.UserId {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if recentEventsRequestBody.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "row count must be > 0",
		})
		return
	}

	events, err := database.GetCalendarQueries().GetRecentEventsCreatedByUser(ctx.Request.Context(), recentEventsRequestBody.UserId, recentEventsRequestBody.RowCount, recentEventsRequestBody.Time)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, events)
}

// func GetEventsJoinedForUser(ctx *gin.Context) {

// }

func IsEventCoincidesWithAnotherJoinedEvent(ctx *gin.Context) {
	var eventCoincideRequest models.EventCoincideRequest
	if err := ctx.Bind(&eventCoincideRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if ctx.Keys["userId"] != eventCoincideRequest.UserId {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	events, err := database.GetCalendarQueries().IsEventTimeCoincidesWithAnotherEvent(ctx.Request.Context(), eventCoincideRequest.UserId, eventCoincideRequest.FromTime, eventCoincideRequest.ToTime)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, events)

}

func GetParticipantsOfAnEvent(ctx *gin.Context) {
	var getEventParticipants models.GetCalendarEventParticipants
	if err := ctx.Bind(&getEventParticipants); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if getEventParticipants.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if getEventParticipants.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "row count must be > 0 ",
		})
		return
	}

	// user part of event
	val, err := database.GetCalendarQueries().AreUserAlreadyAParticipantOfCalendarEvent(ctx.Request.Context(), []string{getEventParticipants.UserId}, getEventParticipants.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	participants, err := database.GetCalendarQueries().GetCalendarEventParticipants(ctx.Request.Context(), getEventParticipants.EventId, getEventParticipants.RowCount, getEventParticipants.Time)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, participants)

}

// TODO : create a new conversation for every event, add one column in event - conversationId (stores the conversation of the group) for every addition and deletion to group update this conversation
func PingAllParticipantsOfAnEvent(ctx *gin.Context) {
}

func GetAllActiveRequestsForEvent(ctx *gin.Context) {
	var getActiveRequests models.GetAllCalendarRequests
	if err := ctx.Bind(&getActiveRequests); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if getActiveRequests.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if getActiveRequests.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "row count must be > 0",
		})
		return
	}

	val, err := database.GetCalendarQueries().IsUserOrganizerOfEvent(ctx.Request.Context(), getActiveRequests.UserId, getActiveRequests.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	requests, err := database.GetCalendarQueries().GetAllActiveRequestsForEvent(ctx.Request.Context(), getActiveRequests.EventId, getActiveRequests.RowCount, getActiveRequests.Time)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

func LeaveEventForAParticipant(ctx *gin.Context) {
	var eventLeave models.UserEventParam
	if err := ctx.Bind(&eventLeave); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if eventLeave.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetCalendarQueries().AreUserAlreadyAParticipantOfCalendarEvent(ctx.Request.Context(), []string{eventLeave.UserId}, eventLeave.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	err = database.GetCalendarQueries().RemoveUserFromCalendarEvent(ctx.Request.Context(), []string{eventLeave.UserId}, eventLeave.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func UpdateEventDetails(ctx *gin.Context) {
	var updateCalendarEventDetails models.UpdateCalendarEventDetails
	if err := ctx.Bind(&updateCalendarEventDetails); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if updateCalendarEventDetails.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetCalendarQueries().IsUserOrganizerOfEvent(ctx.Request.Context(), updateCalendarEventDetails.UserId, updateCalendarEventDetails.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	updatedEvent, err := database.GetCalendarQueries().UpdateCalendarEventDetails(ctx.Request.Context(), updateCalendarEventDetails.EventTitle, updateCalendarEventDetails.EventDescription, updateCalendarEventDetails.FromTime, updateCalendarEventDetails.ToTime, updateCalendarEventDetails.GameId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedEvent)
}

func RemoveCalendarParticipantByOrganizer(ctx *gin.Context) {
	var removeCalendarParticipant models.RemoveCalendarParticipant
	if err := ctx.Bind(&removeCalendarParticipant); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "unable to parse json",
		})
		return
	}

	if removeCalendarParticipant.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetCalendarQueries().IsUserOrganizerOfEvent(ctx.Request.Context(), removeCalendarParticipant.UserId, removeCalendarParticipant.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err = database.GetCalendarQueries().AreUserAlreadyAParticipantOfCalendarEvent(ctx.Request.Context(), removeCalendarParticipant.ParticipantId, removeCalendarParticipant.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	if !val {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	err = database.GetCalendarQueries().RemoveUserFromCalendarEvent(ctx.Request.Context(), removeCalendarParticipant.ParticipantId, removeCalendarParticipant.EventId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
