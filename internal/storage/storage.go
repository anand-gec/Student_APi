package storage

import "github.com/anand-gec/student-api/internal/types"

// import (
// 	// _ "modernc.org/sqlite"
// )

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	UpdateStudent(id int64, name string, email string, age int) error
	GetStatusById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
	DeleteStudentById(id int64) error
	DeleteAllStudents() (int64, error)

	//Admin
	CreateAdmin(email, plainPassword string) (int64, error)
	GetAdminByEmail(email string) (types.Admin, error)
	StoreResetToken(email string) (string, error)
	ResetPasswordWithToken(token, newPlainPassword string) error
	
}
