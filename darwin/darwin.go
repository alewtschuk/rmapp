//go:build darwin

// + build darwin
// This Go file contains all the Go code to bridge MacOS native APIs from darwin.mm
package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>
#include <stdbool.h>
#include "darwin.mm"

extern bool MoveToTrash(const char *path);
extern long long GetFileAllocatedSize(const char *path);
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

// GetAllocatedFileSize returns the actual disk usage of the file in bytes
func GetDiskUsageAtPath(path string) int64 {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	return int64(C.GetDiskUsageAtPath(cPath))
}
