package main

const (
	PATTR_RAW_TEXTURE = iota
	PATTR_SEMI_TRANSPARENT
	PATTR_TEXTURE
	PATTR_QUAD
	PATTR_GOURAUD
)

func (gpu *GPU) ProcessMonochromeQuadCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)
	v4 := NewVertex(gpu.fifo.buffer[4], colour, 0)

	gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessShadedQuadCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)
	v4 := NewVertex(gpu.fifo.buffer[7], gpu.fifo.buffer[6], 0)

	gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
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

	gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
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

	gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)

	area := Edge(v2.x, v2.y, v4.x, v4.y, v3.x, v3.y)
	if area < 0 {
		gpu.DoRenderTriangle(v2, v4, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v2, v3, v4, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessMonochromeTrigCommand() {
	colour := gpu.fifo.buffer[0]

	v1 := NewVertex(gpu.fifo.buffer[1], colour, 0)
	v2 := NewVertex(gpu.fifo.buffer[2], colour, 0)
	v3 := NewVertex(gpu.fifo.buffer[3], colour, 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	}
}

func (gpu *GPU) ProcessShadedTrigCommand() {
	v1 := NewVertex(gpu.fifo.buffer[1], gpu.fifo.buffer[0], 0)
	v2 := NewVertex(gpu.fifo.buffer[3], gpu.fifo.buffer[2], 0)
	v3 := NewVertex(gpu.fifo.buffer[5], gpu.fifo.buffer[4], 0)

	area := Edge(v1.x, v1.y, v3.x, v3.y, v2.x, v2.y)
	if area < 0 {
		gpu.DoRenderTriangle(v1, v3, v2, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, 0, 0, 0, 0, 0, gpu.semiTransparency, gpu.shape_attr)
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
		gpu.DoRenderTriangle(v1, v3, v2, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
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
		gpu.DoRenderTriangle(v1, v3, v2, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	} else {
		gpu.DoRenderTriangle(v1, v2, v3, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode, gpu.shape_attr)
	}
}

/*
Note: v1, v2, and v3 must be in clockwise order
*/
func (gpu *GPU) DoRenderTriangle(v1, v2, v3 *Vertex, clutX, clutY, texPageUBase, texPageVBase, texFormat, stMode int, attr uint32) {
	isRawTexture := TestBit(attr, PATTR_RAW_TEXTURE)
	isSemiTransparent := TestBit(attr, PATTR_SEMI_TRANSPARENT)
	isTexture := TestBit(attr, PATTR_TEXTURE)

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
					texel := gpu.GetTexel(u, v, clutX, clutY, texPageUBase, texPageVBase, texFormat)

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
						} else {
							r, g, b = gpu.TextureBlend(r, g, b, tr, tg, tb)
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
					gpu.PutPixel(x, y, r, g, b, m, semiTransparent, stMode)
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
