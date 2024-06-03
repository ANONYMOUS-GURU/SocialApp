/*

conversations
users
messages
conversationParticipants
unsentmessages

user online -> addUser to a global connection list map[userId]ws -> check for unread messages [Read UnsentMessages table and send messages (if still part of that conversation)
 -> delete once acknowledged]

user sends message -> check access -> get Message/conversation message sent to -> confirm access -> write msg/conversation to Db (with not recieved)-> get all participants
					-> case 1 :  participant online -> send message -> wait for acknowledge -> write to db msg sent
					-> case 2 : participant offline -> write to separate Db (UnSentMessages)

Structure:
ConnectionManager{					Func -> removeConnection, addConnection
	map[userId]client
	sync.RWMutex

}

Client{								Func -> readMessage, WriteMessage [also have ping/pong]
	ws.Connection
	userId
	egress chan []byte
}

WebSockets tcp:
1. Messages
2. Friend request sent/recieve
3. Friend request accepted/declined send/recieve
4. Calendar request send/recieve
5. Calendar request accepted/declined send/recieve
6. LFG feed (entirely in websockets)
7. User feed ??
8. Friends status(online/offline) and game playing

voice UDP:

REST:
1. Auth+login
2. connection for WS
3. showing calendars of users/friends
4. user profile updates + user's game profile updates
5. Get request for :
	a. Messages
	b. Calendar Requests
	c. Friend Requests
	d. Sync State

Flow:
1. User login
2. Make Websocket connection
3. Make REST calls for most recent conversations for user
4. Make REST calls for messages from lastTimestamp <- can be conversation based as well
5. Make REST calls for requests/accepts from lastTimestamp
6. Once done for a

Table -> unsent msgs
Table -> msgs

sign up -> token -> /createUser (name, email, phone) ->
login (token) ->  /createTicket -> /ws -> /getAllConversations, /getConversationMessages,  /getunsentmessages, (/getmessages),

message from user A -> user B, first create conversation and put in participants, -> insert message

-> ws
json ->
ack ->
*/

/*
	apis for calendar -
	1. create a new event
	2. delete an event
	3. request for join
	4. get my list of recent events joined or created
	5. get events which i joined
	6. check if a event coincides with another event
	7. accept/decline an event request
	8. get participants for an event
	9. ping all participants of an event

	apis for social
	1. get

	apis for users
	1.

	apis for chat
	1.

*/

// incoming handlers for social+calendar
/*

func handleIncomingCalendarRequest(event ws.Event, client *ws.Client) error {
	var incomingCalendarRequest models.CalendarRequestEvent
	if err := json.Unmarshal(event.Payload, &incomingCalendarRequest); err != nil {
		log.Printf("incorrect payload of incoming calendar request %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the calendar request",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	// TODO : do validation(not already a active request exists or already part of event) [change in DB that accepted requests move to
	//different table CalendarEventMembers, (same for friend/follow)] and then update in Db -> if error handle it

	ackData := ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}
	client.SendAckToClient(ackData, event.Id)
	return nil
}

func handleIncomingSocialRequestStatusChange(event ws.Event, client *ws.Client) error {
	var incomingSocialRequestUpdate ws.IncomingSocialRequestUpdate
	if err := json.Unmarshal(event.Payload, &incomingSocialRequestUpdate); err != nil {
		log.Printf("incorrect payload of incoming friend request accept %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the friend request accept",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	// perform DB update

	_, err := database.GetSocialQueries().UpdateSocialRequestStatusChange(context.Background(), incomingSocialRequestUpdate.RequestId, incomingSocialRequestUpdate.UpdatedBy, incomingSocialRequestUpdate.ModifiedState)

	if err != nil {
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed updating social request",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)
		return err
	}

	ackData := ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}
	client.SendAckToClient(ackData, event.Id)

	return nil
}

func handleIncomingCalendarRequestStatusChange(event ws.Event, client *ws.Client) error {
	var incomingCalendarRequestAccept models.CalendarRequestAcceptEvent
	if err := json.Unmarshal(event.Payload, &incomingCalendarRequestAccept); err != nil {
		log.Printf("incorrect payload of incoming calendar request accept %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the calendar request accept",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	// db update

	ackData := ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}
	client.SendAckToClient(ackData, event.Id)
	return nil
}

func handleIncomingUserStatusChange(event ws.Event, client *ws.Client) error {
	var incomingUserStatusChange ws.IncomingUserStatusChange
	if err := json.Unmarshal(event.Payload, &incomingUserStatusChange); err != nil {
		log.Printf("incorrect payload of incoming calendar request accept %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the calendar request accept",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}
	ackData := ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}

	// send to the required client(s) if online and also Write to DB
	client.SendAckToClient(ackData, event.Id)
	return nil
}

func handleIncomingSocialRequest(event ws.Event, client *ws.Client) error {
	var incomingSocialRequest ws.IncomingSocialRequest
	if err := json.Unmarshal(event.Payload, &incomingSocialRequest); err != nil {
		log.Printf("incorrect payload of incoming friend request %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed parsing the friend request",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	_, err := database.GetSocialQueries().CreateNewSocialRequest(context.Background(), incomingSocialRequest.RequestorId, incomingSocialRequest.TargetUserId, incomingSocialRequest.Description, incomingSocialRequest.RequestType)

	if err != nil {
		log.Printf("failed creating new request : err - %v", err)
		ackData := ws.Acknowledge{
			EventType: event.Type,
			Status:    false,
			Message:   "Failed creating the social request",
			AckTime:   time.Now().UnixNano(),
		}
		client.SendAckToClient(ackData, event.Id)

		return err
	}

	ackData := ws.Acknowledge{
		EventType: event.Type,
		Status:    true,
		Message:   "Success",
		AckTime:   time.Now().UnixNano(),
	}

	client.SendAckToClient(ackData, event.Id)

	return nil
}

*/

package main
