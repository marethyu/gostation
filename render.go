package main

import (
	"fmt"
)

const (
	SHAPE_POLYGON = iota
)

/*
1--2    Note: A 4 point polygon is processed internally as two 3 point
|  |    polygons.
3--4    Note: When drawing a polygon the GPU will not draw the right

	most and bottom edge. So a (0,0)-(32,32) rectangle will actually

be drawn as (0,0)-(31,31). Make sure adjoining polygons have the same
coordinates if you want them to touch eachother!. Haven't checked how this
works with 3 point polygons.
*/
func (gpu *GPU) DoRenderPolygon() {
	switch gpu.shape_attr {
	case 0b10000:
		gpu.RenderTrigGouraud()
	case 0b01000:
		gpu.RenderMonochromeQuad()
	case 0b11000:
		gpu.RenderQuadGouraud()
	case 0b01100:
		gpu.RenderTexturedQuad()
	default:
		panic(fmt.Sprintf("[GPU::DoRenderPolygon] Unknown attribute: %05b\n", gpu.shape_attr))
	}
}

func (gpu *GPU) RenderTrigGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
}

func (gpu *GPU) RenderMonochromeQuad() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)
	v4 := NewVertex(gpu.fifo.buffer[4], colour, 0)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
	gpu.ShadedTriangle(v2, v4, v3)
}

func (gpu *GPU) RenderQuadGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], 0)

	// make sure that vertexes are in clockwise order
	gpu.ShadedTriangle(v1, v2, v3)
	gpu.ShadedTriangle(v2, v4, v3)
}

/*
Example argument list:

2CR1G1B1
Yyy1Xxx1
ClutU1V1
00R2G2B2
Yyy2Xxx2
PageU2V2
00R3G3B3
Yyy3Xxx3
0000U3V3
00R4G4B4
Yyy4Xxx4
0000U4V4

Format for the 16 bit 'Clut':

	0-5      X coordinate X/16  (ie. in 16-halfword steps)
	6-14     Y coordinate 0-511 (ie. in 1-line steps)
	15       Unknown/unused (should be 0)

Format for the 16 bit 'Page':

	0-8    Same as GP0(E1h).Bit0-8 (see there)
	9-10   Unused (does NOT change GP0(E1h).Bit9-10)
	11     Same as GP0(E1h).Bit11  (see there)
	12-13  Unused (does NOT change GP0(E1h).Bit12-13)
	14-15  Unused (should be 0)
*/
func (gpu *GPU) RenderTexturedQuad() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[3], colour, gpu.fifo.buffer[4])
	v3 := NewVertex(gpu.fifo.buffer[5], colour, gpu.fifo.buffer[6])
	v4 := NewVertex(gpu.fifo.buffer[7], colour, gpu.fifo.buffer[8])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[4] >> 16

	clutX := GetValue(clutIndex, 0, 6) * 16
	clutY := GetValue(clutIndex, 6, 9)

	texPageUBase := GetValue(texPage, 0, 4) * 64
	texPageVBase := GetValue(texPage, 4, 1) * 256

	texFormat := GetValue(texPage, 7, 2)

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::TexturedQuad] reserved texture format")
	}

	// make sure that vertexes are in clockwise order
	gpu.TexturedTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat)
	gpu.TexturedTriangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat)
}

/* lol self plagiarised from https://github.com/marethyu/mgl/blob/main/mygl.h#L133 */
func (gpu *GPU) ShadedTriangle(v1, v2, v3 *Vertex) {
	// Area of the parallelogram formed by edge vectors
	area := float32(int32(v3.x-v1.x)*int32(v2.y-v1.y) - int32(v3.y-v1.y)*int32(v2.x-v1.x))

	// top left and bottom right points of a bounding box (-1 bc bottom and right edges are not drawn)
	xmin := MinOf(v1.x, v2.x, v3.x)
	xmax := MaxOf(v1.x, v2.x, v3.x) - 1
	ymin := MinOf(v1.y, v2.y, v3.y)
	ymax := MaxOf(v1.y, v2.y, v3.y) - 1

	// TODO clipping

	for y := ymin; y <= ymax; y += 1 {
		for x := xmin; x <= xmax; x += 1 {
			// Barycentric weights
			w1 := float32(int32(x-v2.x)*int32(v3.y-v2.y)-int32(y-v2.y)*int32(v3.x-v2.x)) / area
			w2 := float32(int32(x-v3.x)*int32(v1.y-v3.y)-int32(y-v3.y)*int32(v1.x-v3.x)) / area
			w3 := float32(int32(x-v1.x)*int32(v2.y-v1.y)-int32(y-v1.y)*int32(v2.x-v1.x)) / area

			if (w1 >= 0.0) && (w2 >= 0.0) && (w3 >= 0.0) {
				r := uint8(w1*float32(v1.r) + w2*float32(v2.r) + w3*float32(v3.r))
				g := uint8(w1*float32(v1.g) + w2*float32(v2.g) + w3*float32(v3.g))
				b := uint8(w1*float32(v1.b) + w2*float32(v2.b) + w3*float32(v3.b))

				gpu.Pixel(uint32(x), uint32(y), r, g, b)
			}
		}
	}
}

/*
Resources to learn more about textures:
- texture section in http://hitmen.c02.at/files/docs/psx/gpu.txt
- https://www.reddit.com/r/EmuDev/comments/fmhtcn/article_the_ps1_gpu_texture_pipeline_and_how_to/
- gpu section in https://web.archive.org/web/20190713020355/http://www.elisanet.fi/6581/PSX/doc/Playstation_Hardware.pdf
*/
func (gpu *GPU) TexturedTriangle(v1, v2, v3 *Vertex, clutX uint32, clutY uint32, texPageUBase uint32, texPageVBase uint32, texFormat uint32) {
	// Area of the parallelogram formed by edge vectors
	area := float32(int32(v3.x-v1.x)*int32(v2.y-v1.y) - int32(v3.y-v1.y)*int32(v2.x-v1.x))

	// top left and bottom right points of a bounding box (-1 bc bottom and right edges are not drawn)
	xmin := MinOf(v1.x, v2.x, v3.x)
	xmax := MaxOf(v1.x, v2.x, v3.x) - 1
	ymin := MinOf(v1.y, v2.y, v3.y)
	ymax := MaxOf(v1.y, v2.y, v3.y) - 1

	// TODO clipping

	for y := ymin; y <= ymax; y += 1 {
		for x := xmin; x <= xmax; x += 1 {
			// Barycentric weights
			w1 := float32(int32(x-v2.x)*int32(v3.y-v2.y)-int32(y-v2.y)*int32(v3.x-v2.x)) / area
			w2 := float32(int32(x-v3.x)*int32(v1.y-v3.y)-int32(y-v3.y)*int32(v1.x-v3.x)) / area
			w3 := float32(int32(x-v1.x)*int32(v2.y-v1.y)-int32(y-v1.y)*int32(v2.x-v1.x)) / area

			if (w1 >= 0.0) && (w2 >= 0.0) && (w3 >= 0.0) {
				u := uint32(uint8(w1*float32(v1.u) + w2*float32(v2.u) + w3*float32(v3.u)))
				v := uint32(uint8(w1*float32(v1.v) + w2*float32(v2.v) + w3*float32(v3.v)))

				var texel uint16

				switch texFormat {
				case TEXTURE_FORMAT_4b:
					texel16 := gpu.vram.Read16(texPageUBase+u/4, texPageVBase+v)
					index := uint32((texel16 >> ((u % 4) * 4)) & 0xf)
					texel = gpu.vram.Read16(clutX+index, clutY)
				case TEXTURE_FORMAT_8b:
					texel16 := gpu.vram.Read16(texPageUBase+u/2, texPageVBase+v)
					index := uint32((texel16 >> ((u % 2) * 8)) & 0xff)
					texel = gpu.vram.Read16(clutX+index, clutY)
				case TEXTURE_FORMAT_15b:
					texel = gpu.vram.Read16(texPageUBase+u, texPageVBase+v)
				}

				// don't draw black texels
				if texel > 0 {
					// r := uint8(GetValue(uint32(texel), 0, 5) << 3)
					// g := uint8(GetValue(uint32(texel), 5, 5) << 3)
					// b := uint8(GetValue(uint32(texel), 10, 5) << 3)
					// gpu.Pixel(uint32(x), uint32(y), r, g, b)

					gpu.vram.Write16(uint32(x), uint32(y), uint16(texel))
				}
			}
		}
	}
}

func (gpu *GPU) Pixel(x uint32, y uint32, r, g, b uint8) {
	var colour uint32 = 0

	PackValue(&colour, 0, uint32(r>>3), 5)
	PackValue(&colour, 5, uint32(g>>3), 5)
	PackValue(&colour, 10, uint32(b>>3), 5)

	gpu.vram.Write16(x, y, uint16(colour))
}
