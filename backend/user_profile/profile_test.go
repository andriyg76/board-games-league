package user_profile

import (
	"github.com/andriyg76/bgl/asserts2"
	"github.com/markbates/goth"
	"testing"
)

func TestCreateAuthToken(t *testing.T) {
	user := goth.User{
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "http://example.com/avatar.jpg",
	}

	token, err := CreateAuthToken(user.Email, user.Name, user.AvatarURL)
	asserts := asserts2.Get(t)
	asserts.
		NoError(err).
		NotEmpty(t, token)

	t.Log("Parsing profile: ")

	restore, err := ParseProfile(token)
	asserts.NoError(err).Equal(user.Email, restore.Email)
}
