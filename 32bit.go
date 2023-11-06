package tipe

import "unsafe"

// tipe only supports 64bit platform for now.
// TODO: support 32bit platform. (maybe never)
type _ [unsafe.Sizeof(uintptr(0)) - 8]struct{}
