//go:build windows

package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	sbExecutable  = "SongBeamer.exe"
	sbWindowClass = "TPresentationThumbNails"
	imgMax        = 4096

	pointersize  = unsafe.Sizeof(uintptr(0))
	bmBitsOffset = 4*4 + 2*2 + (pointersize-4)%pointersize
	bmihOffset   = bmBitsOffset + pointersize
)

type TProcess struct {
	id   uint32
	h    windows.Handle
	hWnd windows.HWND

	vmtSelfCheck   uintptr
	vmtFieldTable  uintptr
	vmtClassName   uintptr
	vmtParentClass uintptr
}

var (
	process   *TProcess
	processMu sync.Mutex
)

func GetThumbnail(index int) (*image.RGBA, error) {
	processMu.Lock()
	defer processMu.Unlock()

	var err error
	if process == nil || !process.IsAlive() {
		if process != nil {
			process.Close()
		}
		process, err = GetProcess()
		if err != nil {
			return nil, err
		}
	}

	p := process

	wndObj, err := getWindowObject(p.hWnd)
	if err != nil {
		return nil, err
	}

	collection, err := p.GetObjFieldByType(wndObj, "TCollection")
	if err != nil {
		return nil, err
	}
	list, err := p.GetObjFieldByName(collection, "FItems")
	if err != nil {
		return nil, err
	}
	row, err := p.GetObjFieldByName(list, "FItems")
	if err != nil {
		return nil, err
	}
	countRaw, err := p.GetObjFieldByName(list, "FCount")
	if err != nil {
		return nil, err
	}

	count := int(int32(countRaw))
	if index < 0 || index >= count {
		return nil, ErrIndexOutOfBounds
	}

	thumbnail, err := p.ReadAddress(row + uintptr(index)*pointersize)
	if err != nil {
		return nil, err
	}

	bmpObj, err := p.GetObjFieldByType(thumbnail, "TBitmap")
	if err != nil {
		return nil, err
	}
	bmpImg, err := p.GetObjFieldByName(bmpObj, "FImage")
	if err != nil {
		return nil, err
	}
	if !p.IsObject(bmpImg) {
		return nil, ErrImgInvalid
	}
	fdib, ok := p.FindObjFieldByName(bmpImg, "FDIB")
	if !ok {
		return nil, ErrFdibMissing
	}

	info, err := p.ReadMemory(bmpImg+fdib, int(bmihOffset+12))
	if err != nil {
		return nil, err
	}

	width := int(int32(binary.LittleEndian.Uint32(info[4:])))
	height := int(int32(binary.LittleEndian.Uint32(info[8:])))
	bytesPerRow := int(int32(binary.LittleEndian.Uint32(info[12:])))
	bitsPerPixel := int(binary.LittleEndian.Uint16(info[18:]))
	upsideDown := int32(binary.LittleEndian.Uint32(info[bmihOffset+8:])) > 0

	var pixelsAddress uintptr
	if pointersize == 8 {
		pixelsAddress = uintptr(binary.LittleEndian.Uint64(info[bmBitsOffset:]))
	} else {
		pixelsAddress = uintptr(binary.LittleEndian.Uint32(info[bmBitsOffset:]))
	}

	bytesPerPixel := bitsPerPixel / 8
	if pixelsAddress == 0 ||
		width <= 0 || height <= 0 ||
		width > imgMax || height > imgMax ||
		(bytesPerPixel != 3 && bytesPerPixel != 4) ||
		bytesPerRow < width*bytesPerPixel {
		return nil, ErrImgInvalid
	}
	pixels, err := p.ReadMemory(pixelsAddress, bytesPerRow*height)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		row := y
		if upsideDown {
			row = height - 1 - y
		}
		for x := 0; x < width; x++ {
			i := row*bytesPerRow + x*bytesPerPixel
			img.SetRGBA(x, y, color.RGBA{R: pixels[i+2], G: pixels[i+1], B: pixels[i], A: 255})
		}
	}

	return img, nil
}

func GetProcess() (*TProcess, error) {
	pid, err := GetProcessByName(sbExecutable)
	if err != nil {
		return nil, err
	}
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		return nil, err
	}

	// check architecture mismatch
	selfHandle, _ := syscall.GetCurrentProcess()
	var selfWow, handleWow bool
	err = windows.IsWow64Process(windows.Handle(selfHandle), &selfWow)
	if err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}
	err = windows.IsWow64Process(handle, &handleWow)
	if err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}

	if selfWow != handleWow {
		windows.CloseHandle(handle)
		return nil, ErrArchMismatch
	}

	p := &TProcess{
		id: pid,
		h:  handle,
	}

	hWnd, err := p.GetWindowHandle(sbWindowClass)
	if err != nil {
		p.Close()
		return nil, err
	}
	p.hWnd = hWnd

	wndObj, err := getWindowObject(hWnd)
	if err != nil {
		p.Close()
		return nil, err
	}

	err = p.LearnVMT(wndObj)
	if err != nil {
		p.Close()
		return nil, err
	}

	return p, nil
}

func GetProcessByName(name string) (uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var pid uint32
	var p windows.ProcessEntry32
	p.Size = uint32(unsafe.Sizeof(p))
	for err := windows.Process32First(snapshot, &p); err == nil; err = windows.Process32Next(snapshot, &p) {
		if strings.EqualFold(syscall.UTF16ToString(p.ExeFile[:]), name) {
			pid = p.ProcessID
			break
		}
	}

	if pid == 0 {
		return 0, ErrProcessNotFound
	}
	return pid, nil
}

func (p *TProcess) IsAlive() bool {
	if p == nil || p.h == 0 || p.hWnd == 0 {
		return false
	}
	if r, err := windows.WaitForSingleObject(p.h, 0); err != nil || r != uint32(windows.WAIT_TIMEOUT) {
		return false
	}
	if !windows.IsWindow(p.hWnd) {
		return false
	}
	return true
}

func (p *TProcess) Close() {
	windows.CloseHandle(p.h)
}

func (p *TProcess) GetWindowHandle(className string) (windows.HWND, error) {
	var handle windows.HWND

	buffer := make([]uint16, 256)
	checkClassName := func(hWnd windows.HWND) bool {
		copied, err := windows.GetClassName(hWnd, &buffer[0], int32(len(buffer)))
		if err != nil {
			return false
		}
		return strings.EqualFold(syscall.UTF16ToString(buffer[:copied]), className)
	}

	EnumWindows(func(hWnd windows.HWND) bool {
		var pid uint32
		windows.GetWindowThreadProcessId(hWnd, &pid)
		if pid != p.id {
			return true
		}
		if checkClassName(hWnd) {
			handle = hWnd
			return false
		}

		EnumChildWindows(hWnd, func(hWndChild windows.HWND) bool {
			if checkClassName(hWndChild) {
				handle = hWndChild
				return false
			}
			return true
		})

		return handle == 0
	})
	if handle == 0 {
		return 0, ErrWinNotFound
	}
	return handle, nil
}

func (p *TProcess) ReadMemory(address uintptr, size int) ([]byte, error) {
	if size <= 0 || address == 0 {
		return nil, ErrInvalidMemoryAccess
	}
	buffer := make([]byte, size)
	var bytesRead uintptr
	err := windows.ReadProcessMemory(p.h, address, &buffer[0], uintptr(size), &bytesRead)

	if err != nil || int(bytesRead) != size {
		return nil, ErrMemoryRead
	}

	return buffer, nil
}

func (p *TProcess) ReadAddress(address uintptr) (uintptr, error) {
	data, err := p.ReadMemory(address, int(pointersize))
	if err != nil {
		return 0, err
	}
	if len(data) == 8 {
		return uintptr(binary.LittleEndian.Uint64(data)), nil
	} else {
		return uintptr(binary.LittleEndian.Uint32(data)), nil
	}
}

func (p *TProcess) LearnVMT(obj uintptr) error {
	class, err := p.ReadAddress(obj)
	if err != nil || class == 0 {
		return ErrVMTRead
	}
	for i := uintptr(10); i <= 40; i++ {
		offset := i * pointersize
		if self, err := p.ReadAddress(class - offset); err != nil || self != class {
			continue
		}
		p.vmtSelfCheck = offset
		p.vmtFieldTable = offset - 5*pointersize
		p.vmtClassName = offset - 8*pointersize
		p.vmtParentClass = offset - 10*pointersize
		return nil
	}
	return ErrVMTRead
}

func (p *TProcess) EnumObjFields(obj uintptr, cb func(name string, address uintptr) bool) {
	sptr := int(pointersize)
	for class, err := p.ReadAddress(obj); err == nil && class != 0; class = p.GetClassParent(class) {
		fieldTableAddress, err := p.ReadAddress(class - p.vmtFieldTable)
		if fieldTableAddress == 0 || err != nil {
			continue
		}

		fieldTable, err := p.ReadMemory(fieldTableAddress, 4096)
		if err != nil {
			continue
		}

		byteOffset := 0
		// Section 1
		count := int(binary.LittleEndian.Uint16(fieldTable[byteOffset:]))
		byteOffset += 2 + sptr
		for i := 0; i < count; i++ {
			position := binary.LittleEndian.Uint32(fieldTable[byteOffset:])
			nameLength := int(fieldTable[byteOffset+6])
			name := string(fieldTable[byteOffset+7 : byteOffset+7+nameLength])
			if !cb(name, uintptr(position)) {
				return
			}
			byteOffset += 7 + nameLength
		}

		// Section 2
		count = int(binary.LittleEndian.Uint16(fieldTable[byteOffset:]))
		byteOffset += 2
		for i := 0; i < count; i++ {
			position := binary.LittleEndian.Uint32(fieldTable[byteOffset+1+sptr:])
			nameLength := int(fieldTable[byteOffset+5+sptr])
			name := string(fieldTable[byteOffset+6+sptr : byteOffset+6+sptr+nameLength])
			if !cb(name, uintptr(position)) {
				return
			}
			byteOffset += 6 + sptr + nameLength
			byteOffset += int(binary.LittleEndian.Uint16(fieldTable[byteOffset:]))
		}
	}
}

func (p *TProcess) FindObjFieldByName(obj uintptr, name string) (uintptr, bool) {
	var field uintptr
	p.EnumObjFields(obj, func(fname string, address uintptr) bool {
		if fname == name {
			field = address
			return false
		}
		return true
	})

	return field, field != 0
}

func (p *TProcess) GetObjFieldByName(obj uintptr, name string) (uintptr, error) {
	if obj == 0 {
		return 0, ErrInvalidObj
	}

	field, ok := p.FindObjFieldByName(obj, name)
	if !ok {
		return 0, ErrFieldNotFound
	}
	address, err := p.ReadAddress(obj + field)
	if err != nil {
		return 0, err
	}
	return address, nil
}

func (p *TProcess) GetObjFieldByType(obj uintptr, className string) (uintptr, error) {
	if obj == 0 {
		return 0, ErrInvalidObj
	}

	var field uintptr
	p.EnumObjFields(obj, func(fname string, address uintptr) bool {
		v, err := p.ReadAddress(obj + address)
		if err != nil || !p.IsObject(v) {
			return true
		}

		for class, err := p.ReadAddress(v); class != 0 && err == nil; class = p.GetClassParent(class) {
			nameAddress, err := p.ReadAddress(class - p.vmtClassName)
			if err != nil {
				continue
			}
			nameLength, err := p.ReadMemory(nameAddress, 1)
			if err != nil {
				continue
			}
			nameBytes, err := p.ReadMemory(nameAddress+1, int(nameLength[0]))
			if err != nil {
				continue
			}

			if string(nameBytes) == className {
				field = v
				return false
			}
		}
		return true
	})
	if field == 0 {
		return 0, ErrFieldNotFound
	}
	return field, nil
}

func (p *TProcess) IsClass(class uintptr) bool {
	self, err := p.ReadAddress(class - p.vmtSelfCheck)
	return err == nil && self == class
}

func (p *TProcess) IsObject(obj uintptr) bool {
	class, err := p.ReadAddress(obj)
	return err == nil && p.IsClass(class)
}

func (p *TProcess) GetClassParent(class uintptr) uintptr {
	parentAddress, err := p.ReadAddress(class - p.vmtParentClass)
	if err != nil || parentAddress == 0 {
		return 0
	}
	parent, err := p.ReadAddress(parentAddress)
	if err != nil || parent == 0 {
		return 0
	}

	if !p.IsClass(parent) {
		return 0
	}

	return parent
}

func getWindowObject(window windows.HWND) (uintptr, error) {
	var objPtr uintptr
	EnumPropsExW(window, func(_, key, val uintptr) bool {
		if key > 0xFFFF && strings.HasPrefix(utf16PtrToString(key, 12), "ControlOfs") {
			objPtr = val
			return false
		}
		return true
	})
	if objPtr == 0 {
		return 0, ErrWinNotFound
	}
	return objPtr, nil
}

func utf16PtrToString(p uintptr, max int) string {
	var s []uint16
	for i := 0; i < max; i++ {
		c := *(*uint16)(unsafe.Pointer(p + uintptr(i)*2))
		if c == 0 {
			break
		}
		s = append(s, c)
	}
	return string(utf16.Decode(s))
}

// ******************************** //
// *           callbacks          * //
// ******************************** //

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	pEnumWindows      = user32.NewProc("EnumWindows")
	pEnumChildWindows = user32.NewProc("EnumChildWindows")
	pEnumPropsExW     = user32.NewProc("EnumPropsExW")

	pEnumWindowsProc = syscall.NewCallback(func(hWnd uintptr, lParam uintptr) uintptr {
		return boolToPtr(lookupCallback[func(hWnd windows.HWND) bool](lParam)(windows.HWND(hWnd)))
	})
	pPropEnumProcEx = syscall.NewCallback(func(p1, p2, p3, lParam uintptr) uintptr {
		return boolToPtr(lookupCallback[func(p1, p2, p3 uintptr) bool](lParam)(p1, p2, p3))
	})
)

func EnumWindows(cb func(hWnd windows.HWND) bool) {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	pEnumWindows.Call(pEnumWindowsProc, cbId)
}

func EnumChildWindows(hWnd windows.HWND, cb func(hWnd windows.HWND) bool) {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	pEnumChildWindows.Call(uintptr(hWnd), pEnumWindowsProc, cbId)
}

func EnumPropsExW(hWnd windows.HWND, cb func(p1, p2, p3 uintptr) bool) {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	pEnumPropsExW.Call(uintptr(hWnd), pPropEnumProcEx, cbId)
}

var (
	cbMu   sync.Mutex
	cbNext uintptr
	cbFns  = map[uintptr]any{}
)

func enqueueCallback[T any](cb T) uintptr {
	cbMu.Lock()
	defer cbMu.Unlock()

	cbNext++
	id := cbNext
	cbFns[id] = cb
	return id
}

func lookupCallback[T any](id uintptr) T {
	cbMu.Lock()
	defer cbMu.Unlock()
	return cbFns[id].(T)
}

func unqueueCallback(id uintptr) {
	cbMu.Lock()
	defer cbMu.Unlock()
	delete(cbFns, id)
}

func boolToPtr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
