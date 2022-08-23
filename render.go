package main

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
