package memory

import (
	"crypto/md5"
	"fmt"
	"unsafe"
)

var objects = make([]monitoredObject, 0)

type monitoredObject struct {
	label           string
	ptr             uintptr
	size            uintptr
	initialChecksum string
}

func Monitor(label string, ptr uintptr, size uintptr) {
	obj := monitoredObject{
		ptr:             ptr,
		size:            size,
		initialChecksum: ChecksumArbitraryUintptr(ptr, size),
	}

	objects = append(objects, obj)
}

func CheckMem() {
	for _, obj := range objects {
		currentChecksum := ChecksumArbitraryUintptr(obj.ptr, obj.size)
		if currentChecksum != obj.initialChecksum {
			fmt.Println("Mem changed", obj.label)
		}
	}
}

func ChecksumArbitraryUintptr(ptr uintptr, size uintptr) string {
	return ChecksumArbitraryPointer(unsafe.Pointer(ptr), size)
}

func ChecksumArbitraryPointer(ptr unsafe.Pointer, size uintptr) string {
	data := ReadArbitraryMemory(ptr, size)
	if len(data) == 0 {
		return "nil"
	}

	checksum := fmt.Sprintf("%x", md5.Sum(data))
	return checksum
}

func ReadArbitraryMemory(ptr unsafe.Pointer, size uintptr) (data []byte) {
	defer func() {
		if r := recover(); r != nil {
			data = nil
		}
	}()

	data = make([]byte, size)
	for i := range data {
		data[i] = *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i)))
	}

	return data
}
