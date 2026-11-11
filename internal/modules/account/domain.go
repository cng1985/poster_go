package account

import "time"

type AccountRegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type AccountLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccountUpdateRequest struct {
	DisplayName string `json:"display_name"`
}

type AccountRequest struct {
	ID int64 `json:"id,string"`
}

type AccountResponse struct {
	ID          int64     `json:"id,string"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type AccountSessionResponse struct {
	Token   string          `json:"token"`
	Account AccountResponse `json:"account"`
}
