package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type CodebergUserResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

type CodebergOauth struct{}

func (o *CodebergOauth) LoginURL(cfg *OAuthConfig) string {
	return fmt.Sprintf("https://codeberg.org/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=STATE", cfg.ClientID, url.QueryEscape(cfg.CallbackURL))
}

func (o *CodebergOauth) RequestAccessToken(code string, cfg *OAuthConfig) (string, error) {
	accessTokenReq := make(map[string]string)
	accessTokenReq["client_id"] = cfg.ClientID
	accessTokenReq["client_secret"] = cfg.Secret
	accessTokenReq["code"] = code
	accessTokenReq["grant_type"] = "authorization_code"
	accessTokenReq["redirect_uri"] = cfg.CallbackURL

	data, err := json.Marshal(accessTokenReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://codeberg.org/login/oauth/access_token", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return requestAccessToken(cfg, req)
}

func (o *CodebergOauth) RequestUserInfo(access_token string, cfg *OAuthConfig) (*OauthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://codeberg.org/api/v1/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	userData := CodebergUserResponse{}
	_, err = doJSON(cfg, "userinfo", req, &userData)
	if err != nil {
		return nil, err
	}

	external_id := strconv.Itoa(userData.ID)
	info := OauthUserInfo{
		Provider:   ProviderTypeCodeberg,
		Name:       userData.Login,
		ExternalID: external_id,
		AvatarURL:  fmt.Sprintf("https://codeberg.org/%s.png", userData.Login),
	}

	return &info, nil
}
