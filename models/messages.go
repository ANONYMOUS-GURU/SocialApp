package models

// Non DB models
type IncomingChatPayload struct {
	ID             string `json:"id"` // for inconimg message - a temp guid to be used in frontend for mapping
	MessageBody    string `json:"message_body"`
	SenderId       string `json:"sender_id"`
	ConversationId string `json:"conversation_id"`
	IsGroup        bool   `json:"is_group"`
	ReceiverId     string `json:"receiver_id"`
	SentAt         int64  `json:"sent_at"`
}

type OutgoingChatPayload struct {
	ID                string `json:"id"` // for inconimg message - a temp guid to be used in frontend for mapping
	MessageBody       string `json:"message_body"`
	Sender            string `json:"sender"`
	ConversationId    string `json:"conversation_id"`
	SentAt            int64  `json:"sent_at"`
	ReceiverId        string `json:"receiver_id"`
	IsGroup           bool   `json:"is_group"`
	SenderId          string `json:"sender_id"`
	ServerRecieveTime int64  `json:"server_recieve_time"`
}
