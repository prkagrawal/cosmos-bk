package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/prkagrawal/cosmos-bk2/graph/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("FRONTEND_URL") + "/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(authSvc *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		token, err := GoogleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusBadRequest)
			return
		}

		client := GoogleOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "Failed to get user info", http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email      string `json:"email"`
			Name       string `json:"name"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to decode user info", http.StatusBadRequest)
			return
		}

		// Handle user creation/login logic
		var user model.User
		if err := authSvc.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
			// Create new user
			user = model.User{
				Email:     userInfo.Email,
				FirstName: userInfo.GivenName,
				LastName:  userInfo.FamilyName,
				Role:      model.Volunteer,
			}
			if err := authSvc.DB.Create(&user).Error; err != nil {
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
		}

		// Generate JWT and return to frontend
		jwtToken, err := GenerateJWT(&user)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"?token="+jwtToken, http.StatusTemporaryRedirect)
	}
}
