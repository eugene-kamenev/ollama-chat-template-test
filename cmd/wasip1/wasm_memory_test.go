package main

import (
	"testing"
)

func TestAlloc_Free(t *testing.T) {
	ptr := Alloc(10)
	if ptr == 0 {
		t.Errorf("Expected non-zero pointer, got zero")
	}

	Free(ptr)
}

func TestWriteString_GetString(t *testing.T) {
	s := "Hello, World!"
	ptr := WriteString(s)
	result, err := GetString(ptr)
	if err != nil {
		t.Errorf("Error getting string: %v", err)
		return
	}
	if result != s {
		t.Errorf("Expected '%s', got '%s'", s, result)
	}
}

func TestAllocIndex_FreeIndex(t *testing.T) {
	size := int32(10)
	index := AllocIndex(size)
	if index < 0 || index >= circularBufferCapacity {
		t.Errorf("Invalid index: %d", index)
	}

	FreeIndex(index)
}

func TestCircularBuffer(t *testing.T) {
	circularCount = 0
	circularEnd = 0
	circularStart = 0
	for i := 0; i < circularBufferCapacity*2; i++ {
		index := AllocIndex(int32(1))
		if i < circularBufferCapacity && index != i {
			t.Errorf("Expected index %d, got %d", i, index)
		}
		if i >= circularBufferCapacity && index > (i%circularBufferCapacity) {
			t.Errorf("Expected index %d, got %d", i%circularBufferCapacity, index)
		}
	}
}

func TestMain(m *testing.M) {
	m.Run()
}
