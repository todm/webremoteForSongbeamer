package thumbnailstealer

import (
	"image"
	"image/color"
	"image/draw"
)

const assumedSafePadding = 32

func AutoCropImage(img *image.RGBA, bgColor color.RGBA) *image.RGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var paddingLeft, paddingRight, paddingBottom int

	for x := 0; x < w; x++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, x, assumedSafePadding)) {
			paddingLeft = x
			break
		}
	}

	for x := 0; x < w; x++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, w-1-x, assumedSafePadding)) {
			paddingRight = x
			break
		}
	}

	for y := 0; y < h; y++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, paddingLeft, h-1-y)) {
			paddingBottom = y
			break
		}
	}

	cropRect := image.Rect(
		img.Bounds().Min.X+paddingLeft,
		img.Bounds().Min.Y,
		img.Bounds().Max.X-paddingRight,
		img.Bounds().Max.Y-paddingBottom,
	)

	cropped := img.SubImage(cropRect).(*image.RGBA)
	return cropped
}

func AutoSplitImage(img *image.RGBA, bgColor color.RGBA) []*image.RGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var isBar bool
	var thumbnailWidth, thumbnailHeight, barWidth, barHeight, startX int

	for y := 0; y < h; y++ {
		isBar = true
		for x := 0; x < w; x++ {
			if !ColorsEqual(bgColor, RGBAAtP(img, x, h-y-1)) {
				isBar = false
				break
			}
		}
		if isBar {
			thumbnailHeight = y
			break
		}
	}
	for y := thumbnailHeight; y < h; y++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, 0, h-y-1)) {
			barHeight = y - thumbnailHeight
			break
		}
	}

	for x := 0; x < w; x++ {
		isBar = true
		for y := 0; y < h; y++ {
			if !ColorsEqual(bgColor, RGBAAtP(img, x, y)) {
				isBar = false
				break
			}
		}
		if isBar {
			thumbnailWidth = x
			break
		}
	}
	for x := thumbnailWidth; x < w; x++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, x, assumedSafePadding)) {
			barWidth = x - thumbnailWidth
			break
		}
	}

	for x := 0; x < w; x++ {
		if !ColorsEqual(bgColor, RGBAAtP(img, w-1-x, h-1)) {
			startX = w - x
			break
		}
	}

	thumbs := make([]*image.RGBA, 0)

	if thumbnailWidth <= 0 || thumbnailHeight <= 0 {
		thumbs = append(thumbs, img)
		return thumbs
	}

	currentX := startX - thumbnailWidth
	currentY := h - thumbnailHeight
	for currentY > 0 {
		cropRect := image.Rect(
			img.Bounds().Min.X+currentX,
			img.Bounds().Min.Y+currentY,
			img.Bounds().Min.X+currentX+thumbnailWidth,
			img.Bounds().Min.Y+currentY+thumbnailHeight,
		)
		th := img.SubImage(cropRect).(*image.RGBA)

		thumbs = append(thumbs, th)

		currentX -= thumbnailWidth + barWidth
		if currentX < 0 {
			currentX = w - thumbnailWidth
			currentY -= barHeight + thumbnailHeight
		}
	}

	return thumbs
}

func GridSplitImage(img *image.RGBA, cols int, numThumbnails int) []*image.RGBA {
	if cols > numThumbnails {
		cols = numThumbnails
	}

	rows := (numThumbnails + cols - 1) / cols
	thumbnailWidth := img.Bounds().Dx() / cols
	thumbnailHeight := img.Bounds().Dy() / rows

	thumbs := make([]*image.RGBA, 0)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			if len(thumbs) >= numThumbnails {
				break
			}
			cropRect := image.Rect(
				img.Bounds().Min.X+col*thumbnailWidth,
				img.Bounds().Min.Y+row*thumbnailHeight,
				img.Bounds().Min.X+(col+1)*thumbnailWidth,
				img.Bounds().Min.Y+(row+1)*thumbnailHeight,
			)
			th := img.SubImage(cropRect).(*image.RGBA)
			thumbs = append(thumbs, th)
		}
	}

	for i, j := 0, len(thumbs)-1; i < j; i, j = i+1, j-1 {
		thumbs[i], thumbs[j] = thumbs[j], thumbs[i]
	}

	return thumbs
}

func CombineImages(images []*image.RGBA) *image.RGBA {
	c := len(images)
	w := images[0].Bounds().Dx()
	h := images[0].Bounds().Dy() * c
	combined := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(images); i++ {
		img := images[len(images)-1-i]
		draw.Draw(
			combined,
			image.Rect(0, i*img.Bounds().Dy(), img.Bounds().Dx(), i*img.Bounds().Dy()+img.Bounds().Dy()),
			img,
			img.Bounds().Min,
			draw.Src,
		)
	}

	return combined
}

func RGBAAtP(img *image.RGBA, x int, y int) color.RGBA {
	return img.RGBAAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y)
}

func ColorsEqual(c1 color.RGBA, c2 color.RGBA) bool {
	return c1.R == c2.R && c1.G == c2.G && c1.B == c2.B && c1.A == c2.A
}
