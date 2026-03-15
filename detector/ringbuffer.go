package detector

// RingBuffer is a fixed-size circular buffer of float64 values
// that tracks a running sum for efficient average computation.
type RingBuffer struct {
	data  []float64
	size  int
	head  int
	count int
	sum   float64
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float64, size),
		size: size,
	}
}

// Push adds a value to the buffer, evicting the oldest value if full.
func (rb *RingBuffer) Push(v float64) {
	if rb.count == rb.size {
		rb.sum -= rb.data[rb.head]
	} else {
		rb.count++
	}

	rb.data[rb.head] = v
	rb.sum += v
	rb.head = (rb.head + 1) % rb.size
}

// Average returns the arithmetic mean of all stored values.
func (rb *RingBuffer) Average() float64 {
	if rb.count == 0 {
		return 0
	}
	return rb.sum / float64(rb.count)
}

// Full returns true once the buffer has been filled at least once.
func (rb *RingBuffer) Full() bool {
	return rb.count == rb.size
}

// Count returns the number of values currently stored.
func (rb *RingBuffer) Count() int {
	return rb.count
}