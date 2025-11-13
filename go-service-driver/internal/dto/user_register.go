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
