//go:build windows

package thumbnailstealer

import (
	"errors"
	"image"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func Steal(skipReset bool, numThumbnails int, cols int) (*image.RGBA, error) {
	proc := findProcessByName("SongBeamer.exe")
	if proc == 0 {
		return nil, errors.New("could not find process")
	}

	hWndMain := getWindowHandle(proc, "TFormSBMain")
	hWndThumb := getWindowHandle(proc, "TPresentationThumbNails")

	if IsIconic(hWndMain) {
		ShowWindow(hWndMain, 9) // SW_RESTORE
	}

	if hasScrollbar(hWndThumb) {
		var thumbRect RECT
		GetWindowRect(hWndThumb, &thumbRect)

		scrHeight := getScrollableHeight(hWndThumb)
		SetWindowPos(hWndThumb, 0, int(thumbRect.Left), int(thumbRect.Top), int(thumbRect.Right-thumbRect.Left), int(scrHeight+10), 2) //SWP_NOMOVE
		time.Sleep(16 * time.Millisecond)
		if !skipReset {
			defer SetWindowPos(hWndThumb, 0, int(thumbRect.Left), int(thumbRect.Top), int(thumbRect.Right-thumbRect.Left), int(thumbRect.Bottom-thumbRect.Top), 2)
		}

	}
	var clientRect RECT

	GetWindowRect(hWndThumb, &clientRect)
	width := clientRect.Right - clientRect.Left
	height := clientRect.Bottom - clientRect.Top

	img := captureWindow(hWndThumb, int(width), int(height))
	bgColor := RGBAAtP(img, 0, img.Bounds().Dy()-1)
	cropped := AutoCropImage(img, bgColor)

	var thumbnails []*image.RGBA
	if numThumbnails > 0 && cols > 0 {
		thumbnails = GridSplitImage(cropped, cols, numThumbnails)
	} else {
		thumbnails = AutoSplitImage(cropped, bgColor)
	}

	combined := CombineImages(thumbnails)

	return combined, nil
}

func findProcessByName(name string) uint32 {
	snapshot := CreateToolhelp32Snapshot(0x02, 0) //TH32CS_SNAPPROCESS
	defer syscall.CloseHandle(syscall.Handle(snapshot))
	if snapshot == 0 {
		return 0
	}

	var p PROCESSENTRY32
	for ok := Process32First(snapshot, &p); ok; ok = Process32Next(snapshot, &p) {
		if strings.EqualFold(syscall.UTF16ToString(p.SzExeFile[:]), name) {
			return p.Th32ProcessID
		}
	}

	return 0
}

func getWindowHandle(proc uint32, className string) uintptr {
	var handle uintptr
	var callback uintptr
	buffer := make([]uint16, 256)

	callback = syscall.NewCallback(func(hWnd uintptr, _ uintptr) uintptr {
		windowProc := GetWindowThreadProcessId(hWnd)
		if windowProc != proc {
			return 1
		}
		copied := GetClassName(hWnd, buffer)
		if strings.EqualFold(syscall.UTF16ToString(buffer[:copied]), className) {
			handle = hWnd
			return 0
		}
		EnumChildWindows(hWnd, callback, unsafe.Pointer(nil))
		return 1
	})
	EnumWindows(callback, unsafe.Pointer(nil))
	return uintptr(handle)
}

func hasScrollbar(hWnd uintptr) bool {
	var scrollbarInfo SCROLLBARINFO
	GetScrollBarInfo(hWnd, 0xFFFFFFFB, &scrollbarInfo) // OBJID_VSCROLL
	return scrollbarInfo.RgState[0]&0x8000 == 0        // STATE_SYSTEM_INVISIBLE
}

func getScrollableHeight(hWnd uintptr) int32 {
	var si SCROLLINFO
	si.FMask = 0x0001 | 0x0002 | 0x0004 // SIF_RANGE | SIF_PAGE | SIF_POS
	GetScrollInfo(hWnd, 0x01, &si)      // SB_VERT
	return si.NMax - si.NMin + 1
}

func captureWindow(hWnd uintptr, w int, h int) *image.RGBA {
	if w <= 0 || h <= 0 {
		return nil
	}

	dc := GetDC(hWnd)
	defer ReleaseDC(hWnd, dc)

	mdc := CreateCompatibleDC(dc)
	defer DeleteDC(mdc)

	bmp := CreateCompatibleBitmap(dc, w, h)
	defer DeleteObject(bmp)

	old := SelectObject(mdc, bmp)
	defer SelectObject(mdc, old)

	PrintWindow(hWnd, mdc, 0)

	sdc := GetDC(0)
	defer ReleaseDC(0, sdc)

	bmi := BITMAPINFOHEADER{BiWidth: int32(w), BiHeight: int32(-h), BiPlanes: 1, BiBitCount: 32}
	buf := make([]byte, w*h*4)
	GetDIBits(mdc, bmp, 0, uint(h), &buf[0], &bmi, 0)

	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	dst := img.Pix

	for i := 0; i < len(buf); i += 4 {
		dst[i+0] = buf[i+2] // R
		dst[i+1] = buf[i+1] // G
		dst[i+2] = buf[i+0] // B
		dst[i+3] = buf[i+3] // A
	}

	return img
}
