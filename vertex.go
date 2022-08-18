package main

type Vertex struct {
	/* coordinates in vram (from YyyyXxxx parameters) */
	x int32 /* 0-10   X-coordinate (signed, -1024..+1023) */
	y int32 /* 16-26  Y-coordinate (signed, -1024..+1023) */

	r int32
	g int32
	b int32

	/* coordinates in texture */
	u int32
	v int32
}

func NewVertex(rawPoint uint32, rawColour uint32, rawUV uint32) *Vertex {
	return &Vertex{
		int32(ForceSignExtension16(uint16(rawPoint&0xffff), 11)),
		int32(ForceSignExtension16(uint16(rawPoint>>16), 11)),
		int32(GetRange(rawColour, 0, 8)),
		int32(GetRange(rawColour, 8, 8)),
		int32(GetRange(rawColour, 16, 8)),
		int32(GetRange(rawUV, 0, 8)),
		int32(GetRange(rawUV, 8, 8)),
	}
}
