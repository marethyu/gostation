package main

/* For handling GP0 commands (typically rendering commands) with multiple arguments */
type FIFO struct {
	active bool       /* is FIFO actively collecting arguments? */
	done   bool       /* finished collecting all args */
	nArgs  int        /* remaining number of arguments to collect */
	buffer [16]uint32 /* 16-word buffer */
	idx    int        /* index in buffer */
}

func NewFIFO() *FIFO {
	return &FIFO{
		false,
		false,
		0,
		[16]uint32{},
		0,
	}
}

func (fifo *FIFO) Init(nArgs int) {
	fifo.active = true
	fifo.done = false
	fifo.nArgs = nArgs
	fifo.idx = 0
}

func (fifo *FIFO) Reset() {
	fifo.active = false
	fifo.done = false
	fifo.nArgs = 0
	fifo.idx = 0
}

func (fifo *FIFO) Push(arg uint32) {
	fifo.buffer[fifo.idx] = arg
	fifo.idx++
	fifo.done = fifo.idx == fifo.nArgs
}

func (fifo *FIFO) Done() {
	fifo.active = false
}
