package main

import "errors"

var (
	ErrIndexOutOfBounds    = errors.New("image index out of bounds")
	ErrFdibMissing         = errors.New("could not find fdib field")
	ErrImgInvalid          = errors.New("invalid image")
	ErrArchMismatch        = errors.New("architecture mismatch")
	ErrProcessNotFound     = errors.New("could not find process")
	ErrWinNotFound         = errors.New("window could not be found")
	ErrInvalidMemoryAccess = errors.New("invalid memory access")
	ErrMemoryRead          = errors.New("memory was not read correctly")
	ErrVMTRead             = errors.New("failed learning vmt layout")
	ErrInvalidObj          = errors.New("invalid object provided")
	ErrFieldNotFound       = errors.New("could not find field")
)
