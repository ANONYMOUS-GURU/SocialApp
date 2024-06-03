package controllers

import (
	"g_chat/database"
	"g_chat/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCurrentUserDetails(ctx *gin.Context) {
	user, err := database.GetUserQueries().GetUserFromId(ctx.Request.Context(), ctx.Keys["userId"].(string))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed getting user",
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func GetUserFromId(ctx *gin.Context) {
	user, err := database.GetUserQueries().GetUserFromId(ctx.Request.Context(), ctx.Query("id"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed getting user",
		})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func UpdateUser(ctx *gin.Context) {
	var userContent models.UserUpdates
	if err := ctx.Bind(&userContent); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "incorrect body params",
		})
		return
	}

	if err := database.GetUserQueries().UpdateUser(ctx.Request.Context(), userContent, ctx.Keys["userId"].(string)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed creating user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "updated user",
	})
}

func DeleteUser(ctx *gin.Context) {
	if err := database.GetUserQueries().DeleteUser(ctx.Request.Context(), ctx.Keys["userId"].(string)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "error deleting user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted user",
	})
}

func CreateUser(ctx *gin.Context) {
	var user *models.NewUser
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "incorrect body params",
		})
		return
	}

	if len(user.Username) < 5 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Username not valid",
		})
	}

	if err := database.GetUserQueries().CreateUser(ctx.Request.Context(), user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed creating user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "created user",
	})
}
