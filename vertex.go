package main

type Vertex struct {
	/* coordinates in vram (from YyyyXxxx parameters) */
	x int16 /* 0-10   X-coordinate (signed, -1024..+1023) */
	y int16 /* 16-26  Y-coordinate (signed, -1024..+1023) */

	r uint8
	g uint8
	b uint8

	/* coordinates in texture */
	u uint8
	v uint8
}

func NewVertex(rawPoint uint32, rawColour uint32, rawUV uint32) *Vertex {
	return &Vertex{
		int16(ForceSignExtension16(uint16(rawPoint&0xffff), 11)),
		int16(ForceSignExtension16(uint16(rawPoint>>16), 11)),
		uint8(GetRange(rawColour, 0, 8)),
		uint8(GetRange(rawColour, 8, 8)),
		uint8(GetRange(rawColour, 16, 8)),
		uint8(GetRange(rawUV, 0, 8)),
		uint8(GetRange(rawUV, 8, 8)),
	}
}
