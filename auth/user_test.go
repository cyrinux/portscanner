package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewUser_InvalidInput(t *testing.T) {
	_, err := NewUser("", "secret1", "user")
	assert.Error(t, err, "Should have failed with invalid username")

	_, err = NewUser("admin2", "", "admin")
	assert.Error(t, err, "Should have failed with invalid password")

	_, err = NewUser("user3", "secret3", "")
	assert.Error(t, err, "Should have failed with invalid role")
}

func Test_IsCorrectPassword_ValidatePassword(t *testing.T) {
	goodpass := "goodpass"
	user, _ := NewUser("user1", goodpass, "user")
	assert.True(t, user.IsCorrectPassword("goodpass"), "Password should be ok")
	assert.False(t, user.IsCorrectPassword("badpass"), "Password should not be ok")
	assert.NotEqual(t, goodpass, user.HashedPassword, "Password should not be in cleartext")
}

// func TestCloneUser(t *testing.T) {
// 	user1, err := NewUser("user1", "secret1", "user")
// 	if err != nil {
// 		t.Fatalf(`Empty username should not be allow %v %v`, err, user1)
// 	}

// 	user2, err := NewUser("user2", "secret2", "user")
// 	if err != nil {
// 		t.Fatalf(`Empty username should not be allow %v %v`, err, user2)
// 	}

// 	user3 := user1.Clone()
// 	if !(*user1 == *user3) {
// 		t.Fatalf(`User1 and user3 must be equal %v == %v`, user1, user3)
// 	}

// 	if !(*user1 != *user2) {
// 		t.Fatalf(`User1 and user2 must not be equal %v != %v`, user1, user2)
// 	}

// }
