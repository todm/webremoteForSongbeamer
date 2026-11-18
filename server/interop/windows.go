//go:build windows

package interop

import (
	"sync"
	"syscall"
	"unsafe"
)

// ******************************** //
// *           kernel32           * //
// ******************************** //
var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	pProcess32First           = kernel32.NewProc("Process32FirstW")
	pProcess32Next            = kernel32.NewProc("Process32NextW")
	pOpenProcess              = kernel32.NewProc("OpenProcess")
	pCloseHandle              = kernel32.NewProc("CloseHandle")
	pReadProcessMemory        = kernel32.NewProc("ReadProcessMemory")
	pIsWow64Process           = kernel32.NewProc("IsWow64Process")
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

func OpenProcess(dwDesiredAccess uintptr, bInheritHandle uintptr, dwProcessId uintptr) uintptr {
	r, _, _ := pOpenProcess.Call(dwDesiredAccess, bInheritHandle, dwProcessId)
	return r
}

func CloseHandle(hObject uintptr) {
	pCloseHandle.Call(hObject)
}

func IsWow64Process(processHandle uintptr) bool {
	var answer int32
	pIsWow64Process.Call(processHandle, uintptr(unsafe.Pointer(&answer)))
	return answer != 0
}

func ReadProcessMemory(hProcess uintptr, lpBaseAddress uintptr, nSize int, lpNumberOfBytesRead uintptr) []byte {
	buffer := make([]byte, nSize)
	pReadProcessMemory.Call(hProcess, lpBaseAddress, uintptr(unsafe.Pointer(&buffer[0])), uintptr(nSize), lpNumberOfBytesRead)
	return buffer
}

// ******************************** //
// *            user32            * //
// ******************************** //

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pEnumWindows              = user32.NewProc("EnumWindows")
	pEnumChildWindows         = user32.NewProc("EnumChildWindows")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pGetClassNameW            = user32.NewProc("GetClassNameW")
	pEnumPropsExW             = user32.NewProc("EnumPropsExW")

	pEnumWindowsProc = syscall.NewCallback(func(hWnd uintptr, lParam uintptr) uintptr {
		return boolToPtr(lookupCallback[func(hWnd uintptr) bool](lParam)(hWnd))
	})
	pEnumChildProc  = pEnumWindowsProc
	pPropEnumProcEx = syscall.NewCallback(func(p1, p2, p3, lParam uintptr) uintptr {
		return boolToPtr(lookupCallback[func(p1, p2, p3 uintptr) bool](lParam)(p1, p2, p3))
	})
)

func EnumWindows(cb func(hWnd uintptr) bool) bool {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	r, _, _ := pEnumWindows.Call(pEnumWindowsProc, cbId)
	return r != 0
}

func EnumChildWindows(hWnd uintptr, cb func(hWnd uintptr) bool) bool {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	r, _, _ := pEnumChildWindows.Call(hWnd, pEnumChildProc, cbId)
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

func EnumPropsExW(hWnd uintptr, cb func(p1, p2, p3 uintptr) bool) uintptr {
	cbId := enqueueCallback(cb)
	defer unqueueCallback(cbId)
	r, _, _ := pEnumPropsExW.Call(hWnd, pPropEnumProcEx, cbId)
	return r
}

// ******************************** //
// *           callbacks          * //
// ******************************** //

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
