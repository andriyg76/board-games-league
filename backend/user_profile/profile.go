package user_profile

import (
	"github.com/andriyg76/bgl/utils"
	"github.com/andriyg76/glog"
	"github.com/golang-jwt/jwt/v5"
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
	Code        string   `json:"code"`
	ExternalIDs []string `json:"ids"`
	Name        string   `json:"name"`
	Picture     string   `json:"picture"`
	jwt.RegisteredClaims
}

func (p UserProfile) GetExpirationTime() (*jwt.NumericDate, error) {
	return p.ExpiresAt, nil
}
func (p UserProfile) GetIssuedAt() (*jwt.NumericDate, error) {
	return p.IssuedAt, nil
}
func (p UserProfile) GetNotBefore() (*jwt.NumericDate, error) {
	return p.NotBefore, nil
}
func (p UserProfile) GetIssuer() (string, error) {
	return p.Issuer, nil
}
func (p UserProfile) GetSubject() (string, error) {
	return p.Subject, nil
}
func (p UserProfile) GetAudience() (jwt.ClaimStrings, error) {
	return p.Audience, nil
}

func CreateAuthToken(IDs []string, Code, name, avatar string) (string, error) {
	if Code == "" {
		return "", glog.Error("code should be specified for usertoken.")
	}
	claims := UserProfile{
		Code:        Code,
		ExternalIDs: IDs,
		Name:        name,
		Picture:     avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JwtSecret)
}

func ParseProfile(cookie string) (*UserProfile, error) {
	profile := &UserProfile{}
	_, err := jwt.ParseWithClaims(cookie, profile, func(token *jwt.Token) (any, error) {
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
		return nil, glog.Error("user profile is not found in profile.")
	}
	return profile, nil
}

type UserResponse struct {
	Code        string   `json:"code"`
	ExternalIDs []string `json:"external_ids"`
	Name        string   `json:"name"`
	Names       []string `json:"names"`
	Avatar      string   `json:"avatar"`
	Avatars     []string `json:"avatars"`
	Alias       string   `json:"alias"`
}
