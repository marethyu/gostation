package main

const (
	ATTR_RAW_TEXTURE = iota
	ATTR_SEMI_TRANSPARENT
	ATTR_TEXTURE
	ATTR_QUAD
	ATTR_GOURAUD
)

/*
1--2    Note: A 4 point polygon is processed internally as two 3 point
|  |    polygons.
3--4    Note: When drawing a polygon the GPU will not draw the right

	most and bottom edge. So a (0,0)-(32,32) rectangle will actually

be drawn as (0,0)-(31,31). Make sure adjoining polygons have the same
coordinates if you want them to touch eachother!. Haven't checked how this
works with 3 point polygons.

Example argument list (quad with gouraud shading and texture blending):

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
func (gpu *GPU) ProcessPolygonCommand() {
	isTextured := TestBit(gpu.shape_attr, ATTR_TEXTURE)
	isShaded := TestBit(gpu.shape_attr, ATTR_GOURAUD)

	if TestBit(gpu.shape_attr, ATTR_QUAD) {
		if !isShaded && !isTextured {
			gpu.ProcessMonochromeQuadCommand()
		} else if isShaded && !isTextured {
			gpu.ProcessShadedQuadCommand()
		} else if !isShaded && isTextured {
			gpu.ProcessTexturedQuadCommand()
		} else {
			gpu.ProcessTexturedShadedQuadCommand()
		}
	} else {
		if !isShaded && !isTextured {
			gpu.ProcessMonochromeTrigCommand()
		} else if isShaded && !isTextured {
			gpu.ProcessShadedTrigCommand()
		} else if !isShaded && isTextured {
			gpu.ProcessTexturedTrigCommand()
		} else {
			gpu.ProcessTexturedShadedTrigCommand()
		}
	}
}

func (gpu *GPU) ProcessMonochromeQuadCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)
	v4 := NewVertex(gpu.fifo.buffer[4], colour, 0)

	gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessShadedQuadCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], 0)

	gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessTexturedQuadCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[3], colour, gpu.fifo.buffer[4])
	v3 := NewVertex(gpu.fifo.buffer[5], colour, gpu.fifo.buffer[6])
	v4 := NewVertex(gpu.fifo.buffer[7], colour, gpu.fifo.buffer[8])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[4] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	texPageUBase := int(GetRange(texPage, 0, 4) * 64)
	texPageVBase := int(GetRange(texPage, 4, 1) * 256)

	stMode := int(GetRange(texPage, 5, 2))

	texFormat := int(GetRange(texPage, 7, 2))

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::ProcessTexturedQuadCommand] reserved texture format")
	}

	gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessTexturedShadedQuadCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[4], gpu.fifo.buffer[3], gpu.fifo.buffer[5])
	v3 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], gpu.fifo.buffer[8])
	v4 := NewVertex(gpu.fifo.buffer[10], gpu.fifo.buffer[9], gpu.fifo.buffer[11])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[5] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	texPageUBase := int(GetRange(texPage, 0, 4) * 64)
	texPageVBase := int(GetRange(texPage, 4, 1) * 256)

	stMode := int(GetRange(texPage, 5, 2))

	texFormat := int(GetRange(texPage, 7, 2))

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::ProcessTexturedShadedQuadCommand] reserved texture format")
	}

	gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessMonochromeTrigCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessShadedTrigCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, int(gpu.semiTransparency), 0, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessTexturedTrigCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[3], colour, gpu.fifo.buffer[4])
	v3 := NewVertex(gpu.fifo.buffer[5], colour, gpu.fifo.buffer[6])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[4] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	texPageUBase := int(GetRange(texPage, 0, 4) * 64)
	texPageVBase := int(GetRange(texPage, 4, 1) * 256)

	stMode := int(GetRange(texPage, 5, 2))

	texFormat := int(GetRange(texPage, 7, 2))

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::ProcessTexturedTrigCommand] reserved texture format")
	}

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessTexturedShadedTrigCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], gpu.fifo.buffer[2])
	v2 := NewVertex(gpu.fifo.buffer[4], gpu.fifo.buffer[3], gpu.fifo.buffer[5])
	v3 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], gpu.fifo.buffer[8])

	clutIndex := gpu.fifo.buffer[2] >> 16
	texPage := gpu.fifo.buffer[5] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	texPageUBase := int(GetRange(texPage, 0, 4) * 64)
	texPageVBase := int(GetRange(texPage, 4, 1) * 256)

	stMode := int(GetRange(texPage, 5, 2))

	texFormat := int(GetRange(texPage, 7, 2))

	if texFormat == TEXTURE_FORMAT_reserved {
		panic("[GPU::ProcessTexturedShadedTrigCommand] reserved texture format")
	}

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat, gpu.shape_attr)
	}
}

/*
Resources to learn more about textures:
- texture section in http://hitmen.c02.at/files/docs/psx/gpu.txt
- https://www.reddit.com/r/EmuDev/comments/fmhtcn/article_the_ps1_gpu_texture_pipeline_and_how_to/
- gpu section in https://web.archive.org/web/20190713020355/http://www.elisanet.fi/6581/PSX/doc/Playstation_Hardware.pdf

Note: v1, v2, and v3 must be in clockwise order
*/
func (gpu *GPU) DoRenderTriangle(v1, v2, v3 *Vertex, clutX, clutY, texPageUBase, texPageVBase, stMode, texFormat int, attr uint32) {
	isRawTexture := TestBit(attr, ATTR_RAW_TEXTURE)
	isSemiTransparent := TestBit(attr, ATTR_SEMI_TRANSPARENT)
	isTexture := TestBit(attr, ATTR_TEXTURE)

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
		panic("[GPU::DoRenderTriangle] negative area!")
	}

	for y := ymin; y <= ymax; y += 1 {
		w1 := w1Row
		w2 := w2Row
		w3 := w3Row

		for x := xmin; x <= xmax; x += 1 {
			if gpu.drawUnmaskedPixels {
				pix := uint32(gpu.vram.Read16(x, y))

				if TestBit(pix, 15) {
					// masked
					w1 += incX23
					w2 += incX31
					w3 += incX12
					continue
				}
			}

			if (w1 > 0 || (w1 == 0 && topLeft23)) &&
				(w2 > 0 || (w2 == 0 && topLeft31)) &&
				(w3 > 0 || (w3 == 0 && topLeft12)) {
				r := (w1*v1.r + w2*v2.r + w3*v3.r) / area
				g := (w1*v1.g + w2*v2.g + w3*v3.g) / area
				b := (w1*v1.b + w2*v2.b + w3*v3.b) / area
				m := false
				semiTransparent := isSemiTransparent
				draw := true

				if isTexture {
					u := (w1*v1.u + w2*v2.u + w3*v3.u) / area
					v := (w1*v1.v + w2*v2.v + w3*v3.v) / area

					var texel uint32
					switch texFormat {
					case TEXTURE_FORMAT_4b:
						texel16 := gpu.vram.Read16(texPageUBase+u/4, texPageVBase+v)
						index := int((texel16 >> ((u % 4) * 4)) & 0xf)
						texel = uint32(gpu.vram.Read16(clutX+index, clutY))
					case TEXTURE_FORMAT_8b:
						texel16 := gpu.vram.Read16(texPageUBase+u/2, texPageVBase+v)
						index := int((texel16 >> ((u % 2) * 8)) & 0xff)
						texel = uint32(gpu.vram.Read16(clutX+index, clutY))
					case TEXTURE_FORMAT_15b:
						texel = uint32(gpu.vram.Read16(texPageUBase+u, texPageVBase+v))
					}

					if texel > 0 {
						tr := int(GetRange(texel, 0, 5) << 3)
						tg := int(GetRange(texel, 5, 5) << 3)
						tb := int(GetRange(texel, 10, 5) << 3)
						stp := TestBit(texel, 15)

						// TODO texture masking?

						if isRawTexture {
							r = tr
							g = tg
							b = tb
						} else { /* texture blend */
							// adjust brightness of each texel (neutral value is 128)
							r = Clamp8((r * tr) >> 7) // shift by 7 is same as dividing by 128
							g = Clamp8((g * tg) >> 7)
							b = Clamp8((b * tb) >> 7)
						}
						m = stp

						/* Summary of semi-transparency handling: https://github.com/ABelliqueux/nolibgs_hello_worlds/wiki/TIM#transparency

						semi-transparency mode is set thru bit 25 of command
						16-bit texel (format: STP, B, G, R) | semi-transparency mode off  | semi-transparency mode on
						(0, 0, 0, 0)                          transparent (no draw)         transparent (no draw)
						(1, 0, 0, 0)                          non-transparent (draw black)  semi-transparent
						(0, n, n, n)                          non-transparent               non-transparent
						(1, n, n, n)                          non-transparent               semi-transparent
						*/
						semiTransparent = semiTransparent && stp
					} else {
						// don't draw black (with MSB=0) texels
						draw = false
					}
				}

				if draw {
					if semiTransparent {
						backp := uint32(gpu.vram.Read16(x, y))

						br := int(GetRange(backp, 0, 5) << 3)
						bg := int(GetRange(backp, 5, 5) << 3)
						bb := int(GetRange(backp, 10, 5) << 3)

						switch stMode {
						case SEMI_TRANSPARENT_MODE0:
							// TODO it needs to be brighter (run RenderTexturePolygonCLUT4BPP.exe)
							r = (br + r) >> 1
							g = (bg + g) >> 1
							b = (bb + b) >> 1
						case SEMI_TRANSPARENT_MODE1:
							r = Clamp8(br + r)
							g = Clamp8(bg + g)
							b = Clamp8(bb + b)
						case SEMI_TRANSPARENT_MODE2:
							r = Clamp8(br - r)
							g = Clamp8(bg - g)
							b = Clamp8(bb - b)
						case SEMI_TRANSPARENT_MODE3:
							r = Clamp8(br + (r >> 2))
							g = Clamp8(bg + (g >> 2))
							b = Clamp8(bb + (b >> 2))
						}
					}

					gpu.PutPixel(x, y, r, g, b, m)
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
