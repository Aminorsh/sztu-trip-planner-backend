package dto

type UserProfileResponse struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastLogin   string `json:"last_login"`
}

type UploadAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

type UpdateUserProfile struct {
	DisplayName *string `json:"display_name,omitempty"`
	// AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio *string `json:"bio,omitempty"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
}
