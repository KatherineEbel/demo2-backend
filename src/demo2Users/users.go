package demo2Users

type UserData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUser struct {
	FullName string `json:"fullName"`
	Alias    string `json:"alias"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
