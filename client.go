package libstreams

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type RedirectError struct {
	Location string
}

func (re *RedirectError) Error() string {
	return fmt.Sprintf("redirect to %s", re.Location)
}

func GetServerData(ctx context.Context, address string) (*Streams, error) {
	var noRedirectClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/octet-stream")

	var resp *http.Response
	resp, err = noRedirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching Streams failed: %w", err)
	}
	defer resp.Body.Close()

	var streams *Streams
	streams, err = handleServerResponse(resp)
	if err != nil {
		return nil, err
	}
	return streams, nil
}

func handleServerResponse(resp *http.Response) (*Streams, error) {
	var err error

	switch resp.StatusCode {
	case http.StatusOK:
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/octet-stream") {
		    return nil, fmt.Errorf("unexpected content type: %s", contentType)
		}

		streams := new(Streams)
		dec := gob.NewDecoder(resp.Body)
		err = dec.Decode(streams)
		if err != nil {
			return nil, fmt.Errorf("decoding Streams failed: %w", err)
		}
		return streams, nil
	case http.StatusFound:
		location := resp.Header.Get("Location")
		var relURL *url.URL
		relURL, err = url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("could not parse redirect location: %w", err)
		}
		var absoluteURL *url.URL
		absoluteURL = resp.Request.URL.ResolveReference(relURL)
		return nil, &RedirectError{Location: absoluteURL.String()}
	default:
		return nil, fmt.Errorf("status getting streams: %d", resp.StatusCode)
	}
}
