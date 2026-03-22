package libstreams

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type TwitchStreams struct {
	Data []TwitchStreamData `json:"data"`
}

type TwitchStreamData struct {
	GameName     string    `json:"game_name"`
	Language     string    `json:"language"`
	StartedAt    time.Time `json:"started_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Title        string    `json:"title"`
	UserName     string    `json:"user_name"`
	ViewerCount  int       `json:"viewer_count"`
}

type UserDatas struct {
	Data []UserData `json:"data"`
}

type UserData struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

func (ts *TwitchStreamData) GetName() string {
	return ts.UserName
}

func (ts *TwitchStreamData) GetService() string {
	return "twitch"
}

func (ts *TwitchStreamData) IsFollowed() bool {
	return true
}

func (ts *TwitchStreams) update(rhs *TwitchStreams) {
	ts.Data = append(ts.Data, rhs.Data...)
}

func (ts *TwitchStreams) Less(i, j int) bool {
	return ts.Data[i].ViewerCount < ts.Data[j].ViewerCount
}

func (ts *TwitchStreams) Len() int {
	return len(ts.Data)
}

func (ts *TwitchStreams) Swap(i, j int) {
	ts.Data[i], ts.Data[j] = ts.Data[j], ts.Data[i]
}

func getLiveTwitchStreamsPart(token, clientID string, twitchFollows *TwitchFollows, first int, dst *TwitchStreams) error {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Client-Id", clientID)
	query := make(url.Values)
	for i := first; i < twitchFollows.Total && i < (first+100); i++ {
		query.Add("user_id", twitchFollows.Data[i].BroadcasterID)
	}
	query.Add("first", "100")
	req.URL.RawQuery = query.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	defer resp.Body.Close()

	part := new(TwitchStreams)
	err = json.NewDecoder(resp.Body).Decode(part)
	if err != nil {
		return err
	}
	dst.update(part)
	return nil
}

// GetLiveTwitchStreams takes follow IDs and returns which ones are live
func GetLiveTwitchStreams(token, clientID string, twitchFollows *TwitchFollows) (*TwitchStreams, error) {
	twitchStreams := new(TwitchStreams)
	var err error
	for i := 0; i < twitchFollows.Total; i += 100 {
		err = getLiveTwitchStreamsPart(token, clientID, twitchFollows, i, twitchStreams)
		if err != nil {
			return nil, err
		}
	}
	return twitchStreams, nil
}
