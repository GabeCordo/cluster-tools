package buffers

import "errors"

//////////////////////////////////////////////////////////////////////////////
//								Ring Buffer
//////////////////////////////////////////////////////////////////////////////

/* ------------------------ **** Constants ***** -------------------------- */

const MinimumRingBufferSize = 10
const MaximumRingBufferSize = 65536 // max size = 2 ^ 16

/* ------------------------ ****   Errors  ***** -------------------------- */

var InvalidRingBufferSize = errors.New("size of ring buffer is invalid")

var OutOfRoom = errors.New("the ring buffer has no room to write")

var NoData = errors.New("the ring buffer has no data to pop")

/* -------------------------- **** Types ***** ---------------------------- */

type RingBuffer struct {
	buffer   []uint64 // data held by the ring buffer.
	number   uint64
	pointers struct {
		startOfBuffer uint64 // index to the start of the ring buffer.
		endOfBuffer   uint64 // index to the end of the ring buffer.
		startOfData   uint64 // index to the start of the data in the buffer.
		endOfData     uint64 // index to the end of the data in the buffer.
	}
}

/* ------------------------ **** Functions ***** --------------------------- */

// NewRingBuffer
// allocates memory for a RingBuffer on the heap.
//
// The NewRingBuffer function may return an InvalidRingBufferSize
// error when the size is less than MinimumRingBufferSize or greater
// than MaximumRingBufferSize.
//
// Thread Safe: *yes*
// Allocates Memory: *yes*
func NewRingBuffer(size uint64) (*RingBuffer, error) {

	if (size < MinimumRingBufferSize) || (size > MaximumRingBufferSize) {
		return nil, InvalidRingBufferSize
	}

	ringBuffer := new(RingBuffer)

	ringBuffer.buffer = make([]uint64, size)
	ringBuffer.number = 0

	ringBuffer.pointers.startOfBuffer = 0
	ringBuffer.pointers.endOfBuffer = size - 1
	ringBuffer.pointers.startOfData = ringBuffer.pointers.startOfBuffer
	ringBuffer.pointers.endOfData = ringBuffer.pointers.startOfData

	return ringBuffer, nil
}

// Add
// inserts data into the ring buffer when there is sufficient room.
//
// When there is no sufficient room, the Add function shall return
// an OutOfRoom error to the callee.
//
// Thread Safe: *No*
// Allocates Memory: *No*
func (ringBuffer *RingBuffer) Add(value uint64) error {

	// does the current write index equal the read index in the buffer?
	// -> we've run out of room and cannot write to the buffer
	if (ringBuffer.pointers.startOfData == ringBuffer.pointers.endOfData) && (ringBuffer.number > 0) {
		return OutOfRoom
	}

	ringBuffer.buffer[ringBuffer.pointers.startOfData] = value
	ringBuffer.number++

	// does the current write index equal the end of the buffer?
	if ringBuffer.pointers.startOfData == ringBuffer.pointers.endOfBuffer {
		// more the write index to the start of the buffer
		ringBuffer.pointers.startOfData = ringBuffer.pointers.startOfBuffer
	} else {
		ringBuffer.pointers.startOfData = ringBuffer.pointers.startOfData + 1
	}

	return nil
}

// Remove
// removes data from the ring buffer when there is data present.
//
// When there is no data present, the Remove function shall return
// a NoData error to the callee.
//
// Thread Safe: *No*
// Allocates Memory: *No*
func (ringBuffer *RingBuffer) Remove() (value uint64, err error) {

	if ringBuffer.number == 0 {
		return 0, NoData
	}

	tmp := ringBuffer.buffer[ringBuffer.pointers.endOfData]

	if ringBuffer.pointers.endOfData == ringBuffer.pointers.endOfBuffer {
		ringBuffer.pointers.endOfData = ringBuffer.pointers.startOfData
	} else {
		ringBuffer.pointers.endOfData = ringBuffer.pointers.endOfData + 1
	}

	ringBuffer.number--
	return tmp, nil
}
