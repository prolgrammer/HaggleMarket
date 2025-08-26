package user

type User struct {
	Name     string
	Password string
	Email    string
	Phone    string
	IsStore  bool
}

type UpdateUser struct {
	Name     string
	Password string
	Email    string
}
