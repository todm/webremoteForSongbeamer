//go:build !windows

package thumbnailstealer

import (
	"errors"
	"image"
)

func Steal(skipReset bool, numThumbnails int, cols int) (*image.RGBA, error) {
	return nil, errors.New("not available on this platform")
}
