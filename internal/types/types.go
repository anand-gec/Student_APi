package types

type Student struct {
	Id    int    `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age" validate:"required"`
}

type Admin struct{
	Id int64 `json:"id"`
	Email string `json:"email"`
	Password string `json:"-"`
}
