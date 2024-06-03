package models

type GetCalendarEventsRequestParam struct {
	UserId   string
	Time     int64
	RowCount uint
}

type EventCoincideRequest struct {
	FromTime int64
	ToTime   int64
	UserId   string
}

type UserEventParam struct {
	UserId  string
	EventId string
}

type UpdateCalendarEventDetails struct {
	EventId          string
	UserId           string
	EventTitle       string
	EventDescription string
	FromTime         int64
	ToTime           int64
	GameId           string
}

type RemoveCalendarParticipant struct {
	UserId        string
	EventId       string
	ParticipantId []string
}

type UpdateRequestStatus struct {
	RequestId      string
	UpdaterId      string
	EventId        string
	RequestorId    string
	ApprovalStatus string
}

// TODO : add participants in new calendar event
type NewCalendarEvent struct {
	UserID           string
	EventTitle       string
	EventDescription string
	FromTime         int64
	ToTime           int64
	IsRecurring      bool
	GameID           string
}

type DeleteCalendarEvent struct {
	UserId  string
	EventId string
}

type NewCalendarEventRequest struct {
	UserId  string
	EventId string
	Message string
}

type CalendarParticipant struct {
	EventID     string
	UserID      string
	JoinedAt    int64
	IsOrganizer bool
	Name        string
	ImageUrl    string
}

type GetAllCalendarRequests struct {
	UserId   string
	EventId  string
	RowCount int32
	Time     int64
}

type GetCalendarEventParticipants struct {
	UserId   string
	EventId  string
	RowCount int32
	Time     int64
}
