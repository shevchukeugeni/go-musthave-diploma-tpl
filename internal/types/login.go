package types

type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *UserLoginRequest) User() *User {
	return &User{
		Login:    req.Login,
		Password: req.Password,
	}
}
