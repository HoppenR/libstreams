package libstreams

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type AuthData struct {
	AppAccessToken  *AppAccessToken
	ClientID        string
	UserAccessToken *UserAccessToken
	UserID          string
	UserName        string
	cacheFolder     string
	clientSecret    string
}

// Helper embeddable struct to implement helper functions like IsExpired
type expirableTokenBase struct {
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresIn int       `json:"expires_in"` // In seconds
}

type AppAccessToken struct {
	expirableTokenBase
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type UserAccessToken struct {
	expirableTokenBase
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type ValidateUserAccessTokenResponse struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserID    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
}

var ErrUnauthorized = errors.New("401 Unauthorized")

func NewAuthData() *AuthData {
	return &AuthData{
		AppAccessToken:  new(AppAccessToken),
		UserAccessToken: new(UserAccessToken),
	}
}

func (etb *expirableTokenBase) IsExpired(buffer time.Duration) bool {
	expiresInDuration := time.Duration(etb.ExpiresIn) * time.Second
	expirationTime := etb.IssuedAt.Add(expiresInDuration).Add(-buffer)
	return time.Now().After(expirationTime)
}

func (ad *AuthData) SetCacheFolder(name string) error {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	abspath := filepath.Join(cachedir, name)
	err = os.MkdirAll(abspath, os.ModePerm)
	if err != nil {
		return err
	}
	ad.cacheFolder = abspath
	return nil
}

func (ad *AuthData) SetClientID(clientID string) *AuthData {
	ad.ClientID = clientID
	return ad
}

func (ad *AuthData) SetClientSecret(clientSecret string) *AuthData {
	ad.clientSecret = clientSecret
	return ad
}

func (ad *AuthData) SetUserName(userName string) *AuthData {
	ad.UserName = userName
	return ad
}

func (ad *AuthData) GetCachedData() error {
	if ad.cacheFolder == "" {
		return errors.New("cache folder not set")
	}
	var appAccessToken AppAccessToken
	err := ad.readCache("cachedtoken", &appAccessToken)
	if err != nil {
		return err
	}
	if !appAccessToken.IsExpired(time.Duration(0)) {
		ad.AppAccessToken = &appAccessToken
	}
	var userAccessToken UserAccessToken
	err = ad.readCache("cacheduseraccesstoken", &userAccessToken)
	if err != nil {
		return err
	}
	if !userAccessToken.IsExpired(time.Duration(0)) {
		ad.UserAccessToken = &userAccessToken
	}
	var userID string
	err = ad.readCache("cacheduserid", &userID)
	if err != nil {
		return err
	}
	ad.UserID = userID
	return nil
}

func (ad *AuthData) GetAppAccessToken() error {
	if ad.AppAccessToken.IsExpired(time.Duration(0)) {
		err := ad.FetchAppAccessToken()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ad *AuthData) GetUserID() error {
	if ad.UserID == "" {
		err := ad.FetchUserID()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ad *AuthData) SaveCachedData() error {
	if ad.cacheFolder == "" {
		return errors.New("cache folder not set")
	}
	if ad.AppAccessToken != nil {
		err := ad.writeCache("cachedtoken", ad.AppAccessToken)
		if err != nil {
			return err
		}
	}
	if ad.UserAccessToken != nil {
		err := ad.writeCache("cacheduseraccesstoken", ad.UserAccessToken)
		if err != nil {
			return err
		}
	}
	if ad.UserID != "" {
		err := ad.writeCache("cacheduserid", ad.UserID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ad *AuthData) writeCache(fileName string, data any) error {
	tokenfile, err := os.Create(filepath.Join(ad.cacheFolder, fileName))
	if err != nil {
		return err
	}
	defer tokenfile.Close()

	return json.NewEncoder(tokenfile).Encode(data)
}

func (ad *AuthData) FetchAppAccessToken() error {
	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Client-Id", ad.ClientID)
	query := make(url.Values)
	query.Add("client_secret", ad.clientSecret)
	query.Add("grant_type", "client_credentials")
	req.URL.RawQuery = query.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(ad.AppAccessToken)
	if err != nil {
		return err
	}
	ad.AppAccessToken.IssuedAt = time.Now()
	return err
}

func (ad *AuthData) ExchangeCodeForUserAccessToken(authorizationCode string, redirectURL string) (*UserAccessToken, error) {
	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		return nil, err
	}

	query := make(url.Values)
	query.Add("client_id", ad.ClientID)
	query.Add("client_secret", ad.clientSecret)
	query.Add("code", authorizationCode)
	query.Add("grant_type", "authorization_code")
	query.Add("redirect_uri", redirectURL)
	req.URL.RawQuery = query.Encode()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	token := new(UserAccessToken)
	err = json.NewDecoder(resp.Body).Decode(token)
	token.IssuedAt = time.Now()
	return token, err
}

func (ad *AuthData) ValidateUserAccessToken(token *UserAccessToken) (*ValidateUserAccessTokenResponse, error) {
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "OAuth "+token.AccessToken)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	validateTokenResp := new(ValidateUserAccessTokenResponse)
	err = json.NewDecoder(resp.Body).Decode(validateTokenResp)
	if err != nil {
		return nil, err
	}
	return validateTokenResp, nil
}

func (ad *AuthData) RefreshUserAccessToken() error {
	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		return err
	}

	query := make(url.Values)
	query.Add("client_id", ad.ClientID)
	query.Add("client_secret", ad.clientSecret)
	query.Add("grant_type", "refresh_token")
	query.Add("refresh_token", ad.UserAccessToken.RefreshToken)
	req.URL.RawQuery = query.Encode()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	err = json.NewDecoder(resp.Body).Decode(ad.UserAccessToken)
	if err != nil {
		return err
	}
	ad.UserAccessToken.IssuedAt = time.Now()
	return err
}

func (ad *AuthData) FetchUserID() error {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+ad.AppAccessToken.AccessToken)
	req.Header.Add("Client-Id", ad.ClientID)
	query := make(url.Values)
	query.Add("login", ad.UserName)
	req.URL.RawQuery = query.Encode()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	userDatas := new(UserDatas)
	err = json.NewDecoder(resp.Body).Decode(userDatas)
	if err != nil {
		return err
	}
	if len(userDatas.Data) == 0 {
		return errors.New("userid response contained no user id ")
	}
	ad.UserID = userDatas.Data[0].ID
	return nil
}

func (ad *AuthData) readCache(fileName string, v any) error {
	path := filepath.Join(ad.cacheFolder, fileName)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}
