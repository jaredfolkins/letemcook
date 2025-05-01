package models

type FormRegister struct {
	Username string `db:"username" validate:"required,min=3,max=32"`
	Email    string `db:"email" validate:"required,email"`
	Password string `db:"-" validate:"required,min=8,max=64"`
}

type FormLogin struct {
	Username string `db:"username" validate:"required,min=3,max=32"`
	Password string `db:"-" validate:"required,min=8,max=64"`
}
