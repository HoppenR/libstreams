// Package libstreams provides common typing and utility functions for
// client-server relationships to exchange data about twitch and strims streams.
package libstreams

import (
	"time"
)

type StreamData interface {
	GetName() string
	GetService() string
	IsFollowed() bool
}

type Streams struct {
	Strims          *StrimsStreams
	Twitch          *TwitchStreams
	LastFetched     time.Time
	RefreshInterval time.Duration
}
