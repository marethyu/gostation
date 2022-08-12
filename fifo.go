package main

/* For handling GP0 commands (typically rendering commands) with multiple arguments */
type FIFO struct {
	active bool       /* is FIFO actively collecting arguments? */
	done   bool       /* finished collecting all args */
	nArgs  int        /* remaining number of arguments to collect */
	shape  int        /* what shape to render when FIFO finished collecting all args */
	attr   uint8      /* rendering attributes */
	buffer [16]uint32 /* 16-word buffer */
	idx    int        /* index in buffer */
}

func NewFIFO() *FIFO {
	return &FIFO{
		false,
		false,
		0,
		0,
		0,
		[16]uint32{},
		0,
	}
}

func (fifo *FIFO) Init(shape int, attr uint8, nArgs int) {
	fifo.active = true
	fifo.done = false
	fifo.nArgs = nArgs
	fifo.shape = shape
	fifo.attr = attr
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
