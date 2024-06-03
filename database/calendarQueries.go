package database

import (
	"context"
	"database/sql"
	"g_chat/models"
	"time"

	"github.com/google/uuid"
)

type CalendarQueries struct {
	*Queries
}

func GetCalendarQueries() *CalendarQueries {
	queries := getQueries()
	return &CalendarQueries{queries}
}

func (db *CalendarQueries) CreateNewCalendarRequest(ctx context.Context, userId string, eventId string, description string) (Calendareventrequest, error) {
	createdEvent, err := db.Queries.createNewCalendarEventRequest(ctx, createNewCalendarEventRequestParams{
		ID:               uuid.NewString(),
		RequestingUserID: userId,
		EventID:          eventId,
		RequestMessage:   description,
		CreatedAt:        time.Now().UnixNano(),
	})

	if err != nil {
		return Calendareventrequest{}, err
	}

	return createdEvent, nil
}

func (db *CalendarQueries) UpdateCalendarRequestStatusChange(ctx context.Context, requestId string, updateType string) (Calendareventrequest, error) {
	updatedRequest, err := db.Queries.updateCalendarRequest(ctx, updateCalendarRequestParams{
		ID: requestId,
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: time.Now().UnixNano(),
		},
		RequestStatus: updateType,
	})

	if err != nil {
		return Calendareventrequest{}, err
	}

	return updatedRequest, nil
}

func (db *CalendarQueries) CreateNewCalendarEvent(ctx context.Context, calendarEvent models.NewCalendarEvent) (Calendarevent, error) {
	newCalendarEvent, err := db.Queries.createNewCalendarEvent(ctx, createNewCalendarEventParams{
		ID:               uuid.NewString(),
		UserID:           calendarEvent.UserID,
		EventTitle:       calendarEvent.EventTitle,
		EventDescription: calendarEvent.EventDescription,
		FromTime:         calendarEvent.FromTime,
		ToTime:           calendarEvent.ToTime,
		IsRecurring:      calendarEvent.IsRecurring,
		GameID:           calendarEvent.GameID,
		CreatedAt:        time.Now().UnixNano(),
	})

	// TODO : create transaction and also update calendar participants + a conversation for this

	if err != nil {
		return Calendarevent{}, nil
	}

	return newCalendarEvent, nil
}

func (db *CalendarQueries) GetRecentEventsCreatedByUser(ctx context.Context, userId string, rowCount uint, time int64) ([]Calendarevent, error) {
	events, err := db.Queries.getScheduledEventsCreatedByUser(ctx, getScheduledEventsCreatedByUserParams{
		UserID:   userId,
		Limit:    int32(rowCount),
		FromTime: time,
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (db *CalendarQueries) GetCalendarEventsScheduledForUser(ctx context.Context, userId string, rowCount uint, time int64) ([]Calendarevent, error) {
	events, err := db.Queries.getScheduledEventsForUser(ctx, getScheduledEventsForUserParams{
		UserID:   userId,
		Limit:    int32(rowCount),
		FromTime: time,
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (db *CalendarQueries) IsUserOrganizerOfEvent(ctx context.Context, userId string, eventId string) (bool, error) {
	val, err := db.Queries.fIsUserOrganizerOfEvent(ctx, fIsUserOrganizerOfEventParams{
		UserID: userId,
		ID:     eventId,
	})

	if err != nil {
		return false, err
	}

	return val, nil
}

func (db *CalendarQueries) DeleteCalendarEvent(ctx context.Context, eventId string) error {
	if _, err := db.Queries.organizerRequestToDeleteEvent(ctx, eventId); err != nil {
		return err
	}

	return nil
}

func (db *CalendarQueries) IsCalendarEventExistsInFuture(ctx context.Context, eventId string) (bool, error) {
	calendarevent, err := db.Queries.getCalendarEventForId(ctx, eventId)

	if err != nil {
		return false, err
	}

	if calendarevent.ToTime < time.Now().UnixNano() {
		return false, nil
	}

	return true, nil
}

func (db *CalendarQueries) IsUserRequestOnEventExists(ctx context.Context, userId string, eventId string) (bool, error) {
	val, err := db.Queries.isUserRequestOnEventExists(ctx, isUserRequestOnEventExistsParams{
		RequestingUserID: userId,
		EventID:          eventId,
	})

	if err != nil {
		return false, err
	}

	return val, nil
}

func (db *CalendarQueries) AreUserAlreadyAParticipantOfCalendarEvent(ctx context.Context, userIds []string, eventId string) (bool, error) {
	retVal := true
	for _, userId := range userIds {
		val, err := db.Queries.fIsUserAlreadyAParticipantOfCalendarEvent(ctx, fIsUserAlreadyAParticipantOfCalendarEventParams{
			EventID: eventId,
			UserID:  userId,
		})

		if err != nil {
			return false, err
		}
		retVal = retVal && val
	}

	return retVal, nil
}

func (db *CalendarQueries) GetCalendarEventParticipants(ctx context.Context, eventId string, rowCount int32, time int64) ([]models.CalendarParticipant, error) {
	participants, err := db.Queries.getAllCalendarEventParticipants(ctx, getAllCalendarEventParticipantsParams{
		EventID:  eventId,
		JoinedAt: time,
		Limit:    rowCount,
	})

	if err != nil {
		return nil, err
	}

	retValue := make([]models.CalendarParticipant, len(participants))
	for i, participant := range participants {
		retValue[i] = models.CalendarParticipant{
			EventID:     participant.EventID,
			UserID:      participant.UserID,
			ImageUrl:    participant.ImageUrl,
			JoinedAt:    participant.JoinedAt,
			IsOrganizer: participant.IsOrganizer,
			Name:        participant.Name,
		}
	}

	return retValue, nil
}

func (db *CalendarQueries) RemoveUserFromCalendarEvent(ctx context.Context, participants []string, eventId string) error {
	for _, participant := range participants {
		_, err := db.Queries.userRequestToLeaveEvent(ctx, userRequestToLeaveEventParams{
			EventID: eventId,
			UserID:  participant,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *CalendarQueries) UpdateCalendarEventDetails(ctx context.Context, eventTitle string, eventDescription string, fromTime int64, toTime int64, gameId string) (Calendarevent, error) {
	updatedEvent, err := db.Queries.updateCalendarEventDetails(ctx, updateCalendarEventDetailsParams{
		EventTitle:       eventTitle,
		EventDescription: eventDescription,
		FromTime:         fromTime,
		ToTime:           toTime,
		GameID:           gameId,
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: time.Now().UnixNano(),
		},
	})

	if err != nil {
		return Calendarevent{}, err
	}
	return updatedEvent, nil
}

func (db *CalendarQueries) GetAllActiveRequestsForEvent(ctx context.Context, eventId string, rowCount int32, time int64) ([]Calendareventrequest, error) {
	activeRequests, err := db.Queries.getAllCalendarRequestForEvent(ctx, getAllCalendarRequestForEventParams{
		EventID:       eventId,
		RequestStatus: "ACTIVE",
		Limit:         rowCount,
		CreatedAt:     time,
	})

	if err != nil {
		return nil, err
	}

	return activeRequests, nil
}

func (db *CalendarQueries) CreateNewCalendarInviteWithTransaction(qtx *Queries, ctx context.Context, eventId string, userId string, inviteMessage string) (Calendareventinvite, error) {
	invite, err := qtx.createNewCalendarInvite(ctx, createNewCalendarInviteParams{
		EventID:       eventId,
		InvitedUserID: userId,
		InviteMessage: inviteMessage,
		CreatedAt:     time.Now().UnixNano(),
		ID:            uuid.NewString(),
	})

	if err != nil {
		return Calendareventinvite{}, err
	}

	return invite, nil
}

func (db *CalendarQueries) CreateNewCalendarInvite(ctx context.Context, eventId string, userId string, inviteMessage string) (Calendareventinvite, error) {
	invite, err := db.Queries.createNewCalendarInvite(ctx, createNewCalendarInviteParams{
		EventID:       eventId,
		InvitedUserID: userId,
		InviteMessage: inviteMessage,
		CreatedAt:     time.Now().UnixNano(),
		ID:            uuid.NewString(),
	})

	if err != nil {
		return Calendareventinvite{}, err
	}

	return invite, nil
}

func (db *CalendarQueries) DeleteCalendarInvite(ctx context.Context, inviteId string) error {
	if err := db.Queries.deleteCalendarInvite(ctx, inviteId); err != nil {
		return err
	}

	return nil
}

func (db *CalendarQueries) GetAllInvitesForUser(ctx context.Context, userId string, time int64, inviteStatus string, rowCount int) ([]Calendareventinvite, error) {
	invites, err := db.Queries.getAllCalendarInvitesForUser(ctx, getAllCalendarInvitesForUserParams{
		InvitedUserID: userId,
		Limit:         int32(rowCount),
		CreatedAt:     time,
		InviteStatus:  inviteStatus,
	})

	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (db *CalendarQueries) GetAllInvitesForEvent(ctx context.Context, eventId string, time int64, inviteStatus string, rowCount int) ([]Calendareventinvite, error) {
	invites, err := db.Queries.getAllCalendarInvitesForEvent(ctx, getAllCalendarInvitesForEventParams{
		EventID:      eventId,
		Limit:        int32(rowCount),
		CreatedAt:    time,
		InviteStatus: inviteStatus,
	})

	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (db *CalendarQueries) IsEventTimeCoincidesWithAnotherEvent(ctx context.Context, userId string, fromTime int64, toTime int64) (bool, error) {
	// TODO
	return false, nil
}
