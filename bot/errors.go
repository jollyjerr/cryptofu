package bot

import "errors"

var (
	// ErrPing means the api poke failed
	ErrPing = errors.New("ping")
)
