package main

const (
	VRAM_WIDTH  = 1024
	VRAM_HEIGHT = 512
	VRAM_SIZE   = VRAM_WIDTH * VRAM_HEIGHT
)

/* vram is an array of 16 bit pixels; it can be treated as a framebuffer of size 1024x512 for development purposes */
type VRAM struct {
	buffer [VRAM_SIZE]uint16
}

func NewVRAM() *VRAM {
	return &VRAM{
		[VRAM_SIZE]uint16{},
	}
}

func (vram *VRAM) Read16(x uint32, y uint32) uint16 {
	return vram.buffer[y*VRAM_WIDTH+x]
}

/* TODO 4 and 8 bit mode writes? */
func (vram *VRAM) Write16(x uint32, y uint32, data uint16) {
	vram.buffer[y*VRAM_WIDTH+x] = data
}
