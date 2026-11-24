//go:build !windows

package main

import (
	"errors"
	"image"
)

func GetThumbnail(index int) (*image.RGBA, error) {
	return nil, errors.New("thumbnails not supported on this platform")
}
