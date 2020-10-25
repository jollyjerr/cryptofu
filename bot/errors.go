package bot

import "errors"

var (
	// ErrPing means the api poke failed
	ErrPing = errors.New("ping")
	// ErrTicker means the api call to get a market ticker failed
	ErrTicker = errors.New("ticker")
)
