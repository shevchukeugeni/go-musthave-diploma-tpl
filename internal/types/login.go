package types

type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
