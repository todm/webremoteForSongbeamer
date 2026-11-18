//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/todm/webremoteForSongbeamer/interop"
)

const EXECUTABLE = "SongBeamer.exe"
const WINDOW_CLASS = "TPresentationThumbNails"
const pointersize = unsafe.Sizeof(uintptr(0))

type VMTLayout struct {
	selfCheck   uintptr
	fieldTable  uintptr
	className   uintptr
	parentClass uintptr
}

type TProcess struct {
	id uintptr
	h  uintptr

	vmtSelfCheck   uintptr
	vmtFieldTable  uintptr
	vmtClassName   uintptr
	vmtParentClass uintptr
}

func GetThumbnail(index int) (*image.RGBA, error) {
	p, err := OpenProcessByName(EXECUTABLE)
	if err != nil {
		return nil, err
	}
	defer p.Close()

	hWnd := p.GetWindowHandle(WINDOW_CLASS)
	if hWnd == 0 {
		return nil, errors.New("could not find thumbnails window")
	}
	wndObj := getWindowObject(hWnd)
	if wndObj == 0 {
		return nil, errors.New("could not get window object")
	}

	err = p.LearnVMT(wndObj)
	if err != nil {
		return nil, err
	}

	collection := p.GetObjFieldByType(wndObj, "TCollection")
	list := p.GetObjFieldByName(collection, "FItems")
	count := int(p.GetObjFieldByName(list, "FCount")) & 0xFFFFFFFF
	row := p.GetObjFieldByName(list, "FItems")

	if index < 0 || index >= count {
		return nil, errors.New("image index out of bounds")
	}

	thumbnail := p.ReadAddress(row + uintptr(index*8))
	bmpObj := p.GetObjFieldByType(thumbnail, "TBitmap")
	bmpImg := p.GetObjFieldByName(bmpObj, "FImage")
	info := p.ReadMemory(bmpImg+p.FindObjFieldByName(bmpImg, "FDIB"), 44)

	width := int(int32(binary.LittleEndian.Uint32(info[4:])))
	height := int(int32(binary.LittleEndian.Uint32(info[8:])))
	bytesPerRow := int(int32(binary.LittleEndian.Uint32(info[12:])))
	bytesPerPixel := int(binary.LittleEndian.Uint16(info[18:])) / 8
	pixelsAddress := binary.LittleEndian.Uint64(info[24:])
	upsideDown := int32(binary.LittleEndian.Uint32(info[40:])) > 0

	if pixelsAddress == 0 || width <= 0 || height <= 0 || bytesPerPixel < 3 {
		return nil, errors.New("invalid image")
	}
	pixels := p.ReadMemory(uintptr(pixelsAddress), bytesPerRow*height)
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

func OpenProcessByName(name string) (*TProcess, error) {
	snapshot := interop.CreateToolhelp32Snapshot(0x02, 0) //TH32CS_SNAPPROCESS
	defer syscall.CloseHandle(syscall.Handle(snapshot))
	if snapshot == 0 {
		return nil, errors.New("could not open toolhelp snapshot")
	}

	var pid uintptr
	var p interop.PROCESSENTRY32
	for ok := interop.Process32First(snapshot, &p); ok; ok = interop.Process32Next(snapshot, &p) {
		if strings.EqualFold(syscall.UTF16ToString(p.SzExeFile[:]), name) {
			pid = uintptr(p.Th32ProcessID)
			break
		}
	}

	if pid == 0 {
		return nil, errors.New("could not find process")
	}

	handle := interop.OpenProcess(0x0010|0x0400, 0, pid)

	if handle == 0 {
		return nil, errors.New("could not open process")
	}

	// check architecture mismatch
	selfHandle, _ := syscall.GetCurrentProcess()
	if interop.IsWow64Process(uintptr(selfHandle)) != interop.IsWow64Process(handle) {
		return nil, errors.New("architecture mismatch")
	}

	return &TProcess{
		id: pid,
		h:  handle,
	}, nil
}

func (p *TProcess) Close() {
	defer interop.CloseHandle(p.h)
}

func (p *TProcess) GetWindowHandle(className string) uintptr {
	var handle uintptr
	var callback uintptr
	buffer := make([]uint16, 256)

	callback = syscall.NewCallback(func(hWnd uintptr, _ uintptr) uintptr {
		windowProc := interop.GetWindowThreadProcessId(hWnd)
		if uintptr(windowProc) != p.id {
			return 1
		}
		copied := interop.GetClassName(hWnd, buffer)
		if strings.EqualFold(syscall.UTF16ToString(buffer[:copied]), className) {
			handle = hWnd
			return 0
		}
		interop.EnumChildWindows(hWnd, callback, unsafe.Pointer(nil))
		return 1
	})
	interop.EnumWindows(callback, unsafe.Pointer(nil))
	return uintptr(handle)
}

func (p *TProcess) ReadMemory(address uintptr, size int) []byte {
	return interop.ReadProcessMemory(p.h, address, size, 0)
}

func (p *TProcess) ReadAddress(address uintptr) uintptr {
	data := p.ReadMemory(address, int(pointersize))
	if pointersize == 8 {
		return uintptr(binary.LittleEndian.Uint64(data))
	} else {
		return uintptr(binary.LittleEndian.Uint32(data))
	}
}

func (p *TProcess) LearnVMT(obj uintptr) error {
	class := p.ReadAddress(obj)
	if class == 0 {
		return nil
	}
	for i := uintptr(10); i <= 40; i++ {
		offset := i * pointersize
		if p.ReadAddress(class-offset) != class {
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
	for class := p.ReadAddress(obj); class != 0; class = p.ReadAddress(p.ReadAddress(class - p.vmtParentClass)) {
		fieldTableAddress := p.ReadAddress(class - p.vmtFieldTable)
		if fieldTableAddress == 0 {
			continue
		}

		fieldTable := p.ReadMemory(fieldTableAddress, 4096)
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

func (p *TProcess) FindObjFieldByName(obj uintptr, name string) uintptr {
	var field uintptr
	p.EnumObjFields(obj, func(fname string, address uintptr) bool {
		if fname == name {
			field = address
			return false
		}
		return true
	})

	return field
}

func (p *TProcess) GetObjFieldByName(obj uintptr, name string) uintptr {
	return p.ReadAddress(obj + p.FindObjFieldByName(obj, name))
}

func (p *TProcess) GetObjFieldByType(obj uintptr, className string) uintptr {
	var field uintptr
	p.EnumObjFields(obj, func(fname string, address uintptr) bool {
		v := p.ReadAddress(obj + address)
		for class := p.ReadAddress(v); class != 0; class = p.ReadAddress(p.ReadAddress(class - p.vmtParentClass)) {
			nameAddress := p.ReadAddress(class - p.vmtClassName)
			nameLength := int(p.ReadMemory(nameAddress, 1)[0])
			name := string(p.ReadMemory(nameAddress+1, nameLength))
			if name == className {
				field = v
				return false
			}
		}
		return true
	})
	return field
}

func getWindowObject(window uintptr) uintptr {
	var callback uintptr
	var objPtr uintptr

	callback = syscall.NewCallback(func(_, key, val, _ uintptr) uintptr {
		if key > 0xFFFF && strings.HasPrefix(utf16PtrToString(key), "ControlOfs") {
			objPtr = val
			return 0
		}
		return 1
	})
	interop.EnumPropsExW(window, callback, 0)
	return objPtr
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
