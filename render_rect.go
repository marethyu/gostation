package main

const (
	RATTR_RAW_TEXTURE = iota
	RATTR_SEMI_TRANSPARENT
	RATTR_TEXTURE
)

const (
	RSIZE_VARIABLE = iota /* 0 (00) variable size */
	RSIZE_1x1             /* 1 (01) single pixel (1x1) */
	RSIZE_8x8             /* 2 (10) 8x8 sprite */
	RSIZE_16x16           /* 3 (11) 16x16 sprite */
)

/*
Argument format:

Color         ccBBGGRR    - command + color; color is ignored when textured
Vertex        YYYYXXXX    - required, indicates the upper left corner to render
UV            ClutVVUU    - optional, only present for textured rectangles (for 4bpp textures UU must be even!)
Width+Height  YsizXsiz    - optional, dimensions for variable sized rectangles (max 1023x511)
*/
func (gpu *GPU) ProcessRectangleCommand() {
	isTextured := TestBit(gpu.shape_attr, RATTR_TEXTURE)
	isVariable := GetRange(gpu.shape_attr, 3, 2) == 0

	if !isTextured && !isVariable {
		gpu.ProcessMonochromeRectCommand()
	} else if isTextured && !isVariable {
		gpu.ProcessTexturedRectCommand()
	} else if !isTextured && isVariable {
		gpu.ProcessMonochromeVariableRectCommand()
	} else {
		gpu.ProcessTexturedVariableRectCommand()
	}
}

func (gpu *GPU) ProcessMonochromeRectCommand() {
	x1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]&0xffff), 11))
	y1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]>>16), 11))

	r := int(GetRange(gpu.fifo.buffer[0], 0, 8))
	g := int(GetRange(gpu.fifo.buffer[0], 8, 8))
	b := int(GetRange(gpu.fifo.buffer[0], 16, 8))

	var x2, y2 int
	switch GetRange(gpu.shape_attr, 3, 2) {
	case RSIZE_1x1:
		x2 = x1 + 1
		y2 = y1 + 1
	case RSIZE_8x8:
		x2 = x1 + 8
		y2 = y1 + 8
	case RSIZE_16x16:
		x2 = x1 + 16
		y2 = y1 + 16
	default:
		panic("[GPU::ProcessMonochromeRectCommand] ???")
	}

	gpu.DoRenderRectangle(x1, y1, x2, y2, r, g, b, 0, 0, 0, 0, gpu.shape_attr)
}

func (gpu *GPU) ProcessTexturedRectCommand() {
	x1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]&0xffff), 11))
	y1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]>>16), 11))

	r := int(GetRange(gpu.fifo.buffer[0], 0, 8))
	g := int(GetRange(gpu.fifo.buffer[0], 8, 8))
	b := int(GetRange(gpu.fifo.buffer[0], 16, 8))

	var x2, y2 int
	switch GetRange(gpu.shape_attr, 3, 2) {
	case RSIZE_1x1:
		x2 = x1 + 1
		y2 = y1 + 1
	case RSIZE_8x8:
		x2 = x1 + 8
		y2 = y1 + 8
	case RSIZE_16x16:
		x2 = x1 + 16
		y2 = y1 + 16
	default:
		panic("[GPU::ProcessTexturedRectCommand] ???")
	}

	u := int(GetRange(gpu.fifo.buffer[2], 0, 8)) // TODO for 4bpp, it must be even
	v := int(GetRange(gpu.fifo.buffer[2], 8, 8))

	clutIndex := gpu.fifo.buffer[2] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	gpu.DoRenderRectangle(x1, y1, x2, y2, r, g, b, u, v, clutX, clutY, gpu.shape_attr)
}

func (gpu *GPU) ProcessMonochromeVariableRectCommand() {
	x1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]&0xffff), 11))
	y1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]>>16), 11))

	r := int(GetRange(gpu.fifo.buffer[0], 0, 8))
	g := int(GetRange(gpu.fifo.buffer[0], 8, 8))
	b := int(GetRange(gpu.fifo.buffer[0], 16, 8))

	width := int(uint16(gpu.fifo.buffer[2] & 0xffff))
	if width >= VRAM_WIDTH {
		width = VRAM_WIDTH - 1
	}
	x2 := x1 + width

	height := int(uint16(gpu.fifo.buffer[2] >> 16))
	if height >= VRAM_HEIGHT {
		height = VRAM_HEIGHT - 1
	}
	y2 := y1 + height

	gpu.DoRenderRectangle(x1, y1, x2, y2, r, g, b, 0, 0, 0, 0, gpu.shape_attr)
}

func (gpu *GPU) ProcessTexturedVariableRectCommand() {
	x1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]&0xffff), 11))
	y1 := int(ForceSignExtension16(uint16(gpu.fifo.buffer[1]>>16), 11))

	r := int(GetRange(gpu.fifo.buffer[0], 0, 8))
	g := int(GetRange(gpu.fifo.buffer[0], 8, 8))
	b := int(GetRange(gpu.fifo.buffer[0], 16, 8))

	width := int(uint16(gpu.fifo.buffer[3] & 0xffff))
	if width >= VRAM_WIDTH {
		width = VRAM_WIDTH - 1
	}
	x2 := x1 + width

	height := int(uint16(gpu.fifo.buffer[3] >> 16))
	if height >= VRAM_HEIGHT {
		height = VRAM_HEIGHT - 1
	}
	y2 := y1 + height

	u := int(GetRange(gpu.fifo.buffer[2], 0, 8)) // TODO for 4bpp, it must be even
	v := int(GetRange(gpu.fifo.buffer[2], 8, 8))

	clutIndex := gpu.fifo.buffer[2] >> 16

	clutX := int(GetRange(clutIndex, 0, 6) * 16)
	clutY := int(GetRange(clutIndex, 6, 9))

	gpu.DoRenderRectangle(x1, y1, x2, y2, r, g, b, u, v, clutX, clutY, gpu.shape_attr)
}

/*
Note: only draw from (x1,y1) to (x2-1,y2-1)
*/
func (gpu *GPU) DoRenderRectangle(x1, y1, x2, y2, R, G, B, startU, startV, clutX, clutY int, attr uint32) {
	isRawTexture := TestBit(attr, RATTR_RAW_TEXTURE)
	isSemiTransparent := TestBit(attr, RATTR_SEMI_TRANSPARENT)
	isTexture := TestBit(attr, RATTR_TEXTURE)

	texPageUBase := gpu.txBase * 64
	texPageVBase := gpu.tyBase * 256

	v := startV

	var uInc int = 0
	var vInc int = 0

	if isTexture {
		if gpu.rectTextureXFlip {
			uInc = -1
		} else {
			uInc = 1
		}
		if gpu.rectTextureYFlip {
			vInc = -1
		} else {
			vInc = 1
		}
	}

	for y := y1; y < y2; y += 1 {
		u := startU

		for x := x1; x < x2; x += 1 {
			r, g, b := R, G, B
			m := false
			semiTransparent := isSemiTransparent
			draw := true

			if isTexture {
				texel := gpu.GetTexel(u, v, clutX, clutY, texPageUBase, texPageVBase, gpu.textureFormat)

				if texel > 0 {
					tr := int(GetRange(texel, 0, 5) << 3)
					tg := int(GetRange(texel, 5, 5) << 3)
					tb := int(GetRange(texel, 10, 5) << 3)
					stp := TestBit(texel, 15)

					// TODO texture masking? no?

					if isRawTexture {
						r = tr
						g = tg
						b = tb
					} else {
						r, g, b = gpu.TextureBlend(r, g, b, tr, tg, tb)
					}
					m = stp

					semiTransparent = semiTransparent && stp
				} else {
					draw = false
				}
			}

			if draw {
				gpu.PutPixel(x, y, r, g, b, m, semiTransparent, gpu.semiTransparency)
			}

			u = Modulo(u+uInc, 256)
		}

		v = Modulo(v+vInc, 256)
	}
}
