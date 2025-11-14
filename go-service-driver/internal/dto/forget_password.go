package dto

type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgetPasswordByUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type ForgetPasswordVerifyRequest struct {
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
	Code        string `json:"code" binding:"required"`
}

type ForgetPasswordResponse struct {
	Message string `json:"message"`
}
