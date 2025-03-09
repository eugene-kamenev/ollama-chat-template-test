package main

import (
	"sync"
	"unsafe"
	"errors"
)

type bufferEntry struct {
	ptr uintptr
	buf[]byte
}

const circularBufferCapacity = 10

var (
	circularBuffer [circularBufferCapacity]bufferEntry
	circularCount int
	circularStart int
	circularEnd int
	bufferMutex sync.Mutex
)

//go:wasmexport Alloc
func Alloc(size int32) uintptr {
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	// Overwrite the current slot; if an old buffer exists there, it is automatically released.
	Add(bufferEntry{ptr: ptr, buf: buf})

	return ptr
}

//go:wasmexport Free
func Free(ptr uintptr) {
	bufferMutex.Lock()
	for i := range circularBufferCapacity {
		if circularBuffer[i].ptr == ptr {
			// Clear the entry so the buffer can be garbage collected.
			circularBuffer[i] = bufferEntry{}
			break
		}
	}
	bufferMutex.Unlock()
}

func AllocIndex(size int32) int {
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	Add(bufferEntry{ptr: ptr, buf: buf})
	return circularEnd
}

func FreeIndex(index int) {
	bufferMutex.Lock()
	circularBuffer[index] = bufferEntry{}
	bufferMutex.Unlock()
}

func WriteString(s string) uintptr {
	// Calculate required size. For Go strings, len(s) returns the number of bytes.
	size := int32(len(s))
	// Allocate memory for the string.
	index := AllocIndex(size)
	// Get a slice to the allocated memory.
	outBytes := circularBuffer[index]
	// Copy the string’s bytes into the allocated buffer.
	copy(outBytes.buf, s)

	return outBytes.ptr
}

func GetString(p uintptr) (string, error) {
	for i := range circularBufferCapacity {
		if circularBuffer[i].ptr == p {
			// Clear the entry so the buffer can be garbage collected.
			return string(circularBuffer[i].buf), nil
		}
	}
	return "", errors.New("not found")
}

func GetStringIndex(index int) (string, error) {
	outBuf := circularBuffer[index]
	if outBuf.buf != nil {
		return string(outBuf.buf), nil
	}
	return "", errors.New("not found")
}

func Add(entry bufferEntry) {
	bufferMutex.Lock()
	if circularCount >= circularBufferCapacity {
		circularEnd = (circularEnd + 1) % circularBufferCapacity;
		circularStart = (circularStart + 1) % circularBufferCapacity;	
	} else {
		circularEnd = circularCount
	}
	circularBuffer[circularEnd] = entry
	circularCount++
	bufferMutex.Unlock()
}
