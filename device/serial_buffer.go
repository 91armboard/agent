package device

import "sync"

type serialRingBuffer struct {
	mu       sync.Mutex
	data     []byte
	head     int
	tail     int
	length   int
	capacity int
}

func newSerialRingBuffer(capacity int) *serialRingBuffer {
	if capacity <= 0 {
		capacity = 1024
	}
	return &serialRingBuffer{
		data:     make([]byte, capacity),
		capacity: capacity,
	}
}

func (r *serialRingBuffer) Push(data []byte) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(data) == 0 {
		return true
	}
	if len(data) > r.capacity-r.length {
		return false
	}
	for _, b := range data {
		r.data[r.tail] = b
		r.tail = (r.tail + 1) % r.capacity
		r.length++
	}
	return true
}

func (r *serialRingBuffer) PopAll() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.length == 0 {
		return nil
	}
	out := make([]byte, r.length)
	for i := range out {
		out[i] = r.data[r.head]
		r.head = (r.head + 1) % r.capacity
	}
	r.length = 0
	r.tail = r.head
	return out
}

func (r *serialRingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.length
}

func (r *serialRingBuffer) Capacity() int {
	return r.capacity
}
