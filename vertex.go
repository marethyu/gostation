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
		uint8(rawColour & 0xff),
		uint8((rawColour >> 8) & 0xff),
		uint8((rawColour >> 16) & 0xff),
		uint8(rawUV & 0xff),
		uint8((rawUV >> 8) & 0xff),
	}
}
