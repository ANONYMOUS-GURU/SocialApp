package controllers

import (
	"g_chat/database"
	"g_chat/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. make a social request
// 2. update a social request
// 3. unfriend/unfollow
// 4. get all friends + count
// 5. get all followers + count
// 6. get all follows + count
// 7. is mutual friend
// 8. is friend/following/follow
// 9. get all active requests
// 10.delete a request

func CreateNewSocialRequest(ctx *gin.Context) {
	var newSocialRequest models.NewSocialRequestEvent
	if err := ctx.Bind(&newSocialRequest); err != nil {
		return
	}

	// userId check
	if newSocialRequest.RequestorId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	// check not connected already
	if newSocialRequest.RequestType == "FRIEND" {
		val, err := database.GetSocialQueries().FAreUserFriends(ctx.Request.Context(), newSocialRequest.RequestorId, newSocialRequest.TargetUserId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already connected",
			})
			return
		}
	} else if newSocialRequest.RequestType == "FOLLOW" {
		val, err := database.GetSocialQueries().FUserFollowsAnotherUser(ctx.Request.Context(), newSocialRequest.RequestorId, newSocialRequest.TargetUserId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already connected",
			})
			return
		}
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect param",
		})
		return
	}

	// check if an active request already exists with the given type
	val, err := database.GetSocialQueries().FIsActiveRequestOfGivenTypeExists(ctx.Request.Context(), newSocialRequest.RequestorId, newSocialRequest.TargetUserId, newSocialRequest.RequestType)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}
	if val {
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "already connected",
		})
		return
	}

	createdRequest, err := database.GetSocialQueries().CreateNewSocialRequest(ctx.Request.Context(), newSocialRequest.RequestorId, newSocialRequest.TargetUserId, newSocialRequest.Message, newSocialRequest.RequestType)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, createdRequest)

}

// TODO - create trigger for adding to friend/follow on ACCEPT
func UpdateSocialRequest(ctx *gin.Context) {
	var socialRequestUpdateEvent models.SocialRequestUpdateEvent
	if err := ctx.Bind(&socialRequestUpdateEvent); err != nil {
		return
	}

	// userId check
	if socialRequestUpdateEvent.UpdaterId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	// check not already connected
	if socialRequestUpdateEvent.RequestType == "FRIEND" {
		val, err := database.GetSocialQueries().FAreUserFriends(ctx.Request.Context(), socialRequestUpdateEvent.UpdaterId, socialRequestUpdateEvent.RequestorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already connected",
			})
			return
		}
	} else if socialRequestUpdateEvent.RequestType == "FOLLOW" {
		val, err := database.GetSocialQueries().FUserFollowsAnotherUser(ctx.Request.Context(), socialRequestUpdateEvent.UpdaterId, socialRequestUpdateEvent.RequestorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already connected",
			})
			return
		}
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect param",
		})
		return
	}

	// update
	_, err := database.GetSocialQueries().UpdateSocialRequestStatusChange(ctx.Request.Context(), socialRequestUpdateEvent.ID, socialRequestUpdateEvent.UpdateType)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}

func UnfriendOrUnfollowUsers(ctx *gin.Context) {
	var disconnectSocialUser models.DisconnectSocialUsers
	if err := ctx.Bind(&disconnectSocialUser); err != nil {
		return
	}

	// userId check
	if disconnectSocialUser.User1Id != ctx.Keys["userId"] && disconnectSocialUser.User2Id != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	// check not already connected
	if disconnectSocialUser.ConnectionType == "FRIEND" {
		val, err := database.GetSocialQueries().FAreUserFriends(ctx.Request.Context(), disconnectSocialUser.User1Id, disconnectSocialUser.User2Id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if !val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already not connected",
			})
			return
		}
	} else if disconnectSocialUser.ConnectionType == "FOLLOW" {
		val, err := database.GetSocialQueries().FUserFollowsAnotherUser(ctx.Request.Context(), disconnectSocialUser.User1Id, disconnectSocialUser.User2Id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "DB error",
			})
			return
		}
		if !val {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "already not connected",
			})
			return
		}
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "incorrect param",
		})
		return
	}

	err := database.GetSocialQueries().UnfriendOrUnfollowUsers(ctx.Request.Context(), disconnectSocialUser.User1Id, disconnectSocialUser.User2Id, disconnectSocialUser.ConnectionType)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func GetAllFriends(ctx *gin.Context) {
	var getAllSocialConnections models.GetAllSocialConnections
	if err := ctx.Bind(&getAllSocialConnections); err != nil {
		return
	}

	if getAllSocialConnections.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if getAllSocialConnections.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid row count param",
		})
		return
	}

	friends, err := database.GetSocialQueries().GetFriends(ctx.Request.Context(), getAllSocialConnections.UserId, getAllSocialConnections.Time, getAllSocialConnections.RowCount)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, friends)

}

func GetAllFollowers(ctx *gin.Context) {
	var getAllSocialConnections models.GetAllSocialConnections
	if err := ctx.Bind(&getAllSocialConnections); err != nil {
		return
	}

	if getAllSocialConnections.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if getAllSocialConnections.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid row count param",
		})
		return
	}

	followers, err := database.GetSocialQueries().GetFollowers(ctx.Request.Context(), getAllSocialConnections.UserId, getAllSocialConnections.Time, getAllSocialConnections.RowCount)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, followers)

}

func GetAllFollows(ctx *gin.Context) {
	var getAllSocialConnections models.GetAllSocialConnections
	if err := ctx.Bind(&getAllSocialConnections); err != nil {
		return
	}

	if getAllSocialConnections.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if getAllSocialConnections.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid row count param",
		})
		return
	}

	follows, err := database.GetSocialQueries().GetFollows(ctx.Request.Context(), getAllSocialConnections.UserId, getAllSocialConnections.Time, getAllSocialConnections.RowCount)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, follows)

}

func GetFriendsCount(ctx *gin.Context) {
	var userId string
	if err := ctx.Bind(&userId); err != nil {
		return
	}

	count, err := database.GetSocialQueries().GetFriendsCountForUser(ctx.Request.Context(), userId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"count":   count,
	})
}

func GetFollowsCount(ctx *gin.Context) {
	var userId string
	if err := ctx.Bind(&userId); err != nil {
		return
	}

	count, err := database.GetSocialQueries().GetFollowsCountForUser(ctx.Request.Context(), userId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"count":   count,
	})
}

func GetFollowersCount(ctx *gin.Context) {
	var userId string
	if err := ctx.Bind(&userId); err != nil {
		return
	}

	count, err := database.GetSocialQueries().GetFollowersCountForUser(ctx.Request.Context(), userId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
		"count":   count,
	})
}

func CheckMutualFriend(ctx *gin.Context) {
	var pairOfUsers models.PairUsers
	ctx.Bind(&pairOfUsers)

	if pairOfUsers.User1Id != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetSocialQueries().FAreUsersMutualFriends(ctx.Request.Context(), pairOfUsers.User1Id, pairOfUsers.User2Id)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "success",
		"isMutual": val,
	})
}

func FUserAreFriends(ctx *gin.Context) {
	var pairOfUsers models.PairUsers
	ctx.Bind(&pairOfUsers)

	if pairOfUsers.User1Id != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetSocialQueries().FAreUserFriends(ctx.Request.Context(), pairOfUsers.User1Id, pairOfUsers.User2Id)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "success",
		"isMutual": val,
	})
}

func FUserFollowsOtherUser(ctx *gin.Context) {
	var pairOfUsers models.PairUsers
	ctx.Bind(&pairOfUsers)

	if pairOfUsers.User1Id != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	val, err := database.GetSocialQueries().FUserFollowsAnotherUser(ctx.Request.Context(), pairOfUsers.User1Id, pairOfUsers.User2Id)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "success",
		"isMutual": val,
	})
}

func GetAllActiveSocialRequests(ctx *gin.Context) {
	var activeSocialRequests models.GetSocialRequests
	if err := ctx.Bind(&activeSocialRequests); err != nil {
		return
	}

	if activeSocialRequests.UserId != ctx.Keys["userId"] {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	if activeSocialRequests.RowCount <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid row count param",
		})
		return
	}

	allRequests, err := database.GetSocialQueries().GetAllActiveSocialRequests(ctx.Request.Context(), activeSocialRequests.UserId, activeSocialRequests.Time, activeSocialRequests.RowCount)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "DB error",
		})
		return
	}

	ctx.JSON(http.StatusOK, allRequests)

}
