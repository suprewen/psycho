package oauth

// SignupForm to be posted
type SignupForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// SigninForm to be exported
type SigninForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authData struct {
	UserType int `json:"userType"`
}
