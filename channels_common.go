// Package libstreams provides common typing and utility functions for
// client-server relationships to exchange data about twitch and strims streams.
package libstreams

import (
	"encoding/gob"
	"fmt"
	"io"
)

type StreamData interface {
	GetName() string
	GetService() string
	IsFollowed() bool
}

type Streams struct {
	Strims *StrimsStreams
	Twitch *TwitchStreams
}

func DecodeStreams(r io.Reader) (*Streams, error) {
	streams := new(Streams)
	dec := gob.NewDecoder(r)
	err := dec.Decode(streams)
	if err != nil {
		return nil, fmt.Errorf("decoding Streams failed: %w", err)
	}
	return streams, nil
}
