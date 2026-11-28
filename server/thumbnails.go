//go:build !windows

package main

import (
	"errors"
	"image"

	"github.com/kbinani/screenshot"
)

func GetThumbnail(index int) (*image.RGBA, error) {
	return nil, errors.New("thumbnails not supported on this platform")
}

func GetPreview() (*image.RGBA, error) {
	bounds := screenshot.GetDisplayBounds(0)
	return screenshot.CaptureRect(bounds)
}
