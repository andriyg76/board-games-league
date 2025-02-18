package auth

import (
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"net/http"
	"sync"
)

var gothInitOnce sync.Once

func ensureGothInit(r *http.Request) {
	gothInitOnce.Do(func() {
		glog.Info("Late goth init...")
		hostName := utils.GetHostUrl(r)

		callbackUrl := hostName + "/ui/auth-callback" // defined at frontend/src/router/index.ts

		glog.Info("Google auth callback url: %v", callbackUrl)

		goth.UseProviders(
			google.New(
				config.GoogleClientID,
				config.GoogleClientSecret,
				callbackUrl,
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			),
		)
	})
}

type GothProvider struct{}

func (p GothProvider) BeginUserAuthHandler(w http.ResponseWriter, r *http.Request) {
	ensureGothInit(r)

	gothic.BeginAuthHandler(w, r)
}

func init() {
	gothic.Store = store
}

func (p GothProvider) CompleteUserAuthHandler(w http.ResponseWriter, r *http.Request) (ExternalUser, error) {
	ensureGothInit(r)

	auth, err := gothic.CompleteUserAuth(w, r)
	var user ExternalUser
	if err == nil {
		user.Name = auth.Name
		user.Email = auth.Email
		user.Avatar = auth.AvatarURL
	}
	return user, err
}

func (p GothProvider) LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	ensureGothInit(r)

	return gothic.Logout(w, r)
}
