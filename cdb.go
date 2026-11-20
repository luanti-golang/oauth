package oauth

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

type CDBUserResponse struct {
	Username string `json:"username"`
}

type CDBUser struct {
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	ProfilePicURL string `json:"profile_pic_url"` // "/uploads/xyz.jpg"
}

type CDBOauth struct{}

func (o *CDBOauth) LoginURL(cfg *OAuthConfig) string {
	return fmt.Sprintf("https://content.luanti.org/oauth/authorize/?response_type=code&client_id=%s&redirect_uri=%s", cfg.ClientID, url.QueryEscape(cfg.CallbackURL))
}

func (o *CDBOauth) RequestAccessToken(code string, cfg *OAuthConfig) (string, error) {
	var data bytes.Buffer
	w := multipart.NewWriter(&data)
	w.WriteField("grant_type", "authorization_code")
	w.WriteField("client_id", cfg.ClientID)
	w.WriteField("client_secret", cfg.Secret)
	w.WriteField("code", code)
	w.Close()

	req, err := http.NewRequest("POST", "https://content.luanti.org/oauth/token/", &data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	return requestAccessToken(cfg, req)
}

func (o *CDBOauth) RequestUserInfo(access_token string, cfg *OAuthConfig) (*OauthUserInfo, error) {
	// fetch user data
	req, err := http.NewRequest("GET", "https://content.luanti.org/api/whoami/", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	userData := CDBUserResponse{}
	_, err = doJSON(cfg, "userinfo", req, &userData)
	if err != nil {
		return nil, err
	}

	info := OauthUserInfo{
		Provider:   ProviderTypeCDB,
		Name:       userData.Username,
		ExternalID: userData.Username,
	}

	// fetch user profile
	req, err = http.NewRequest("GET", fmt.Sprintf("https://content.luanti.org/api/users/%s/", url.PathEscape(userData.Username)), nil)
	if err != nil {
		return nil, fmt.Errorf("new user-profile request error: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	userProfile := CDBUser{}
	_, err = doJSON(cfg, "userprofile", req, &userProfile)
	if err != nil {
		return nil, err
	}

	if userProfile.ProfilePicURL != "" {
		info.AvatarURL = fmt.Sprintf("https://content.luanti.org%s", userProfile.ProfilePicURL)
	}
	if userProfile.DisplayName != "" {
		info.DisplayName = userProfile.DisplayName
	}

	return &info, nil
}
