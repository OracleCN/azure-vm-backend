package v1

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}
type LoginResponseData struct {
	AccessToken string `json:"accessToken"`
}
type LoginResponse struct {
	Response
	Data LoginResponseData
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname" example:"alan"`
	Email    string `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
	Avatar   string `json:"avatar" example:"https://www.baidu.com/1.jpg"`
}
type GetProfileResponseData struct {
	UserId   string   `json:"userId"`
	Nickname string   `json:"nickname" example:"alan"`
	Roles    []string `json:"roles" example:"admin"` // 用户角色 ["admin"]
	Email    string   `json:"email"`
	Avatar   string   `json:"avatar"`
}
type GetProfileResponse struct {
	Response
	Data GetProfileResponseData
}
