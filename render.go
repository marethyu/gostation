package main

import (
	"fmt"
)

const (
	SHAPE_POLYGON = iota
)

const (
	TEXTURED = iota
	TEXTURE_RAW
	SEMI_TRANSPARENT /* TODO */
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
	case 0b00000:
		gpu.RenderMonochromeTrig()
	case 0b10000:
		gpu.RenderTrigGouraud()
	case 0b00010:
		gpu.RenderSemitransparentTrig()
	case 0b10010:
		gpu.RenderSemitransparentTrigGouraud()
	case 0b01000:
		gpu.RenderMonochromeQuad()
	case 0b11000:
		gpu.RenderQuadGouraud()
	case 0b01010:
		gpu.RenderSemitransparentQuad()
	case 0b11010:
		gpu.RenderSemitransparentQuadGouraud()
	case 0b01100:
		gpu.RenderTexturedQuadWithBlending()
	default:
		panic(fmt.Sprintf("[GPU::DoRenderPolygon] Unknown attribute: %05b\n", gpu.shape_attr))
	}
}

func (gpu *GPU) RenderMonochromeTrig() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.Triangle(v1, v3, v2, 0, 0, 0, 0, 0, 0)
	} else {
		gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, 0)
	}
}

func (gpu *GPU) RenderTrigGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.Triangle(v1, v3, v2, 0, 0, 0, 0, 0, 0)
	} else {
		gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, 0)
	}
}

func (gpu *GPU) RenderSemitransparentTrig() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)

	var mask uint32 = 0

	ModifyBit(&mask, SEMI_TRANSPARENT, true)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.Triangle(v1, v3, v2, 0, 0, 0, 0, 0, mask)
	} else {
		gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, mask)
	}
}

func (gpu *GPU) RenderSemitransparentTrigGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)

	var mask uint32 = 0

	ModifyBit(&mask, SEMI_TRANSPARENT, true)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.Triangle(v1, v3, v2, 0, 0, 0, 0, 0, mask)
	} else {
		gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, mask)
	}
}

func (gpu *GPU) RenderMonochromeQuad() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)
	v4 := NewVertex(gpu.fifo.buffer[4], colour, 0)

	gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, 0)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.Triangle(v2, v4, v3, 0, 0, 0, 0, 0, 0)
	} else {
		gpu.Triangle(v2, v3, v4, 0, 0, 0, 0, 0, 0)
	}
}

func (gpu *GPU) RenderQuadGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], 0)

	gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, 0)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.Triangle(v2, v4, v3, 0, 0, 0, 0, 0, 0)
	} else {
		gpu.Triangle(v2, v3, v4, 0, 0, 0, 0, 0, 0)
	}
}

func (gpu *GPU) RenderSemitransparentQuad() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)
	v4 := NewVertex(gpu.fifo.buffer[4], colour, 0)

	var mask uint32 = 0

	ModifyBit(&mask, SEMI_TRANSPARENT, true)

	gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, mask)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.Triangle(v2, v4, v3, 0, 0, 0, 0, 0, mask)
	} else {
		gpu.Triangle(v2, v3, v4, 0, 0, 0, 0, 0, mask)
	}
}

func (gpu *GPU) RenderSemitransparentQuadGouraud() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], 0)

	var mask uint32 = 0

	ModifyBit(&mask, SEMI_TRANSPARENT, true)

	gpu.Triangle(v1, v2, v3, 0, 0, 0, 0, 0, mask)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.Triangle(v2, v4, v3, 0, 0, 0, 0, 0, mask)
	} else {
		gpu.Triangle(v2, v3, v4, 0, 0, 0, 0, 0, mask)
	}
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
func (gpu *GPU) RenderTexturedQuadWithBlending() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[3], colour, gpu.fifo.buffer[4])
	v3 := NewVertex(gpu.fifo.buffer[5], colour, gpu.fifo.buffer[6])
	v4 := NewVertex(gpu.fifo.buffer[7], colour, gpu.fifo.buffer[8])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[4] >> 16

	clutX := GetRange(clutIndex, 0, 6) * 16
	clutY := GetRange(clutIndex, 6, 9)

	texPageUBase := GetRange(texPage, 0, 4) * 64
	texPageVBase := GetRange(texPage, 4, 1) * 256

	texFormat := GetRange(texPage, 7, 2)

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::RenderTexturedQuadWithBlending] reserved texture format")
	}

	var mask uint32 = 0

	ModifyBit(&mask, TEXTURED, true)
	ModifyBit(&mask, TEXTURE_RAW, false)

	gpu.Triangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, mask)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.Triangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, mask)
	} else {
		gpu.Triangle(v2, v3, v4, clutX, clutY, texPageUBase, texPageVBase, texFormat, mask)
	}
}

/*
Resources to learn more about textures:
- texture section in http://hitmen.c02.at/files/docs/psx/gpu.txt
- https://www.reddit.com/r/EmuDev/comments/fmhtcn/article_the_ps1_gpu_texture_pipeline_and_how_to/
- gpu section in https://web.archive.org/web/20190713020355/http://www.elisanet.fi/6581/PSX/doc/Playstation_Hardware.pdf

Note: v1, v2, and v3 must be in clockwise order
*/
func (gpu *GPU) Triangle(v1, v2, v3 *Vertex, clutX uint32, clutY uint32, texPageUBase uint32, texPageVBase uint32, texFormat uint32, settings uint32) {
	xmin := MinOf(v1.x, v2.x, v3.x)
	xmax := MaxOf(v1.x, v2.x, v3.x)
	ymin := MinOf(v1.y, v2.y, v3.y)
	ymax := MaxOf(v1.y, v2.y, v3.y)

	// TODO clipping

	topLeft12 := IsTopLeft(v1, v2) // v1-v2 edge
	topLeft23 := IsTopLeft(v2, v3) // v2-v3 edge
	topLeft31 := IsTopLeft(v3, v1) // v3-v1 edge

	w1Row := Edge(v3.x, v3.y, v2.x, v2.y, xmin, ymin) // 2-3
	w2Row := Edge(v1.x, v1.y, v3.x, v3.y, xmin, ymin) // 3-1
	w3Row := Edge(v2.x, v2.y, v1.x, v1.y, xmin, ymin) // 1-2

	incX23 := v2.y - v3.y
	incX31 := v3.y - v1.y
	incX12 := v1.y - v2.y
	incY23 := v3.x - v2.x
	incY31 := v1.x - v3.x
	incY12 := v2.x - v1.x

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)

	if area < 0 {
		panic("[GPU::Triangle] negative area!")
	}

	for y := ymin; y <= ymax; y += 1 {
		w1 := w1Row
		w2 := w2Row
		w3 := w3Row

		for x := xmin; x <= xmax; x += 1 {
			if gpu.drawUnmaskedPixels {
				pix := uint32(gpu.vram.Read16(uint32(x), uint32(y)))

				if TestBit(pix, 15) {
					continue // masked
				}
			}

			if (w1 > 0 || (w1 == 0 && topLeft23)) &&
				(w2 > 0 || (w2 == 0 && topLeft31)) &&
				(w3 > 0 || (w3 == 0 && topLeft12)) {
				r := uint8((w1*v1.r + w2*v2.r + w3*v3.r) / area)
				g := uint8((w1*v1.g + w2*v2.g + w3*v3.g) / area)
				b := uint8((w1*v1.b + w2*v2.b + w3*v3.b) / area)

				if TestBit(settings, TEXTURED) {
					u := uint32(uint8((w1*v1.u + w2*v2.u + w3*v3.u) / area))
					v := uint32(uint8((w1*v1.v + w2*v2.v + w3*v3.v) / area))

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
						tr := uint8(GetRange(uint32(texel), 0, 5) << 3)
						tg := uint8(GetRange(uint32(texel), 5, 5) << 3)
						tb := uint8(GetRange(uint32(texel), 10, 5) << 3)

						if TestBit(settings, TEXTURE_RAW) {
							gpu.Pixel(uint32(x), uint32(y), tr, tg, tb, TestBit(uint32(texel), 15))
						} else { /* texture blend */
							// adjust brightness of each texel (neutral value is 128)
							finalR := Clamp8((int32(r) * int32(tr)) >> 7) // shift by 7 is same as dividing by 128
							finalG := Clamp8((int32(g) * int32(tg)) >> 7)
							finalB := Clamp8((int32(b) * int32(tb)) >> 7)

							gpu.Pixel(uint32(x), uint32(y), finalR, finalG, finalB, TestBit(uint32(texel), 15))
						}
					}
				} else {
					if TestBit(settings, SEMI_TRANSPARENT) {
						// see semi-transparency section in http://hitmen.c02.at/files/docs/psx/gpu.txt

						backp := gpu.vram.Read16(uint32(x), uint32(y))

						br := uint8(GetRange(uint32(backp), 0, 5) << 3)
						bg := uint8(GetRange(uint32(backp), 5, 5) << 3)
						bb := uint8(GetRange(uint32(backp), 10, 5) << 3)

						var finalR, finalG, finalB uint8
						switch gpu.semiTransparency {
						case SEMI_TRANSPARENT_MODE0:
							finalR = uint8((uint16(br) + uint16(r)) >> 1)
							finalG = uint8((uint16(bg) + uint16(g)) >> 1)
							finalB = uint8((uint16(bb) + uint16(b)) >> 1)
						case SEMI_TRANSPARENT_MODE1:
							finalR = Clamp8(int32(br) + int32(r))
							finalG = Clamp8(int32(bg) + int32(g))
							finalB = Clamp8(int32(bb) + int32(b))
						case SEMI_TRANSPARENT_MODE2:
							finalR = Clamp8(int32(br) - int32(r))
							finalG = Clamp8(int32(bg) - int32(g))
							finalB = Clamp8(int32(bb) - int32(b))
						case SEMI_TRANSPARENT_MODE3:
							finalR = Clamp8(int32(br) + (int32(r) >> 2))
							finalG = Clamp8(int32(bg) + (int32(g) >> 2))
							finalB = Clamp8(int32(bb) + (int32(b) >> 2))
						}

						gpu.Pixel(uint32(x), uint32(y), finalR, finalG, finalB, false)
					} else {
						gpu.Pixel(uint32(x), uint32(y), r, g, b, false)
					}
				}
			}

			w1 += incX23
			w2 += incX31
			w3 += incX12
		}

		w1Row += incY23
		w2Row += incY31
		w3Row += incY12
	}
}

func (gpu *GPU) Pixel(x uint32, y uint32, r, g, b uint8, m bool) {
	var colour uint32 = 0

	PackRange(&colour, 0, uint32(r>>3), 5)
	PackRange(&colour, 5, uint32(g>>3), 5)
	PackRange(&colour, 10, uint32(b>>3), 5)
	ModifyBit(&colour, 15, gpu.setMaskBit || m)

	gpu.vram.Write16(x, y, uint16(colour))
}
