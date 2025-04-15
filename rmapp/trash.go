//go:build darwin

//+ build darwin

package rmapp

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>
#include <stdbool.h>
#include "trash_darwin.mm"

extern bool MoveToTrash(const char *path);
*/
import "C"
import "unsafe"

// MoveFileToTrash attempts to use CGo and unsafe to move a file to trash
// by interfacing with native NSFileManager API
func MoveFileToTrash(path string) bool {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	return bool(C.MoveToTrash(cPath))
}
