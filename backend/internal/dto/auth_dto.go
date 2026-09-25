package dto

// LoginRequest is the login payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=2,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// LoginResponse carries the JWT token and the user.
type LoginResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}
