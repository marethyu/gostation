package main

type Vertex struct {
	/* coordinates in vram */
	x int16
	y int16

	r uint8
	g uint8
	b uint8
}

func NewVertex(rawPoint uint32, rawColour uint32) *Vertex {
	return &Vertex{
		int16(rawPoint & 0xffff),
		int16(rawPoint >> 16),
		uint8(rawColour & 0xff),
		uint8((rawColour >> 8) & 0xff),
		uint8((rawColour >> 16) & 0xff),
	}
}
