package main

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
ClutV1U1
00R2G2B2
Yyy2Xxx2
PageV2U2
00R3G3B3
Yyy3Xxx3
0000V3U3
00R4G4B4
Yyy4Xxx4
0000V4U4

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
	isTextured := TestBit(gpu.shape_attr, PATTR_TEXTURE)
	isShaded := TestBit(gpu.shape_attr, PATTR_GOURAUD)

	if TestBit(gpu.shape_attr, PATTR_QUAD) {
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

/*
Resources to learn more about textures:
- texture section in http://hitmen.c02.at/files/docs/psx/gpu.txt
- https://www.reddit.com/r/EmuDev/comments/fmhtcn/article_the_ps1_gpu_texture_pipeline_and_how_to/
- gpu section in https://web.archive.org/web/20190713020355/http://www.elisanet.fi/6581/PSX/doc/Playstation_Hardware.pdf
*/
func (gpu *GPU) GetTexel(u, v, clutX, clutY, texPageUBase, texPageVBase, texFormat int) uint32 {
	switch texFormat {
	case TEXTURE_FORMAT_4b:
		texel16 := gpu.vram.Read16(texPageUBase+u/4, texPageVBase+v)
		index := int((texel16 >> ((u % 4) * 4)) & 0xf)
		return uint32(gpu.vram.Read16(clutX+index, clutY))
	case TEXTURE_FORMAT_8b:
		texel16 := gpu.vram.Read16(texPageUBase+u/2, texPageVBase+v)
		index := int((texel16 >> ((u % 2) * 8)) & 0xff)
		return uint32(gpu.vram.Read16(clutX+index, clutY))
	case TEXTURE_FORMAT_15b:
		return uint32(gpu.vram.Read16(texPageUBase+u, texPageVBase+v))
	}

	panic("[GPU::GetTexel] unreachable!")
}

func (gpu *GPU) TextureBlend(r, g, b, tr, tg, tb int) (int, int, int) {
	// adjust brightness of each texel (neutral value is 128)
	return Clamp8((r * tr) >> 7), Clamp8((g * tg) >> 7), Clamp8((b * tb) >> 7) // shift by 7 is same as dividing by 128
}

func (gpu *GPU) PutPixel(x, y, r, g, b int, m bool, semiTransparent bool, stMode int) {
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

	var colour uint32 = 0

	PackRange(&colour, 0, uint32(r>>3), 5)
	PackRange(&colour, 5, uint32(g>>3), 5)
	PackRange(&colour, 10, uint32(b>>3), 5)
	ModifyBit(&colour, 15, gpu.setMaskBit || m)

	gpu.vram.Write16(x, y, uint16(colour))
}
