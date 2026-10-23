package sqlite

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/anand-gec/student-api/internal/config"
	"github.com/anand-gec/student-api/internal/types"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite" // registers "sqlite"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	// Use "sqlite" as the registered driver name for modernc
	db, err := sql.Open("sqlite", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	// Fixed typos: "EXISTS" and "INTEGER of Student Table "
	Studentquery := `CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		age INTEGER
	);`

	_, err = db.Exec(Studentquery)
	if err != nil {
		return nil, err
	}

	// for Admin table
	Adminquery := `CREATE TABLE IF NOT EXISTS admins(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL
	)`

	_, err = db.Exec(Adminquery)
	if err != nil {
		return nil, err
	}

	// for forget or reset password
	ResetQuery := `CREATE TABLE IF NOT EXISTS password_resets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL
	);`

	_, err = db.Exec(ResetQuery)
	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil
}

// for Admin
func (s *Sqlite) CreateAdmin(email, plainPassword string) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}
	stmt, err := s.Db.Prepare("INSERT INTO admins(email,password) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, string(hashedPassword))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Sqlite) GetAdminByEmail(email string) (types.Admin, error) {
	stmt, err := s.Db.Prepare("SELECT id,email,password FROM admins WHERE email = ? LIMIT 1")
	if err != nil {
		return types.Admin{}, err
	}
	defer stmt.Close()

	var admin types.Admin
	err = stmt.QueryRow(email).Scan(&admin.Id, &admin.Email, &admin.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Admin{}, fmt.Errorf("no admin found with email %s", email)
		}
		return types.Admin{}, fmt.Errorf("query error: %w", err)
	}
	return admin, err

}

// for Student
func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {

	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES(?, ?, ?)") // this is used for protection of sql injection -> VALUES(?, ?, ?)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(name, email, age)
	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int) error {
	result, err := s.Db.Exec("UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?", name, email, age, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Sqlite) GetStatusById(id int64) (types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students WHERE id=?LIMIT 1")
	if err != nil {
		return types.Student{}, err
	}
	defer stmt.Close()

	var student types.Student

	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Student{}, fmt.Errorf("no student found with id %s", fmt.Sprint(id))
		}
		return types.Student{}, fmt.Errorf("Query error: %w", err)
	}
	return student, nil

}

func (s *Sqlite) GetStudents() ([]types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []types.Student

	for rows.Next() {
		var student types.Student
		err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, nil

}

func (s *Sqlite) DeleteStudentById(id int64) error {
	stmt, err := s.Db.Prepare("DELETE FROM  students WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no student found with id: %d", id)
	}
	return nil
}

func (s *Sqlite) DeleteAllStudents() (int64, error) {
	query := "DELETE FROM students"
	res, err := s.Db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all students: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to count affected rows: %w", err)
	}
	return rowsAffected, nil
}

func (s *Sqlite) StoreResetToken(email string) (string, error) {
	_, _ = s.Db.Exec("DELETE FROM password_resets WHERE email = ?", email)

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	expireAt := time.Now().Add(15 * time.Minute)

	stmt, err := s.Db.Prepare("INSERT INTO password_reset(email, token, expires_at) VALUES(?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(email, token, expireAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Sqlite) ResetPasswordWithToken(token, newPlainPassword string) error {
	var email string
	var expireAt time.Time

	// token in sql view
	row := s.Db.QueryRow("SELECT email, expires_at FROM password_resets WHERE token = ? LIMIT 1", token)
	err := row.Scan(&email, &expireAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid query error: %w", err)
		}
		return fmt.Errorf("database query error: %w", err)
	}

	// check token expire
	if time.Now().After(expireAt) {
		_, _ = s.Db.Exec("DELETE FROM password_resets WHERE token = ?", token)
		return fmt.Errorf("reset token has expired")
	}

	//pass decrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPlainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// update the admin
	updateStmt, err := s.Db.Prepare("UPDATE admins SET password = ? WHERE email = ?")
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	res, err := updateStmt.Exec(string(hashedPassword), email)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("admin account not found")
	}

	_, _ = s.Db.Exec("DELETE FROM password_resets WHERE token = ?", token)

	return nil
}
