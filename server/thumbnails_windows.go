//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/todm/webremoteForSongbeamer/interop"
)

const EXECUTABLE = "SongBeamer.exe"
const WINDOW_CLASS = "TPresentationThumbNails"
const IMG_MAX = 4096

const pointersize = unsafe.Sizeof(uintptr(0))
const bmBitsOffset = 4*4 + 2*2 + (pointersize-4)%pointersize
const bmihOffset = bmBitsOffset + pointersize

type VMTLayout struct {
	selfCheck   uintptr
	fieldTable  uintptr
	className   uintptr
	parentClass uintptr
}

type TProcess struct {
	id   uintptr
	h    uintptr
	hWnd uintptr

	vmtSelfCheck   uintptr
	vmtFieldTable  uintptr
	vmtClassName   uintptr
	vmtParentClass uintptr
}

var process *TProcess
var processMu sync.Mutex

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
		return nil, errors.New("image index out of bounds")
	}

	thumbnail, ok := p.ReadAddress(row + uintptr(index)*pointersize)
	if !ok {
		return nil, errors.New("could not get thumbnail")
	}

	bmpObj, err := p.GetObjFieldByType(thumbnail, "TBitmap")
	if err != nil {
		return nil, err
	}
	bmpImg, err := p.GetObjFieldByName(bmpObj, "FImage")
	if err != nil || !p.IsObject(bmpImg) {
		return nil, fmt.Errorf("%w could not find image", err)
	}
	fdib, ok := p.FindObjFieldByName(bmpImg, "FDIB")
	if !ok {
		return nil, errors.New("could not find fdib field")
	}

	info, ok := p.ReadMemory(bmpImg+fdib, int(bmihOffset+12))
	if !ok {
		return nil, errors.New("could not get bitmap info")
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
		width > IMG_MAX || height > IMG_MAX ||
		(bytesPerPixel != 3 && bytesPerPixel != 4) ||
		bytesPerRow < width*bytesPerPixel {
		return nil, errors.New("invalid image")
	}
	pixels, ok := p.ReadMemory(pixelsAddress, bytesPerRow*height)
	if !ok {
		return nil, errors.New("could not read pixel data")
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
	pid, err := GetProcessByName(EXECUTABLE)
	if err != nil {
		return nil, err
	}

	handle := interop.OpenProcess(0x00100000|0x0010|0x0400, 0, pid) // SYNCHRONIZE | PROCESS_VM_READ | PROCESS_QUERY_INFORMATION

	if handle == 0 {
		return nil, errors.New("could not open process")
	}

	// check architecture mismatch
	selfHandle, _ := syscall.GetCurrentProcess()
	if interop.IsWow64Process(uintptr(selfHandle)) != interop.IsWow64Process(handle) {
		interop.CloseHandle(handle)
		return nil, errors.New("architecture mismatch")
	}

	p := &TProcess{
		id: pid,
		h:  handle,
	}

	hWnd, err := p.GetWindowHandle(WINDOW_CLASS)
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

func GetProcessByName(name string) (uintptr, error) {
	snapshot := interop.CreateToolhelp32Snapshot(0x02, 0) //TH32CS_SNAPPROCESS
	if snapshot == 0 {
		return 0, errors.New("could not open toolhelp snapshot")
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var pid uintptr
	var p interop.PROCESSENTRY32
	for ok := interop.Process32First(snapshot, &p); ok; ok = interop.Process32Next(snapshot, &p) {
		if strings.EqualFold(syscall.UTF16ToString(p.SzExeFile[:]), name) {
			pid = uintptr(p.Th32ProcessID)
			break
		}
	}

	if pid == 0 {
		return 0, fmt.Errorf("could not find process '%s'", name)
	}
	return pid, nil
}

func (p *TProcess) IsAlive() bool {
	if p == nil || p.h == 0 || p.hWnd == 0 {
		return false
	}
	if interop.WaitForSingleObject(p.h, 0) != 0x00000102 { // WAIT_TIMEOUT
		return false
	}
	if !interop.IsWindow(p.hWnd) || interop.GetWindowThreadProcessId(p.hWnd) != uint32(p.id) {
		return false
	}
	return true
}

func (p *TProcess) Close() {
	interop.CloseHandle(p.h)
}

func (p *TProcess) GetWindowHandle(className string) (uintptr, error) {
	var handle uintptr

	buffer := make([]uint16, 256)
	checkClassName := func(hWnd uintptr) bool {
		copied := interop.GetClassName(hWnd, buffer)
		return strings.EqualFold(syscall.UTF16ToString(buffer[:copied]), className)
	}

	interop.EnumWindows(func(hWnd uintptr) bool {
		if interop.GetWindowThreadProcessId(hWnd) != uint32(p.id) {
			return true
		}
		if checkClassName(hWnd) {
			handle = hWnd
			return false
		}

		interop.EnumChildWindows(hWnd, func(hWndChild uintptr) bool {
			if checkClassName(hWndChild) {
				handle = hWndChild
				return false
			}
			return true
		})

		return handle == 0
	})
	if handle == 0 {
		return 0, fmt.Errorf("could not find window with class '%s'", className)
	}
	return handle, nil
}

func (p *TProcess) ReadMemory(address uintptr, size int) ([]byte, bool) {
	if size < 0 || address == 0 {
		return nil, false
	}
	buffer := make([]byte, size)
	ok := interop.ReadProcessMemory(p.h, address, buffer)
	return buffer, ok
}

func (p *TProcess) ReadAddress(address uintptr) (uintptr, bool) {
	data, ok := p.ReadMemory(address, int(pointersize))
	if !ok {
		return 0, false
	}
	if len(data) == 8 {
		return uintptr(binary.LittleEndian.Uint64(data)), true
	} else {
		return uintptr(binary.LittleEndian.Uint32(data)), true
	}
}

func (p *TProcess) LearnVMT(obj uintptr) error {
	class, ok := p.ReadAddress(obj)
	if !ok || class == 0 {
		return errors.New("faild learning vmt layout")
	}
	for i := uintptr(10); i <= 40; i++ {
		offset := i * pointersize
		if self, ok := p.ReadAddress(class - offset); !ok || self != class {
			continue
		}
		p.vmtSelfCheck = offset
		p.vmtFieldTable = offset - 5*pointersize
		p.vmtClassName = offset - 8*pointersize
		p.vmtParentClass = offset - 10*pointersize
		return nil
	}
	return errors.New("faild learning vmt layout")
}

func (p *TProcess) EnumObjFields(obj uintptr, cb func(name string, address uintptr) bool) {
	for class, ok := p.ReadAddress(obj); ok && class != 0; class = p.GetClassParent(class) {
		fieldTableAddress, ok := p.ReadAddress(class - p.vmtFieldTable)
		if fieldTableAddress == 0 || !ok {
			continue
		}

		fieldTable, ok := p.ReadMemory(fieldTableAddress, 4096)
		if !ok {
			continue
		}

		byteOffset := 0
		// Section 1
		count := int(binary.LittleEndian.Uint16(fieldTable[byteOffset:]))
		byteOffset += 10
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
			position := binary.LittleEndian.Uint32(fieldTable[byteOffset+9:])
			nameLength := int(fieldTable[byteOffset+13])
			name := string(fieldTable[byteOffset+14 : byteOffset+14+nameLength])
			if !cb(name, uintptr(position)) {
				return
			}
			byteOffset += 14 + nameLength
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
		return 0, errors.New("invalid object provided")
	}

	field, ok := p.FindObjFieldByName(obj, name)
	if !ok {
		return 0, fmt.Errorf("could not find filed with name %s", name)
	}
	address, ok := p.ReadAddress(obj + field)
	if !ok {
		return 0, fmt.Errorf("could not read field with name %s", name)
	}
	return address, nil
}

func (p *TProcess) GetObjFieldByType(obj uintptr, className string) (uintptr, error) {
	if obj == 0 {
		return 0, errors.New("invalid object provided")
	}

	var field uintptr
	p.EnumObjFields(obj, func(fname string, address uintptr) bool {
		v, ok := p.ReadAddress(obj + address)
		if !ok || !p.IsObject(v) {
			return true
		}

		for class, ok := p.ReadAddress(v); class != 0 && ok; class = p.GetClassParent(class) {
			nameAddress, ok := p.ReadAddress(class - p.vmtClassName)
			if !ok {
				continue
			}
			nameLength, ok := p.ReadMemory(nameAddress, 1)
			if !ok {
				continue
			}
			nameBytes, ok := p.ReadMemory(nameAddress+1, int(nameLength[0]))
			if !ok {
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
		return 0, fmt.Errorf("could not find field of type %s", className)
	}
	return field, nil
}

func (p *TProcess) IsClass(class uintptr) bool {
	self, ok := p.ReadAddress(class - p.vmtSelfCheck)
	return ok && self == class
}

func (p *TProcess) IsObject(obj uintptr) bool {
	class, ok := p.ReadAddress(obj)
	return ok && p.IsClass(class)
}

func (p *TProcess) GetClassParent(class uintptr) uintptr {
	parentAddress, ok := p.ReadAddress(class - p.vmtParentClass)
	if !ok || parentAddress == 0 {
		return 0
	}
	parent, ok := p.ReadAddress(parentAddress)
	if !ok || parent == 0 {
		return 0
	}

	if !p.IsClass(parent) {
		return 0
	}

	return parent
}

func getWindowObject(window uintptr) (uintptr, error) {
	var objPtr uintptr

	interop.EnumPropsExW(window, func(_, key, val uintptr) bool {
		if key > 0xFFFF && strings.HasPrefix(utf16PtrToString(key), "ControlOfs") {
			objPtr = val
			return false
		}
		return true
	})
	if objPtr == 0 {
		return 0, errors.New("could not get window object")
	}
	return objPtr, nil
}

func utf16PtrToString(p uintptr) string {
	var s []uint16
	for ; ; p += 2 {
		c := *(*uint16)(unsafe.Pointer(p))
		if c == 0 {
			return string(utf16.Decode(s))
		}
		s = append(s, c)
	}
}
