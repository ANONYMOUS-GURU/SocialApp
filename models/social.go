package models

type NewSocialRequestEvent struct {
	RequestorId  string
	TargetUserId string
	RequestType  string
	Message      string
	CreatedAt    int64
}

type FriendRequestEvent struct {
	ID             string `json:"id"`
	SenderId       string `json:"sender_id"`
	ReceiverId     string `json:"receiver_id"`
	Description    string `json:"description"`
	DeliveryStatus uint   `json:"delivery_status"`
	CreatedAt      int64  `json:"created_at"`
}

type SocialRequestUpdateEvent struct {
	ID          string `json:"id"`
	UpdaterId   string `json:"acceptor_id"`
	RequestorId string
	UpdateType  string `json:"update_type"`
	RequestType string `json:"request_type"`
}

type DisconnectSocialUsers struct {
	User1Id        string
	User2Id        string
	ConnectionType string
}

type GetAllSocialConnections struct {
	UserId   string
	RowCount uint
	Time     int64
}

type PairUsers struct {
	User1Id string
	User2Id string
}

type GetSocialRequests struct {
	UserId   string
	RowCount uint
	Time     int64
}

type SocialUser struct {
	UserId    string
	Name      string
	ImageUrl  string
	CreatedAt int64
}
