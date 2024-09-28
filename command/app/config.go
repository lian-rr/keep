package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type Config struct {
	BasePath string
}

// ErrInvalidDir indicates that the root directory for the store is not valid.
var ErrInvalidDir = errors.New("invalid base directory")

func New(ctx context.Context, path string) (Config, error) {
	if path == "" {
		def, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		path = def
	}
	path = fmt.Sprintf("%s/.keep", path)

	err := os.Mkdir(path, 0o740)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return Config{}, fmt.Errorf("%w: %w", ErrInvalidDir, err)
	}

	return Config{
		BasePath: path,
	}, nil
}
