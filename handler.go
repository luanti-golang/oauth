package oauth

import (
	"net/http"
)

type OauthHandler struct {
	provider OauthProvider
	cfg      *OAuthConfig
	cb       OauthCallback
}

func NewHandler(cb OauthCallback, cfg *OAuthConfig) *OauthHandler {
	var provider OauthProvider
	switch cfg.Provider {
	case ProviderTypeCDB:
		provider = &CDBOauth{}
	case ProviderTypeDiscord:
		provider = &DiscordOauth{}
	case ProviderTypeGithub:
		provider = &GithubOauth{}
	case ProviderTypeMesehub:
		provider = &MesehubOauth{}
	case ProviderTypeCodeberg:
		provider = &CodebergOauth{}
	default:
		panic("unkown provider-type: " + cfg.Provider)
	}

	return &OauthHandler{provider: provider, cfg: cfg, cb: cb}
}

func (h *OauthHandler) LoginURL() string {
	return h.provider.LoginURL(h.cfg)
}

func (h *OauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := h.cfg.logger()

	list := r.URL.Query()["code"]
	if len(list) == 0 {
		// providers redirect with ?error=...&error_description=... if the user denied access
		log.Warn("oauth callback without code",
			"error", r.URL.Query().Get("error"),
			"error_description", r.URL.Query().Get("error_description"),
		)
		SendError(w, 500, "no code found")
		return
	}

	code := list[0]

	access_token, err := h.provider.RequestAccessToken(code, h.cfg)
	if err != nil {
		log.Error("access token request failed", "error", err)
		SendError(w, 500, err.Error())
		return
	}

	info, err := h.provider.RequestUserInfo(access_token, h.cfg)
	if err != nil {
		log.Error("user info request failed", "error", err)
		SendError(w, 500, err.Error())
		return
	}

	if info.ExternalID == "" {
		log.Error("user info without external_id", "name", info.Name)
		SendError(w, 500, "empty external_id")
		return
	}

	log.Debug("oauth login successful", "name", info.Name, "external_id", info.ExternalID)

	err = h.cb(w, r, info)
	if err != nil {
		log.Error("oauth callback failed", "name", info.Name, "error", err)
		SendError(w, 500, err.Error())
		return
	}

}
