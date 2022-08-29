package main

const FIFO_MAX_SIZE = 16

type FIFO[T any] struct {
	buffer  [FIFO_MAX_SIZE]T
	maxSize int /* should be less than 16 */
	head    int
	tail    int
}

func NewFIFO[T any]() *FIFO[T] {
	return &FIFO[T]{
		[FIFO_MAX_SIZE]T{},
		FIFO_MAX_SIZE,
		0,
		0,
	}
}

func (fifo *FIFO[T]) Reset(maxSize int) {
	fifo.buffer = [16]T{}
	fifo.maxSize = maxSize
	fifo.head = 0
	fifo.tail = 0
}

func (fifo *FIFO[T]) Push(data T) {
	if fifo.tail == FIFO_MAX_SIZE {
		return
	}

	fifo.buffer[fifo.tail] = data
	fifo.tail += 1
}

func (fifo *FIFO[T]) Pop() T {
	data := fifo.buffer[fifo.head]
	fifo.head += 1
	return data
}

func (fifo *FIFO[T]) Front() T {
	return fifo.buffer[fifo.head]
}

func (fifo *FIFO[T]) Done() bool {
	return fifo.tail == fifo.maxSize
}

/* call this before using FIFO::Pop */
func (fifo *FIFO[T]) Empty() bool {
	return fifo.head == fifo.tail
}

func (fifo *FIFO[T]) Size() int {
	return fifo.tail - fifo.head
}
