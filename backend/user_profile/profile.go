package user_profile

import (
	"fmt"
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
	//ID is a player unique in database
	ID          string   `json:"id"`
	ExternalIDs []string `json:"ids"`
	Name        string   `json:"name"`
	Picture     string   `json:"picture"`
	jwt.StandardClaims
}

func CreateAuthToken(IDs []string, ID, name, avatar string) (string, error) {
	if ID == "" {
		return "", fmt.Errorf("ID should be specified for usertoken")
	}
	claims := UserProfile{
		ID:          ID,
		ExternalIDs: IDs,
		Name:        name,
		Picture:     avatar,
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
	_, err := jwt.ParseWithClaims(cookie, profile, func(token *jwt.Token) (interface{}, error) {
		return config.JwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return profile, err
}

func GetUserProfile(r *http.Request) (*UserProfile, error) {
	profile, ok := r.Context().Value("user").(*UserProfile)
	if !ok || profile == nil {
		return nil, glog.Error("user profile is not found in profile")
	}
	return profile, nil
}

type UserResponse struct {
	Code        string   `json:"code"`
	ExternalIDs []string `json:"external_ids"`
	Name        string   `json:"name"`
	Avatar      string   `json:"avatar"`
	Alias       string   `json:"alias"`
}
