package user_profile

import (
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"time"
)

var config = struct {
	JwtSecret []byte
}{
	JwtSecret: func() []byte {
		secret := []byte(os.Getenv("JWT_SECRET"))
		if len(secret) == 0 {
			glog.Warn("Generating JWT_SECRET")
			secret = utils.GenerateRandomKey(32)
		} else {
			glog.Info("JWT_SECRET is resolved %d-th lenght", len(secret))
		}
		return secret
	}(),
}

func Test() {

}

type UserProfile struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.StandardClaims
}

func CreateAuthToken(email, name, avatar string) (string, error) {
	claims := UserProfile{
		Email:   email,
		Name:    name,
		Picture: avatar,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JwtSecret)
}

func ParseProfile(cookie string) (*UserProfile, error) {
	profile := &UserProfile{}
	_, error := jwt.ParseWithClaims(cookie, profile, func(token *jwt.Token) (interface{}, error) {
		return config.JwtSecret, nil
	})
	if error != nil {
		return nil, error
	}
	return profile, error
}

func GetUserProfile(r *http.Request) (*UserProfile, error) {
	profile, ok := r.Context().Value("user").(*UserProfile)
	if !ok || profile == nil {
		return nil, glog.Error("user profile is not found in profile")
	}
	return profile, nil
}
