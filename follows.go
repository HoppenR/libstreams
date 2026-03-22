package libstreams

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type TwitchFollows struct {
	Data       []TwitchFollowID `json:"data"`
	Pagination FollowPagination `json:"pagination"`
	Total      int              `json:"total"`
}

type TwitchFollowID struct {
	BroadcasterID   string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
}

type FollowPagination struct {
	Cursor string `json:"cursor"`
}

func (lhs *TwitchFollows) update(rhs *TwitchFollows) {
	lhs.Data = append(lhs.Data, rhs.Data...)
	lhs.Pagination = rhs.Pagination
}

func getTwitchFollowsPart(userAccessToken, clientID, userID, pagCursor string) (*TwitchFollows, error) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/channels/followed", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+userAccessToken)
	req.Header.Add("Client-Id", clientID)
	query := make(url.Values)
	query.Add("user_id", userID)
	query.Add("first", "100")
	if pagCursor != "" {
		query.Add("after", pagCursor)
	}
	req.URL.RawQuery = query.Encode()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	defer resp.Body.Close()

	page := new(TwitchFollows)
	err = json.NewDecoder(resp.Body).Decode(page)
	if err != nil {
		return nil, err
	}
	return page, nil
}

// GetTwitchFollows takes a userID and returns all follows
func GetTwitchFollows(userAccessToken, clientID, userID string) (*TwitchFollows, error) {
	follows, err := getTwitchFollowsPart(userAccessToken, clientID, userID, "")
	if err != nil {
		return nil, err
	}
	for len(follows.Data) < follows.Total {
		nextPage, err := getTwitchFollowsPart(userAccessToken, clientID, userID, follows.Pagination.Cursor)
		if err != nil {
			return nil, err
		}
		follows.update(nextPage)
	}
	return follows, nil
}
