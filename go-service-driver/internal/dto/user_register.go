package dto

type SendVerificationEmail struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterUser struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Code        string `json:"code" binding:"required,len=6"`
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name"`
}

type UserResponse struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type RegisterResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
	Token   string       `json:"token"`
}
