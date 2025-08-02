package buffers

import (
	"errors"
	"testing"
)

func TestNew_AllocateRingBuffer_SizeLessThanMin(t *testing.T) {

	_, err := NewRingBuffer(MinimumRingBufferSize - 1)
	if !errors.Is(err, InvalidRingBufferSize) {
		t.Error("NewRingBuffer() should return InvalidRingBufferSize when size is less than MinimumRingBufferSize")
	}
}

func TestNew_AllocateRingBuffer_SizeMoreThanMax(t *testing.T) {

	_, err := NewRingBuffer(MaximumRingBufferSize + 1)
	if !errors.Is(err, InvalidRingBufferSize) {
		t.Error("NewRingBuffer() should return InvalidRingBufferSize when size is more than MaximumRingBufferSize")
	}
}
func TestNew_AllocateRingBuffer_SizeTen(t *testing.T) {

	var size uint64 = 10

	rb, err := NewRingBuffer(size)
	if err != nil {
		t.Error("NewRingBuffer() returned an unexpected error", err)
		return
	}

	if len(rb.buffer) != int(size) {
		t.Error("NewRingBuffer() returned an unexpected buffer size {expected: ", size, ", got:", len(rb.buffer), "}")
		return
	}

	if rb.pointers.startOfBuffer != 0 {
		t.Error("expected the startOfBuffer to point to the beginning of the buffer array")
		return
	}

	if rb.pointers.endOfBuffer != size-1 {
		t.Error("expected the endOfBuffer to point to the end of the buffer array")
		return
	}
}

func TestRingBuffer_Push_TenElements(t *testing.T) {

	var size uint64 = 10

	rb, err := NewRingBuffer(size)
	if err != nil {
		t.Error("NewRingBuffer() returned an unexpected error", err)
		return
	}

	for i := 0; i < int(size); i++ {
		err = rb.Add(uint64(i + 1))
		if err != nil {
			t.Error("Add() returned an unexpected error", err)
			return
		}
	}

	if rb.pointers.startOfData != 0 {
		t.Error("expected the startOfData to point to the beginning of the buffer array")
		return
	}

	for i := 0; i < int(size); i++ {
		_, err = rb.Remove()
		if err != nil {
			t.Error("Remove() returned an unexpected error", err)
			return
		}
	}

	if (rb.pointers.startOfData != rb.pointers.endOfData) && (rb.pointers.endOfData != 0) {
		t.Error("expected the startOfData and endOfData pointers to match")
	}
}

func TestRingBuffer_Push_ElementOnFullRing(t *testing.T) {

	rb, err := NewRingBuffer(MinimumRingBufferSize)
	if err != nil {
		t.Error("NewRingBuffer() returned an unexpected error", err)
		return
	}

	var irrelevantValue uint64 = 0
	for i := 0; i < MinimumRingBufferSize; i++ {
		err = rb.Add(irrelevantValue)
		if err != nil {
			t.Error("Add() returned an unexpected error", err)
			return
		}
	}

	err = rb.Add(irrelevantValue)
	if !errors.Is(err, OutOfRoom) {
		t.Error("Add() should return an OutOfRoom error")
	}
}

func TestRingBuffer_Remove_NoElementsInRing(t *testing.T) {

	rb, err := NewRingBuffer(MinimumRingBufferSize)
	if err != nil {
		t.Error("NewRingBuffer() returned an unexpected error", err)
		return
	}

	_, err = rb.Remove()
	if !errors.Is(err, NoData) {
		t.Error("Remove() should return NoData error")
	}
}

func TestRingBuffer_Remove_AddAndRemoveElement(t *testing.T) {

	rb, err := NewRingBuffer(MinimumRingBufferSize)
	if err != nil {
		t.Error("NewRingBuffer() returned an unexpected error", err)
		return
	}

	var irrelevantValue uint64 = 0
	if err = rb.Add(irrelevantValue); err != nil {
		t.Error("Add() returned an unexpected error", err)
		return
	}

	if rb.size != 1 {
		t.Error("Add() should have incremented the amount of data in the RingBuffer by 1")
		return
	}

	_, err = rb.Remove()
	if err != nil {
		t.Error("Remove() returned an unexpected error", err)
		return
	}

	if rb.size != 0 {
		t.Error("Remove() should have decremented the amount of data in the RingBuffer by 1")
		return
	}

	if (rb.pointers.startOfData != rb.pointers.endOfData) && (rb.pointers.endOfData != 1) {
		t.Error("expected the startOfData and endOfData pointers to match")
	}
}
