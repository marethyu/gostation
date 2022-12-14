package main

const (
	VRAM_WIDTH  = 1024
	VRAM_HEIGHT = 512
	VRAM_SIZE   = VRAM_WIDTH * VRAM_HEIGHT
)

/*
vram is an array of 16 bit pixels; it can be treated as a framebuffer of size 1024x512 for development purposes
pixel format for 15bit direct display:

pixel|                                               |
bit  |0f|0e 0d 0c 0b 0a|09 08 07 06 05|04 03 02 01 00|
desc.|M |Blue          |Green         |Red           |
*/
type VRAM struct {
	buffer [VRAM_SIZE]uint16
}

func NewVRAM() *VRAM {
	return &VRAM{
		[VRAM_SIZE]uint16{},
	}
}

func (vram *VRAM) Read16(x int, y int) uint16 {
	return vram.buffer[y*VRAM_WIDTH+x]
}

func (vram *VRAM) Write16(x int, y int, data uint16) {
	vram.buffer[y*VRAM_WIDTH+x] = data
}
