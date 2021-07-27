package auth

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User define a user struct
type User struct {
	Username       string
	HashedPassword string
	Role           string
}

// NewUser create a new user
func NewUser(username string, password string, role string) (*User, error) {
	if username == "" {
		return nil, errors.New("username can't be empty")
	}

	if password == "" {
		return nil, errors.New("password can't be empty")
	}

	if role == "" {
		return nil, errors.New("role can't be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "cannot hash password")
	}

	user := &User{
		Username:       username,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}

	return user, nil
}

// IsCorrectPassword test if username/password is correct
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}

// Clone a user
func (user *User) Clone() *User {
	return &User{
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		Role:           user.Role,
	}
}
