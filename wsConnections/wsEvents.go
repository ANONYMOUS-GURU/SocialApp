package websockets

type EventHandler func(event Event, client *Client) error
type EventType uint
type NotificationType uint

const (
	/*
		when entire payload could not parsed
	*/
	Unknown EventType = iota

	/*
		fired back to client for every event type received, payload's event type will signify the type of
		event it acknowledges along with an ID to keep track on client side which event has been received
	*/
	AckEvent

	/*
		events related to incoming messages, their delivery and seen status, note ack will be fired for
		all the events
	*/
	EventIncomingChatMessage
	EventIncomingDeliveredUpdate
	EventIncomingReadUpdate

	EventOutgoingChatMessage
	EventOutgoingReadUpdate
	EventOutgoingDeliveredUpdate

	// NOT IMPLEMENTED---------------------------------------------------------------------------------------------------------
	EventIncomingUserStatusChange
	EventNotifyFriendStatusChange
	EventFailedMessageRetry
	// LFG feed (entirely in websockets)
	// friends status(online/offline) and game playing
)

type UserOnlineStatus uint

const (
	Offline UserOnlineStatus = 0
	Online
)
