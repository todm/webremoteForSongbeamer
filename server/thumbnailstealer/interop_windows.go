//go:build windows

package thumbnailstealer

import (
	"syscall"
	"unsafe"
)

// **************** kernel32 **************** //
var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	pProcess32First           = kernel32.NewProc("Process32FirstW")
	pProcess32Next            = kernel32.NewProc("Process32NextW")
)

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

func CreateToolhelp32Snapshot(flags uint32, proc uint32) uintptr {
	r, _, _ := pCreateToolhelp32Snapshot.Call(uintptr(flags), uintptr(proc))
	return r
}

func Process32First(snapshot uintptr, entry *PROCESSENTRY32) bool {
	entry.DwSize = uint32(unsafe.Sizeof(*entry))
	r, _, _ := pProcess32First.Call(snapshot, uintptr(unsafe.Pointer(entry)))
	return r != 0
}

func Process32Next(snapshot uintptr, entry *PROCESSENTRY32) bool {
	entry.DwSize = uint32(unsafe.Sizeof(*entry))
	r, _, _ := pProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(entry)))
	return r != 0
}

// **************** user32 **************** //
var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pEnumWindows              = user32.NewProc("EnumWindows")
	pEnumChildWindows         = user32.NewProc("EnumChildWindows")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pGetClassNameW            = user32.NewProc("GetClassNameW")
	pGetWindowRect            = user32.NewProc("GetWindowRect")
	pGetClientRect            = user32.NewProc("GetClientRect")
	pSetWindowPos             = user32.NewProc("SetWindowPos")
	pIsIconic                 = user32.NewProc("IsIconic")
	pShowWindow               = user32.NewProc("ShowWindow")
	pPrintWindow              = user32.NewProc("PrintWindow")
	pGetDC                    = user32.NewProc("GetDC")
	pReleaseDC                = user32.NewProc("ReleaseDC")
	pGetScrollBarInfo         = user32.NewProc("GetScrollBarInfo")
	pGetScrollInfo            = user32.NewProc("GetScrollInfo")
)

type RECT struct{ Left, Top, Right, Bottom int32 }

type SCROLLBARINFO struct {
	CbSize        uint32
	RcScrollBar   RECT
	DxyLineButton int32
	XyThumbTop    int32
	XyThumbBottom int32
	Reserved      int32
	RgState       [6]uint32
}

type SCROLLINFO struct {
	CbSize    uint32
	FMask     uint32
	NMin      int32
	NMax      int32
	NPage     uint32
	NPos      int32
	NTrackPos int32
}

func EnumWindows(cb uintptr, params unsafe.Pointer) bool {
	r, _, _ := pEnumWindows.Call(cb, 0)
	return r != 0
}

func EnumChildWindows(hWnd uintptr, cb uintptr, params unsafe.Pointer) bool {
	r, _, _ := pEnumChildWindows.Call(hWnd, cb, 0)
	return r != 0
}

func GetWindowThreadProcessId(hWnd uintptr) uint32 {
	var proc uint32
	pGetWindowThreadProcessId.Call(hWnd, uintptr(unsafe.Pointer(&proc)))
	return proc
}

func GetClassName(hWnd uintptr, buffer []uint16) int32 {
	r, _, _ := pGetClassNameW.Call(hWnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return int32(r)
}

func GetWindowRect(hWnd uintptr, rect *RECT) bool {
	r, _, _ := pGetWindowRect.Call(hWnd, uintptr(unsafe.Pointer(rect)))
	return r != 0
}

func GetClientRect(hWnd uintptr, rect *RECT) bool {
	r, _, _ := pGetClientRect.Call(hWnd, uintptr(unsafe.Pointer(rect)))
	return r != 0
}

func SetWindowPos(hWnd uintptr, hWndAfter uintptr, x int, y int, w int, h int, flags int) bool {
	r, _, _ := pSetWindowPos.Call(hWnd, hWndAfter, uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(flags))
	return r != 0
}

func IsIconic(hWnd uintptr) bool {
	r, _, _ := pIsIconic.Call(hWnd)
	return r != 0
}

func ShowWindow(hWnd uintptr, nCmdShow int) bool {
	r, _, _ := pShowWindow.Call(hWnd, uintptr(nCmdShow))
	return r != 0
}

func PrintWindow(hWnd uintptr, hdc uintptr, flags uint) bool {
	r, _, _ := pPrintWindow.Call(hWnd, hdc, uintptr(flags))
	return r != 0
}

func GetDC(hwnd uintptr) uintptr {
	r, _, _ := pGetDC.Call(hwnd)
	return r
}

func ReleaseDC(hwnd, hdc uintptr) {
	pReleaseDC.Call(hwnd, hdc)
}

func GetScrollBarInfo(hWnd uintptr, idObject uint, scrollbarInfo *SCROLLBARINFO) bool {
	scrollbarInfo.CbSize = uint32(unsafe.Sizeof(*scrollbarInfo))
	r, _, _ := pGetScrollBarInfo.Call(hWnd, uintptr(idObject), uintptr(unsafe.Pointer(scrollbarInfo)))
	return r != 0
}

func GetScrollInfo(hWnd uintptr, fnBar int32, scrollInfo *SCROLLINFO) bool {
	scrollInfo.CbSize = uint32(unsafe.Sizeof(*scrollInfo))
	r, _, _ := pGetScrollInfo.Call(hWnd, uintptr(fnBar), uintptr(unsafe.Pointer(scrollInfo)))
	return r != 0
}

// **************** gdi32 **************** //

var (
	gdi32 = syscall.NewLazyDLL("gdi32.dll")

	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pGetDIBits              = gdi32.NewProc("GetDIBits")
)

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

func CreateCompatibleDC(dc uintptr) uintptr {
	r, _, _ := pCreateCompatibleDC.Call(dc)
	return r
}

func DeleteDC(dc uintptr) bool {
	r, _, _ := pDeleteDC.Call(dc)
	return r != 0
}

func CreateCompatibleBitmap(dc uintptr, w, h int) uintptr {
	r, _, _ := pCreateCompatibleBitmap.Call(dc, uintptr(w), uintptr(h))
	return r
}

func DeleteObject(h uintptr) bool {
	r, _, _ := pDeleteObject.Call(h)
	return r != 0
}

func SelectObject(hdc, h uintptr) uintptr {
	r, _, _ := pSelectObject.Call(hdc, h)
	return r
}

func GetDIBits(hdc, hbm uintptr, start, lines uint, bits *byte, bmi *BITMAPINFOHEADER, usage uint) int {
	bmi.BiSize = uint32(unsafe.Sizeof(*bmi))
	r, _, _ := pGetDIBits.Call(hdc, hbm, uintptr(start), uintptr(lines), uintptr(unsafe.Pointer(bits)), uintptr(unsafe.Pointer(bmi)), uintptr(usage))
	return int(r)
}
