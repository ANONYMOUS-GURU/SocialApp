// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"database/sql"
)

type Calendarevent struct {
	ID               string
	UserID           string
	EventTitle       string
	EventDescription string
	FromTime         int64
	ToTime           int64
	IsRecurring      bool
	GameID           string
	CreatedAt        int64
	UpdatedAt        sql.NullInt64
	DeletedAt        sql.NullInt64
}

type Calendareventinvite struct {
	ID            string
	EventID       string
	InvitedUserID string
	InviteMessage string
	InviteStatus  string
	CreatedAt     int64
	UpdatedAt     sql.NullInt64
}

type Calendareventparticipant struct {
	EventID     string
	UserID      string
	JoinedAt    int64
	IsOrganizer bool
}

type Calendareventrequest struct {
	ID               string
	EventID          string
	RequestingUserID string
	RequestMessage   string
	RequestStatus    string
	CreatedAt        int64
	UpdatedAt        sql.NullInt64
}

type Conversation struct {
	ID            string
	IsGroup       bool
	OwnerID       sql.NullString
	Name          sql.NullString
	Description   sql.NullString
	ImageUrl      sql.NullString
	CreatedAt     int64
	UpdatedAt     sql.NullInt64
	DeletedAt     sql.NullInt64
	LastMessageAt int64
}

type Conversationparticipant struct {
	ConversationID    string
	UserID            string
	IsOwner           bool
	LastMessageSeenAt int64
	JoinedAt          int64
	DeletedAt         sql.NullInt64
}

type Follow struct {
	FollowerID string
	FollowedID string
	CreatedAt  int64
}

type Friend struct {
	User1ID   string
	User2ID   string
	CreatedAt int64
}

type Message struct {
	ID             string
	Body           string
	ConversationID string
	SenderID       string
	DeliveredCount int32
	SeenCount      int32
	SentToCount    int32
	SentAt         int64
	CreatedAt      int64
}

type Messageusermap struct {
	MessageID  string
	ReceiverID string
}

type Socialrequest struct {
	ID             string
	UserID         string
	TargetUserID   string
	RequestType    string
	RequestMessage string
	RequestStatus  string
	CreatedAt      int64
	UpdatedAt      sql.NullInt64
}

type User struct {
	ID          string
	Name        string
	Username    string
	Email       string
	Description string
	ImageUrl    string
	CreatedAt   int64
	UpdatedAt   sql.NullInt64
	DeletedAt   sql.NullInt64
}