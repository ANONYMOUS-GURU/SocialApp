package models

type Notifications struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	SenderId       string `json:"sender_id"`
	ReceiverId     string `json:"receiver_id"`
	DeliveryStatus uint   `json:"delivery_status"`
	CreatedAt      int64  `json:"created_at"`
}
