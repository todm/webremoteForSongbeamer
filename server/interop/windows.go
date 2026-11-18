//go:build windows

package interop

import (
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
)

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

func EnumPropsExW(hWnd uintptr, lpEnumFunc uintptr, lParam uintptr) uintptr {
	r, _, _ := pEnumPropsExW.Call(hWnd, lpEnumFunc, lParam)
	return r
}
