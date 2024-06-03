package models

type UserUpdates struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
}

type NewUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
	Username    string `json:"username"`
}
