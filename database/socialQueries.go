package database

import (
	"context"
	"database/sql"
	"errors"
	"g_chat/models"
	"log"
	"time"

	"github.com/google/uuid"
)

type SocialQueries struct {
	*Queries
}

func GetSocialQueries() *SocialQueries {
	queries := getQueries()
	return &SocialQueries{queries}
}

func (db *SocialQueries) CreateNewSocialRequest(ctx context.Context, userId string, targetUserId string, description string, requestType string) (Socialrequest, error) {
	tx, err := getDatabase().BeginTx(ctx, nil)
	if err != nil {
		return Socialrequest{}, err
	}
	defer tx.Rollback() // Rollback on any error

	qtx := db.Queries.WithTx(tx)

	val := false

	if requestType == "FOLLOW" {
		val, err = qtx.fUserFollowsAnotherUser(ctx, fUserFollowsAnotherUserParams{
			FollowerID: userId,
			FollowedID: targetUserId,
		})

		if err != nil {
			log.Printf("DB error : unable to read follow status : error - %v", err)
			return Socialrequest{}, err
		}

		if !val {
			log.Printf("already friend, invalid request")
			return Socialrequest{}, errors.New("already follow, invalid request")
		}
	} else if requestType == "FRIEND" {
		val, err = qtx.fAreUsersFriends(ctx, fAreUsersFriendsParams{
			User1ID: userId,
			User2ID: targetUserId,
		})

		if err != nil {
			log.Printf("DB error : unable to read friend status : error - %v", err)
			return Socialrequest{}, err
		}

		if !val {
			log.Printf("already friend, invalid request")
			return Socialrequest{}, errors.New("already friend, invalid request")
		}

	} else {
		return Socialrequest{}, errors.New("wrong params : invalid request type : f(CreateNewSocialRequest)")
	}

	val, err = qtx.activeSocialRequestExists(ctx, activeSocialRequestExistsParams{
		UserID:       userId,
		TargetUserID: targetUserId,
		RequestType:  requestType,
	})

	if err != nil {
		log.Printf("DB error : unable to read social status : error - %v", err)
		return Socialrequest{}, err
	}

	if !val {
		log.Printf("already active request exists, invalid request")
		return Socialrequest{}, errors.New("invalid request")
	}

	createdRequest, err := qtx.createNewSocialRequest(ctx, createNewSocialRequestParams{
		ID:             uuid.NewString(),
		UserID:         userId,
		TargetUserID:   targetUserId,
		RequestType:    requestType,
		RequestMessage: description,
		CreatedAt:      time.Now().UnixNano(),
	})

	if err != nil {
		log.Printf("DB error : unable to create request : error - %v", err)
		return Socialrequest{}, err
	}

	return createdRequest, nil
}

func (db *SocialQueries) UpdateSocialRequestStatusChange(ctx context.Context, requestId string, updateType string) (Socialrequest, error) {
	updatedRequest, err := db.Queries.updateSocialRequestById(ctx, updateSocialRequestByIdParams{
		ID:            requestId,
		RequestStatus: updateType,
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: time.Now().UnixNano(),
		},
	})

	if err != nil {
		log.Printf("DB error : error updating value in DB : f(UpdateSocialRequestStatusChange) : error - %v", err)
		return Socialrequest{}, err
	}

	return updatedRequest, nil
}

func (db *SocialQueries) FAreUserFriends(ctx context.Context, userId1 string, userId2 string) (bool, error) {
	val, err := db.Queries.fAreUsersFriends(ctx, fAreUsersFriendsParams{
		User1ID: userId1,
		User2ID: userId2,
	})

	if err != nil {
		return false, err
	}

	return val, nil
}

func (db *SocialQueries) FUserFollowsAnotherUser(ctx context.Context, userId1 string, userId2 string) (bool, error) {
	val, err := db.Queries.fUserFollowsAnotherUser(ctx, fUserFollowsAnotherUserParams{
		FollowerID: userId1,
		FollowedID: userId2,
	})
	if err != nil {
		return false, err
	}

	return val, nil
}

func (db *SocialQueries) FIsActiveRequestOfGivenTypeExists(ctx context.Context, requestorId string, targetUserId string, requestType string) (bool, error) {
	val, err := db.Queries.fIsSocialRequestActive(ctx, fIsSocialRequestActiveParams{
		UserID:       requestorId,
		TargetUserID: targetUserId,
		RequestType:  requestType,
	})
	if err != nil {
		return false, err
	}

	return val, nil
}

func (db *SocialQueries) UnfriendOrUnfollowUsers(ctx context.Context, user1Id string, user2Id string, connectionType string) error {
	if connectionType == "FRIEND" {
		if err := db.Queries.unfriendUsers(ctx, unfriendUsersParams{
			User1ID: user1Id,
			User2ID: user2Id,
		}); err != nil {
			return err
		}

	} else if connectionType == "FOLLOW" {
		if err := db.Queries.makeUserUnfollow(ctx, makeUserUnfollowParams{
			FollowerID: user1Id,
			FollowedID: user2Id,
		}); err != nil {
			return err
		}
	} else {
		return errors.New("unsupported type")
	}
	return nil
}

func (db *SocialQueries) GetFriends(ctx context.Context, userId string, time int64, rowCount uint) ([]models.SocialUser, error) {
	friends, err := db.Queries.getFriendsOfUser(ctx, getFriendsOfUserParams{
		User1ID:   userId,
		CreatedAt: time,
		Limit:     int32(rowCount),
	})

	if err != nil {
		return nil, err
	}

	users := make([]models.SocialUser, len(friends))
	for i, friend := range friends {
		users[i] = models.SocialUser{
			UserId:    friend.ID,
			Name:      friend.Name,
			ImageUrl:  friend.ImageUrl,
			CreatedAt: friend.CreatedAt,
		}
	}

	return users, nil
}

func (db *SocialQueries) GetFollowers(ctx context.Context, userId string, time int64, rowCount uint) ([]models.SocialUser, error) {
	followers, err := db.Queries.getFollowersOfUser(ctx, getFollowersOfUserParams{
		FollowedID: userId,
		CreatedAt:  time,
		Limit:      int32(rowCount),
	})

	if err != nil {
		return nil, err
	}

	users := make([]models.SocialUser, len(followers))
	for i, follower := range followers {
		users[i] = models.SocialUser{
			UserId:    follower.ID,
			Name:      follower.Name,
			ImageUrl:  follower.ImageUrl,
			CreatedAt: follower.CreatedAt,
		}
	}

	return users, nil
}

func (db *SocialQueries) GetFollows(ctx context.Context, userId string, time int64, rowCount uint) ([]models.SocialUser, error) {
	follows, err := db.Queries.getUsersFollowedByUser(ctx, getUsersFollowedByUserParams{
		FollowerID: userId,
		CreatedAt:  time,
		Limit:      int32(rowCount),
	})

	if err != nil {
		return nil, err
	}

	users := make([]models.SocialUser, len(follows))
	for i, follow := range follows {
		users[i] = models.SocialUser{
			UserId:    follow.ID,
			Name:      follow.Name,
			ImageUrl:  follow.ImageUrl,
			CreatedAt: follow.CreatedAt,
		}
	}

	return users, nil
}

func (db *SocialQueries) GetFriendsCountForUser(ctx context.Context, userId string) (int64, error) {
	count, err := db.Queries.getNumberOfUserFriends(ctx, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *SocialQueries) GetFollowsCountForUser(ctx context.Context, userId string) (int64, error) {
	count, err := db.Queries.getNumberOfUsersFollowedByUser(ctx, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *SocialQueries) GetFollowersCountForUser(ctx context.Context, userId string) (int64, error) {
	count, err := db.Queries.getNumberOfFollowersOfUser(ctx, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *SocialQueries) FAreUsersMutualFriends(ctx context.Context, user1Id string, user2Id string) (bool, error) {
	return false, nil
}

func (db *SocialQueries) GetAllActiveSocialRequests(ctx context.Context, userId string, time int64, rowCount uint) ([]Socialrequest, error) {
	requests, err := db.Queries.getAllPendingSocialRequestForUser(ctx, getAllPendingSocialRequestForUserParams{
		UserID:    userId,
		Limit:     int32(rowCount),
		CreatedAt: time,
	})

	if err != nil {
		return nil, err
	}

	return requests, nil
}
