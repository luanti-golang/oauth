package oauth

import (
	"log/slog"
	"net/http"
)

type AccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type OauthUserInfo struct {
	Provider    ProviderType
	Name        string
	ExternalID  string
	AvatarURL   string
	DisplayName string
}

type OAuthConfig struct {
	Provider    ProviderType
	ClientID    string
	Secret      string
	CallbackURL string
	// optional logger, defaults to slog.Default()
	Logger *slog.Logger
}

type ProviderType string

const (
	ProviderTypeGithub   ProviderType = "GITHUB"
	ProviderTypeDiscord  ProviderType = "DISCORD"
	ProviderTypeMesehub  ProviderType = "MESEHUB"
	ProviderTypeCDB      ProviderType = "CDB"
	ProviderTypeCodeberg ProviderType = "CODEBERG"
)

type OauthProvider interface {
	// client login url
	LoginURL(cfg *OAuthConfig) string
	RequestAccessToken(code string, cfg *OAuthConfig) (string, error)
	RequestUserInfo(access_token string, cfg *OAuthConfig) (*OauthUserInfo, error)
}

type OauthCallback func(w http.ResponseWriter, r *http.Request, user_info *OauthUserInfo) error
