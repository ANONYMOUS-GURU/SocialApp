CREATE TABLE IF NOT EXISTS Users (
    id VARCHAR(255) PRIMARY KEY,
    -- Assuming ID is a string
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    deleted_at BIGINT
);

CREATE TABLE IF NOT EXISTS Conversations (
    id VARCHAR(255) PRIMARY KEY,
    is_group BOOLEAN NOT NULL,
    owner_id VARCHAR(255),
    name VARCHAR(255),
    description TEXT,
    image_url VARCHAR(255),
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    deleted_at BIGINT,
    last_message_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS Messages (
    id VARCHAR(255) PRIMARY KEY,
    body TEXT NOT NULL,
    conversation_id VARCHAR(255) NOT NULL,
    sender_id VARCHAR(255) NOT NULL,
    delivered_count INT NOT NULL DEFAULT 1,
    seen_count INT NOT NULL,
    sent_to_count INT NOT NULL,
    sent_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES Conversations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS MessageUserMap (
    message_id VARCHAR(255) NOT NULL,
    receiver_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (message_id, receiver_id),
    -- Composite primary key
    FOREIGN KEY (message_id) REFERENCES Messages(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ConversationParticipants (
    conversation_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    last_message_seen_at BIGINT NOT NULL,
    joined_at BIGINT NOT NULL,
    deleted_at BIGINT,
    PRIMARY KEY (conversation_id, user_id),
    -- Composite primary key
    FOREIGN KEY (conversation_id) REFERENCES Conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE
);

-- every new friend/follow request will come here and will be marked as active. If accepted it will be 
-- marked as accepted and the record will be added in a new Friends/Follows table. If 
-- declined its status here will be changed and kept as is that is not moved to a separate table.
-- If calendar event owner decides to remove someone it will be removed from friends/follow column, 
-- NO CHANGE IN REQUESTS TABLE. If someone requests for friend/follow, it will check if an 
-- active request exists and if the person is not already a friend/follow only then 
-- you can add the request. Meaning declined and accepted are of no importance for creating new 
-- request. Only check if user is already a friend/follow OR an active request exists.

CREATE TABLE IF NOT EXISTS Friends (
  user1_id VARCHAR(255) NOT NULL,
  user2_id VARCHAR(255) NOT NULL,
  created_at BIGINT NOT NULL,
  PRIMARY KEY (user1_id, user2_id),  -- Bi-directional friend relationship
  FOREIGN KEY (user1_id) REFERENCES Users(id) ON DELETE CASCADE,
  FOREIGN KEY (user2_id) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Follows (
  follower_id VARCHAR(255) NOT NULL,
  followed_id VARCHAR(255) NOT NULL,
  created_at BIGINT NOT NULL,
  PRIMARY KEY (follower_id, followed_id),  -- Uni-directional follow relationship
  FOREIGN KEY (follower_id) REFERENCES Users(id) ON DELETE CASCADE,
  FOREIGN KEY (followed_id) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS SocialRequests (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  target_user_id VARCHAR(255) NOT NULL,
  request_type VARCHAR(20) NOT NULL NOT NULL,  -- Enum for request type
  request_message TEXT NOT NULL,  -- Optional message for the request
  request_status VARCHAR(20) NOT NULL  DEFAULT 'ACTIVE',  -- Default to active
  created_at BIGINT NOT NULL,
  updated_at BIGINT,
  FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE,
  FOREIGN KEY (target_user_id) REFERENCES Users(id) ON DELETE CASCADE
);


-- every new calendar request will come here and will be marked as active. If accepted it will be 
-- marked as accepted and the record will be added in a new CalendarEventParticipants table. If 
-- declined its status here will be changed and kept as is that is not moved to a separate table.
-- If calendar event owner decides to remove someone it will be removed from participants column, 
-- NO CHANGE IN REQUESTS TABLE. If someone requests for join in calendar events, it will check if an 
-- active request exists and if the person is not already a participant in that event only then 
-- you can add the request. Meaning declined and accepted are of no importance for creating new 
-- request. Only check if user is part of event OR an active request exists. 

CREATE TABLE IF NOT EXISTS CalendarEvents (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  event_title TEXT NOT NULL,
  event_description TEXT NOT NULL,
  from_time BIGINT NOT NULL,
  to_time BIGINT NOT NULL,
  is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
  game_id VARCHAR(255) NOT NULL,
  created_at BIGINT NOT NULL,
  updated_at BIGINT,
  deleted_at BIGINT,
  FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CalendarEventRequests (
  id VARCHAR(255) PRIMARY KEY,
  event_id VARCHAR(255) REFERENCES CalendarEvents(id) NOT NULL,
  requesting_user_id VARCHAR(255) NOT NULL,
  request_message TEXT NOT NULL,
  request_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',  -- Default to active
  created_at BIGINT NOT NULL,
  updated_at BIGINT,
  FOREIGN KEY (requesting_user_id) REFERENCES Users(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES CalendarEvents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CalendarEventParticipants (
  event_id VARCHAR(255) NOT NULL REFERENCES CalendarEvents(id) ON DELETE CASCADE,
  user_id VARCHAR(255) NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
  joined_at BIGINT NOT NULL,  -- Timestamp when the user joined the event
  is_organizer BOOLEAN NOT NULL DEFAULT FALSE,  -- Flag to indicate if the user is the event organizer
  -- status ENUM ('ATTENDING', 'MAYBE', 'NOT_ATTENDING') DEFAULT 'ATTENDING',  -- Participant's attendance status
  PRIMARY KEY (event_id, user_id)  -- Composite primary key
);

CREATE TABLE IF NOT EXISTS CalendarEventInvites (
  id VARCHAR(255) PRIMARY KEY,
  event_id VARCHAR(255) REFERENCES CalendarEvents(id) NOT NULL,
  invited_user_id VARCHAR(255) NOT NULL,
  invite_message TEXT NOT NULL,
  invite_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',  -- Default to active
  created_at BIGINT NOT NULL,
  updated_at BIGINT,
  FOREIGN KEY (invited_user_id) REFERENCES Users(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES CalendarEvents(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION notify_chat()
RETURNS TRIGGER AS $$
BEGIN
  -- Send notification with row data as payload
  PERFORM pg_notify('chat_written', to_json(new)::text);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER notify_chat_trigger
AFTER INSERT ON MessageUserMap
FOR EACH ROW
EXECUTE PROCEDURE notify_chat();


CREATE OR REPLACE FUNCTION notify_chat_received()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify('chat_received', to_json(new)::text);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER notify_chat_received_trigger
AFTER UPDATE ON Messages
FOR EACH ROW
WHEN (NEW.delivered_count = NEW.sent_to_count and OLD.delivered_count != NEW.delivered_count)
EXECUTE PROCEDURE notify_chat_received();

CREATE OR REPLACE FUNCTION notify_chat_read()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify('chat_read', to_json(new)::text);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER notify_chat_read_trigger
AFTER Update ON Messages
FOR EACH ROW
WHEN (NEW.seen_count = NEW.sent_to_count and OLD.seen_count != NEW.seen_count)
EXECUTE PROCEDURE notify_chat_read();
