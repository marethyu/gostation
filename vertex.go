package main

type Vertex struct {
	/* coordinates in vram (from YyyyXxxx parameters) */
	x int /* 0-10   X-coordinate (signed, -1024..+1023) */
	y int /* 16-26  Y-coordinate (signed, -1024..+1023) */

	r int
	g int
	b int

	/* coordinates in texture */
	u int
	v int
}

func NewVertex(rawPoint uint32, rawColour uint32, rawUV uint32, xOffset int, yOffset int) *Vertex {
	return &Vertex{
		int(ForceSignExtension16(uint16(rawPoint&0xffff), 11)) + xOffset,
		int(ForceSignExtension16(uint16(rawPoint>>16), 11)) + yOffset,
		int(GetRange(rawColour, 0, 8)),
		int(GetRange(rawColour, 8, 8)),
		int(GetRange(rawColour, 16, 8)),
		int(GetRange(rawUV, 0, 8)),
		int(GetRange(rawUV, 8, 8)),
	}
}
