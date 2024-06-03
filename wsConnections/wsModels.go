package websockets

type IncomingUserStatusChange struct {
	ID     string           `json:"id"`
	Status UserOnlineStatus `json:"status"`
}

type Acknowledge struct {
	ReceiverID string    `json:"sender_id"`
	EventType  EventType `json:"event_type"`
	Status     bool      `json:"status"`
	Message    string    `json:"message"`
	AckTime    int64     `json:"ack_time"`
}

type IncomingReadUpdate struct {
	ID             string `json:"id"`
	SenderId       string `json:"sender_id"`
	ConversationId string `json:"conversation_id"`
	Time           int64  `json:"time"`
}

type OutgoingReadUpdate struct {
	MessageId  string `json:"message_id"`
	ReceiverID string `json:"receiver_id"`
	Time       int64  `json:"time"`
}

type OutgoingDeliveredUpdate struct {
	MessageId  string `json:"message_id"`
	ReceiverID string `json:"receiver_id"`
	Time       int64  `json:"time"`
}

type IncomingDeliveredUpdate struct {
	MessageId string `json:"message_id"`
	SenderId  string `json:"receiver_id"`
	Time      int64  `json:"time"`
}

type IncomingSocialRequest struct {
	RequestType  string `json:"request_type"`
	RequestorId  string `json:"requestor_id"`
	TargetUserId string `json:"target_user_id"`
	Description  string `json:"description"`
}

type IncomingSocialRequestUpdate struct {
	RequestId     string `json:"request_id"`
	UpdatedBy     string `json:"updated_by"`
	ModifiedState string `json:"modified_state"`
}
