package auth

import (
	// "regexp"
	"testing"
)

// TestNewUser create a new user with a name, password and user role
// checking for a valid return value
func TestNewUser(t *testing.T) {
	user, err := NewUser("user1", "secret1", "user")
	if err != nil {
		t.Fatalf(`Something wrong %v %v`, err, user)
	}
}

// TestNewUser create a new user with a name, password and user role
// checking for a valid return value
func TestNewUserWithEmptyUsername(t *testing.T) {
	user, err := NewUser("", "secret1", "user")
	if err == nil {
		t.Fatalf(`Empty username should not be allow %v %v`, err, user)
	}
}

// TestNewUserWithoutPassword create a new user with a name, without password and admin role
// checking for a valid return value
func TestNewUserWithoutPassword(t *testing.T) {
	user, err := NewUser("admin2", "", "admin")
	if err == nil {
		t.Fatalf(`User with empty password should not be allowed: %#v`, user)
	}
}

// TestNewUserWithEmptyRole create a new user with a name, password and an empty role
// checking for a valid return value
func TestNewUserWithEmptyRole(t *testing.T) {
	user, err := NewUser("user3", "secret3", "")
	if err == nil {
		t.Fatalf(`Empty role should not be allow %v %v`, err, user)
	}
}

// TestIsCorrectPassword username/password checking
// checking for a valid return value
func TestUsernamePassword(t *testing.T) {
	user, err := NewUser("user1", "secret1", "user")
	if err != nil {
		t.Fatalf(`Something wrong %v %v`, err, user)
	}
	if ok := user.IsCorrectPassword("secret1"); !ok {
		t.Fatalf(`Password checking failed %v`, err)
	}
}

func TestCloneUser(t *testing.T) {
	user1, err := NewUser("user1", "secret1", "user")
	if err != nil {
		t.Fatalf(`Empty username should not be allow %v %v`, err, user1)
	}

	user2, err := NewUser("user2", "secret2", "user")
	if err != nil {
		t.Fatalf(`Empty username should not be allow %v %v`, err, user2)
	}

	user3 := user1.Clone()
	if !(*user1 == *user3) {
		t.Fatalf(`User1 and user3 must be equal %v == %v`, user1, user3)
	}

	if !(*user1 != *user2) {
		t.Fatalf(`User1 and user2 must not be equal %v != %v`, user1, user2)
	}

}
