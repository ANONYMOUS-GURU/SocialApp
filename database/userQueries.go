package database

import (
	"context"
	"g_chat/models"
	"log"
	"time"
)

type UserQueries struct {
	*Queries
}

func GetUserQueries() *UserQueries {
	queries := getQueries()
	return &UserQueries{queries}
}

// Creates a new User with the New User type
func (db *UserQueries) CreateUser(ctx context.Context, newUser *models.NewUser) error {
	if err := db.Queries.createUser(ctx, createUserParams{
		ID:          newUser.ID,
		Name:        newUser.Name,
		Username:    newUser.Username,
		Email:       newUser.Email,
		Description: newUser.Description,
		ImageUrl:    newUser.ImageUrl,
		CreatedAt:   time.Now().Unix(),
	}); err != nil {
		log.Printf("error creating new user %v", err)
		return err
	}
	return nil
}

// Updates a User with the update Type
func (db *UserQueries) UpdateUser(ctx context.Context, updatedUser models.UserUpdates, userId string) error {
	if err := db.Queries.updateUser(ctx, updateUserParams{
		ID:          userId,
		Description: updatedUser.Description,
		ImageUrl:    updatedUser.ImageUrl,
		Name:        updatedUser.Name,
	}); err != nil {
		log.Print("error updating user")
		return err
	}
	return nil
}

// Delete a User (soft delete)
func (db *UserQueries) DeleteUser(ctx context.Context, userId string) error {
	return nil
}

// Get user for a given Id
func (db *UserQueries) GetUserFromId(ctx context.Context, userId string) (User, error) {
	user, err := db.Queries.getUserByID(ctx, userId)
	if err != nil {
		log.Print("error getting user")
		return User{}, err
	}
	return user, nil
}

// Get all users
func (db *UserQueries) GetAllUsers(ctx context.Context) ([]User, error) {

	return nil, nil
}
