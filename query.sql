-- name: getUserByID :one
SELECT * FROM Users WHERE id = $1;


-- name: getMessageByID :one
SELECT * FROM Messages WHERE id = $1;


-- name: getConversationByID :one
SELECT * FROM Conversations WHERE id = $1;


-- name: getAllUsersInConversation :many
SELECT * FROM ConversationParticipants WHERE conversation_id = $1;


-- name: getMostRecentConversationsForUser :many
SELECT c.*
FROM Conversations c
INNER JOIN ConversationParticipants cp ON c.id = cp.conversation_id
WHERE cp.user_id = $1 AND c.last_message_at < $2
ORDER BY c.last_message_at DESC
LIMIT $3;

-- name: getMostRecentMessagesForUser :many
SELECT * FROM Messages
WHERE conversation_id IN (
  SELECT conversation_id
  FROM ConversationParticipants
  WHERE user_id = $1
) AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;


-- name: getMostRecentMessagesForUserInConversation :many
SELECT * FROM Messages
WHERE conversation_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;


-- name: createMessage :one
INSERT INTO Messages (id, body, conversation_id, sender_id, delivered_count, seen_count, sent_to_count, sent_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;


-- name: createConversation :one
INSERT INTO Conversations (id, is_group, owner_id, name, description, image_url, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;


-- name: createConversationParticipant :exec
INSERT INTO ConversationParticipants (conversation_id, user_id, is_owner, joined_at, last_message_seen_at)
VALUES ($1, $2, $3, $4, $5);


-- name: createMessageUserMap :exec
INSERT INTO MessageUserMap (message_id, receiver_id)
VALUES ($1, $2);


-- name: createUser :exec
INSERT INTO Users (id, name, username, email, description, image_url, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);


-- name: updateUser :exec
UPDATE Users
SET name = $2, description = $3, image_url = $4
WHERE id = $1;

-- name: fMessageIsInUsersConversation :one
SELECT EXISTS (
  SELECT 1
  FROM Conversations c
  INNER JOIN Messages m ON c.id = m.conversation_id
  INNER JOIN ConversationParticipants cp ON c.id = cp.conversation_id
  WHERE m.id = $1 AND cp.user_id = $2
);


-- name: getLastMessageSeenTimeByUserInConversation :one
SELECT last_message_seen_at FROM ConversationParticipants
WHERE user_id = $1 AND conversation_id = $2;


-- name: isUserPartOfConversation :one
SELECT EXISTS (
  SELECT 1
  FROM ConversationParticipants cp
  WHERE cp.conversation_id = $1 AND cp.user_id = $2
);


-- name: getAllMessagesForConversation :many
WITH ranked_messages AS (
  SELECT m.*
  FROM Messages m
  WHERE m.conversation_id = $1 AND m.created_at > $2
  ORDER BY m.created_at ASC
)
SELECT *
FROM ranked_messages
LIMIT $3;

-- name: updateLastMessageAtInConversation :exec
UPDATE Conversations
SET last_message_at = (
  SELECT MAX(created_at)
  FROM Messages
  WHERE conversation_id = $1
)
WHERE id = $1;

-- name: updateLastMessageSeenAtInConversationParticipant :exec
UPDATE ConversationParticipants
SET last_message_seen_at = $3
WHERE conversation_id = $1 AND user_id = $2;


-- name: getAllMessagesAfterGivenTime :many
WITH ranked_messages AS (
  SELECT m.*
  FROM Messages m
  INNER JOIN ConversationParticipants cp ON m.conversation_id = cp.conversation_id
  WHERE cp.user_id = $1 AND m.created_at > $2
  ORDER BY m.created_at ASC
)
SELECT *
FROM ranked_messages
LIMIT $3;


-- name: markMessageAsReceivedByUser :many
DELETE FROM MessageUserMap
USING (
  SELECT * FROM UNNEST($1::VARCHAR[]) AS id
) AS message_ids
WHERE message_id = message_ids.id
  AND receiver_id = $2 RETURNING *;

-- name: updateSeenCountForMessages :exec
UPDATE Messages
SET seen_count = seen_count + 1
WHERE id IN (SELECT UNNEST($1));


-- name: updateDeliveredCountForMessages :exec
UPDATE Messages
SET delivered_count = delivered_count + 1
WHERE id IN (SELECT UNNEST($1));


-- name: updateSeenCountForMessagesBetweenTimestamps :exec
UPDATE Messages
SET seen_count = seen_count + 1
WHERE EXISTS (
  SELECT 1
  FROM Messages AS m2
  WHERE m2.id = Messages.id
    AND m2.created_at > $1 AND m2.created_at <= $2
);

-- name: getFriendsOfUser :many
SELECT u1.id, u1.name, u1.image_url, f.created_at
FROM Friends f
INNER JOIN Users u1 ON f.user1_id = u1.id
INNER JOIN Users u2 ON f.user2_id = u2.id
WHERE f.user1_id = $1 AND f.created_at > $2
ORDER BY f.created_at ASC
LIMIT $3;

-- name: getFollowersOfUser :many
SELECT u.id, u.name, u.image_url, f.created_at
FROM Follows f
INNER JOIN Users u ON f.follower_id = u.id
WHERE f.followed_id = $1 AND f.created_at > $2
ORDER BY f.created_at ASC
LIMIT $3;

-- name: getUsersFollowedByUser :many
SELECT u.id, u.name, u.image_url, f.created_at
FROM Follows f
INNER JOIN Users u ON f.followed_id = u.id
WHERE f.follower_id = $1 AND f.created_at > $2
ORDER BY f.created_at ASC
LIMIT $3;

-- name: getNumberOfUserFriends :one
SELECT COUNT(*)
FROM Friends
WHERE user1_id = $1 OR user2_id = $1;

-- name: getNumberOfUsersFollowedByUser :one
SELECT COUNT(*)
FROM Follows
WHERE follower_id = $1;

-- name: getNumberOfFollowersOfUser :one
SELECT COUNT(*)
FROM Follows
WHERE followed_id = $1;

-- name: fAreUsersFriends :one
SELECT EXISTS(
  SELECT 1
  FROM Friends
  WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
);

-- name: fUserFollowsAnotherUser :one
SELECT EXISTS(
  SELECT 1
  FROM Follows
  WHERE follower_id = $1 AND followed_id = $2
);

-- name: getAllPendingSocialRequestForUser :many
SELECT *
FROM SocialRequests sr
WHERE sr.user_id = $1
  AND sr.request_status = 'ACTIVE'
  AND sr.created_at > $2
ORDER BY sr.created_at ASC
LIMIT $3;


-- name: updateSocialRequest :one
UPDATE SocialRequests
SET request_status = $4, updated_at = $5
WHERE target_user_id = $2 AND user_id = $1 AND request_type = $3 AND request_status = 'ACTIVE'
RETURNING *;

-- name: updateSocialRequestById :one
UPDATE SocialRequests
SET request_status = $2,
  updated_at = $3
WHERE id = $1
RETURNING *;

-- name: fIsSocialRequestActive :one
SELECT EXISTS(
  SELECT 1
  FROM SocialRequests sr
  WHERE sr.user_id = $1
    AND sr.request_status = 'ACTIVE'
    AND sr.target_user_id = $2
    AND sr.request_type = $3
);

-- name: makeUsersFollow :one
INSERT INTO Follows (follower_id, followed_id, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: makeUsersFriends :one
INSERT INTO Friends (user1_id, user2_id, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: makeUserUnfollow :exec
DELETE FROM Follows
WHERE follower_id = $1 AND followed_id = $2;

-- name: unfriendUsers :exec
DELETE FROM Friends
WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1);

-- name: createNewSocialRequest :one
INSERT INTO SocialRequests (id, user_id, target_user_id, request_type, request_message, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: activeSocialRequestExists :one
SELECT EXISTS(
  SELECT 1
  FROM SocialRequests sr
  WHERE sr.user_id = $1
    AND sr.target_user_id = $2
    AND sr.request_type = $3
    AND sr.request_status == 'ACTIVE'
);

-- name: getMutualFriendsOfUser :many
WITH mutual_friends AS (
  SELECT f1.user2_id AS friend_id
  FROM Friends f1
  INNER JOIN Friends f2 ON f1.user1_id = f2.user2_id
  WHERE f1.user1_id = $1
  AND f2.user1_id = $2
  UNION ALL
  SELECT f1.user1_id AS friend_id
  FROM Friends f1
  INNER JOIN Friends f2 ON f1.user2_id = f2.user1_id
  WHERE f1.user1_id = $1
  AND f2.user1_id = $2
)
SELECT u.id, u.name
FROM mutual_friends mf
INNER JOIN Users u ON mf.friend_id = u.id
ORDER BY u.name ASC
LIMIT $3 OFFSET $4;


-- name: createNewCalendarEvent :one
INSERT INTO CalendarEvents (id, user_id, event_title, event_description, from_time, to_time, is_recurring, game_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: getAllCalendarRequestForEvent :many
SELECT cer.*
FROM CalendarEventRequests cer
WHERE cer.event_id = $1
  AND cer.request_status = $2
  AND cer.created_at > $3
ORDER BY cer.created_at ASC
LIMIT $4;

-- name: fIsUserOrganizerOfEvent :one
SELECT EXISTS (
  SELECT 1
  FROM CalendarEvents
  WHERE id = $1 AND user_id = $2
);

-- name: updateCalendarRequest :one
UPDATE CalendarEventRequests
SET request_status = $3, updated_at = $2
WHERE id = $1 AND request_status = 'ACTIVE'
RETURNING *;

-- name: getAllCalendarEventParticipants :many
SELECT cep.*, u.name, u.image_url
FROM CalendarEventParticipants cep
INNER JOIN Users u ON cep.user_id = u.id
WHERE cep.event_id = $1
  AND cep.joined_at > $2
ORDER BY cep.joined_at ASC
LIMIT $3;

-- name: getScheduledEventsForUser :many
SELECT ce.*
FROM CalendarEvents ce
INNER JOIN CalendarEventParticipants cep ON ce.id = cep.event_id
WHERE cep.user_id = $1
  AND ce.from_time >= $2
  AND ce.deleted_at IS NULL
ORDER BY cep.joined_at ASC
LIMIT $3;


-- name: getScheduledEventsCreatedByUser :many
SELECT ce.*
FROM CalendarEvents ce
INNER JOIN CalendarEventParticipants cep ON ce.id = cep.event_id
WHERE cep.user_id = $1
  AND ce.from_time >= $2
  AND ce.user_id = $4
  AND ce.deleted_at IS NULL
ORDER BY cep.joined_at ASC
LIMIT $3;

-- name: userRequestToLeaveEvent :one
DELETE FROM CalendarEventParticipants
WHERE event_id = $1 AND user_id = $2
RETURNING *;

-- name: userRequestToJoinEvent :one
INSERT INTO CalendarEventRequests (id, event_id, requesting_user_id, request_message, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: organizerRequestToDeleteEvent :one
DELETE FROM CalendarEvents
WHERE id = $1
RETURNING *;

-- name: createNewCalendarEventRequest :one
INSERT INTO CalendarEventRequests (
  id,
  event_id,
  requesting_user_id,
  request_message,
  request_status,
  created_at
)
VALUES (
  $1, $2, $3, $4, "ACTIVE", $5
) RETURNING *;

-- name: isUserRequestOnEventExists :one
SELECT EXISTS (
  SELECT 1
  FROM CalendarEventRequests
  WHERE requesting_user_id = $1
  AND event_id = $2
  AND request_status = 'ACTIVE'
);

-- name: fIsUserAlreadyAParticipantOfCalendarEvent :one
SELECT EXISTS (
  SELECT 1
  FROM CalendarEventParticipants
  WHERE user_id = $1
  AND event_id = $2
);

-- name: updateCalendarEventDetails :one
UPDATE CalendarEvents
SET 
  event_title = $2, 
  event_description = $3,
  from_time = $4,
  to_time = $5,
  game_id = $7,
  updated_at = $6
WHERE id = $1 RETURNING *;

-- name: getCalendarEventForId :one
SELECT * FROM CalendarEvents
WHERE id = $1;

-- name: createNewCalendarInvite :one
INSERT INTO CalendarEventInvites (
  id,
  event_id,
  invited_user_id,
  invite_message,
  invite_status,
  created_at
)
VALUES (
  $1, -- invite ID (string)
  $2, -- event ID (string)
  $3, -- invited user ID (string)
  $4, -- invite message (text)
  $5,
  $6 -- timestamp of creation
) RETURNING *;


-- name: deleteCalendarInvite :exec
DELETE FROM CalendarEventInvites
WHERE id = $1;


-- name: getAllCalendarInvitesForUser :many
SELECT *
FROM CalendarEventInvites
WHERE invited_user_id = $1
  AND created_at < $2
  AND invite_status = $4
ORDER BY created_at DESC
LIMIT $3;


-- name: getAllCalendarInvitesForEvent :many
SELECT *
FROM CalendarEventInvites
WHERE event_id = $1
  AND created_at < $2
  AND invite_status = $4
LIMIT $3;
