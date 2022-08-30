package main

import (
	"fmt"
)

func (gpu *GPU) GP0ExecuteMiscCommand(cmd uint32) {
	op := (cmd >> 24) & 0xf

	switch op {
	case 0x0:
		// NOP
	case 0x1:
		// TODO clear texture cache
	case 0x2:
		gpu.GP0InitFillRectangleVRAM(cmd)
	default:
		panic(fmt.Sprintf("[GPU::ExecuteMiscCommand] Unknown command: %x", cmd))
	}
}

/*
Format for polygon command:

	bit number   value   meaning
	 31-29        001    polygon render
	   28         1/0    gouraud / flat shading
	   27         1/0    4 / 3 vertices
	   26         1/0    textured / untextured
	   25         1/0    semi transparent / solid
	   24         1/0    raw texture / texture blending
	  23-0        rgb    first color value.
*/
func (gpu *GPU) GP0InitRenderPolygonCommand(cmd uint32) {
	gpu.shape = PRIMITIVE_POLYGON
	gpu.shape_attr = GetRange(cmd, 24, 5)

	nvert := 3 // polygons are triangles by default

	if TestBit(cmd, 27) {
		// well, it's quad then
		nvert += 1
	}

	narg := nvert

	if TestBit(cmd, 26) {
		// If doing textured rendering, each vertex sent will also have a U/V texture coordinate attached to it, as well as a CLUT index.
		narg += nvert
	}

	if TestBit(cmd, 28) {
		// If doing gouraud shading, there will be one more color per vertex sent, and the initial color will be the one for vertex 0.
		narg += nvert - 1
	}

	narg += 1

	gpu.mode = MODE_RENDERING
	gpu.fifo.Reset(narg)
	gpu.fifo.Push(cmd) // the initial command can be treated as the first argument
	gpu.fifoActive = true
}

/*
Format for rectangle command:

	bit number   value   meaning
	 31-29        011    rectangle render
	 28-27        sss    rectangle size
	   26         1/0    textured / untextured
	   25         1/0    semi transparent / solid
	   24         1/0    raw texture / texture blending
	  23-0        rgb    first color value.
*/
func (gpu *GPU) GP0InitRenderRectangleCommand(cmd uint32) {
	gpu.shape = PRIMITIVE_RECTANGLE
	gpu.shape_attr = GetRange(cmd, 24, 5)

	narg := 2

	if TestBit(cmd, 26) {
		// textured
		narg += 1
	}

	if GetRange(cmd, 27, 2) == 0 {
		// variable sized
		narg += 1
	}

	gpu.mode = MODE_RENDERING
	gpu.fifo.Reset(narg)
	gpu.fifo.Push(cmd)
	gpu.fifoActive = true
}

/*
For transferring data (like texture or palettes) from cpu to gpu's vram

1st  Command
2nd  Destination Coord (YyyyXxxxh)  ;Xpos counted in halfwords
3rd  Width+Height      (YsizXsizh)  ;Xsiz counted in halfwords
...  Data              (...)      <--- usually transferred via DMA
*/
func (gpu *GPU) GP0InitCPUToVRamBlit(cmd uint32) {
	gpu.mode = MODE_CPUtoVRamBlit
	gpu.fifo.Reset(3)
	gpu.fifo.Push(cmd)
	gpu.fifoActive = true
}

/*
Actual cpu to vram transfer. It transfers data from cpu into a specified rectangular area in vram
*/
func (gpu *GPU) GP0DoCPUToVramTransfer(data uint16) {
	vramX := gpu.startX + gpu.imgX
	vramY := gpu.startY + gpu.imgY

	gpu.vram.Write16(vramX, vramY, data)

	gpu.imgX += 1

	if gpu.imgX == gpu.imgWidth {
		gpu.imgX = 0
		gpu.imgY += 1
	}
}

/*
Opposite of GPU::InitCPUToVRamBlit

1st  Command                       ;\
2nd  Source Coord      (YyyyXxxxh) ; write to GP0 port (as usually)
3rd  Width+Height      (YsizXsizh) ;/
...  Data              (...)       ;<--- read from GPUREAD port (or via DMA)
*/
func (gpu *GPU) GP0InitVramToCPUBlit(cmd uint32) {
	gpu.mode = MODE_VramtoCPUBlit
	gpu.fifo.Reset(3)
	gpu.fifo.Push(cmd)
	gpu.fifoActive = true
}

func (gpu *GPU) GP0DoVramToCPUTransfer() uint16 {
	vramX := gpu.startX + gpu.imgX
	vramY := gpu.startY + gpu.imgY

	data := gpu.vram.Read16(vramX, vramY)

	gpu.imgX += 1

	if gpu.imgX == gpu.imgWidth {
		gpu.imgX = 0
		gpu.imgY += 1
	}

	return data
}

/*
GP0(02h) - Fill Rectangle in VRAM

	1st  Color+Command     (CcBbGgRrh)  ;24bit RGB value (see note)
	2nd  Top Left Corner   (YyyyXxxxh)  ;Xpos counted in halfwords, steps of 10h
	3rd  Width+Height      (YsizXsizh)  ;Xsiz counted in halfwords, steps of 10h
*/
func (gpu *GPU) GP0InitFillRectangleVRAM(cmd uint32) {
	gpu.mode = MODE_FillVRam
	gpu.fifo.Reset(3)
	gpu.fifo.Push(cmd)
	gpu.fifoActive = true
}

func (gpu *GPU) GP0ExecuteEnvironmentCommand(cmd uint32) {
	op := (cmd >> 24) & 0xf

	switch op {
	case 0x1:
		gpu.GP0DrawModeSet(cmd)
	case 0x2:
		gpu.GP0TextureWindowSetup(cmd)
	case 0x3:
		gpu.GP0DrawingAreaTopLeftSet(cmd)
	case 0x4:
		gpu.GP0DrawingAreaBottomRightSet(cmd)
	case 0x5:
		gpu.GP0DrawingOffsetSet(cmd)
	case 0x6:
		gpu.GP0MaskBitSetup(cmd)
	default:
		panic(fmt.Sprintf("[GPU::GP0ExecuteEnvironmentCommand] Unknown command: %x", cmd))
	}
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e1h-draw-mode-setting-aka-texpage

0-3   Texture page X Base   (N*64) (ie. in 64-halfword steps)    ;GPUSTAT.0-3
4     Texture page Y Base   (N*256) (ie. 0 or 256)               ;GPUSTAT.4
5-6   Semi Transparency     (0=B/2+F/2, 1=B+F, 2=B-F, 3=B+F/4)   ;GPUSTAT.5-6
7-8   Texture page colors   (0=4bit, 1=8bit, 2=15bit, 3=Reserved);GPUSTAT.7-8
9     Dither 24bit to 15bit (0=Off/strip LSBs, 1=Dither Enabled) ;GPUSTAT.9
10    Drawing to display area (0=Prohibited, 1=Allowed)          ;GPUSTAT.10
11    Texture Disable (0=Normal, 1=Disable if GP1(09h).Bit0=1)   ;GPUSTAT.15

	(Above might be chipselect for (absent) second VRAM chip?)

12    Textured Rectangle X-Flip   (BIOS does set this bit on power-up...?)
13    Textured Rectangle Y-Flip   (BIOS does set it equal to GPUSTAT.13...?)
14-23 Not used (should be 0)
24-31 Command  (E1h)
*/
func (gpu *GPU) GP0DrawModeSet(data uint32) {
	gpu.txBase = int(GetRange(data, 0, 4))
	gpu.tyBase = int(GetRange(data, 4, 1))
	gpu.semiTransparency = int(GetRange(data, 5, 2))
	gpu.textureFormat = int(GetRange(data, 7, 2))
	gpu.dilthering = TestBit(data, 9)
	gpu.drawToDisplay = TestBit(data, 10)
	gpu.textureDisable = TestBit(data, 11)
	gpu.rectTextureXFlip = TestBit(data, 12)
	gpu.rectTextureYFlip = TestBit(data, 13)
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e2h-texture-window-setting

0-4    Texture window Mask X   (in 8 pixel steps)
5-9    Texture window Mask Y   (in 8 pixel steps)
10-14  Texture window Offset X (in 8 pixel steps)
15-19  Texture window Offset Y (in 8 pixel steps)
20-23  Not used (zero)
24-31  Command  (E2h)
*/
func (gpu *GPU) GP0TextureWindowSetup(data uint32) {
	gpu.texWindowMaskX = int(GetRange(data, 0, 5))
	gpu.texWindowMaskY = int(GetRange(data, 5, 5))
	gpu.texWindowOffsetX = int(GetRange(data, 10, 5))
	gpu.texWindowOffsetY = int(GetRange(data, 15, 5))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e3h-set-drawing-area-top-left-x1y1

0-9    X-coordinate (0..1023)
10-18  Y-coordinate (0..511)   ;\on Old 160pin GPU (max 1MB VRAM)
19-23  Not used (zero)         ;/
10-19  Y-coordinate (0..1023)  ;\on New 208pin GPU (max 2MB VRAM)
20-23  Not used (zero)         ;/(retail consoles have only 1MB though)
24-31  Command  (Exh)
*/
func (gpu *GPU) GP0DrawingAreaTopLeftSet(data uint32) {
	gpu.drawingAreaX1 = int(GetRange(data, 0, 10))
	gpu.drawingAreaY1 = int(GetRange(data, 10, 10))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e4h-set-drawing-area-bottom-right-x2y2

0-9    X-coordinate (0..1023)
10-18  Y-coordinate (0..511)   ;\on Old 160pin GPU (max 1MB VRAM)
19-23  Not used (zero)         ;/
10-19  Y-coordinate (0..1023)  ;\on New 208pin GPU (max 2MB VRAM)
20-23  Not used (zero)         ;/(retail consoles have only 1MB though)
24-31  Command  (Exh)
*/
func (gpu *GPU) GP0DrawingAreaBottomRightSet(data uint32) {
	gpu.drawingAreaX2 = int(GetRange(data, 0, 10))
	gpu.drawingAreaY2 = int(GetRange(data, 10, 10))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e5h-set-drawing-offset-xy

0-10   X-offset (-1024..+1023) (usually within X1,X2 of Drawing Area)
11-21  Y-offset (-1024..+1023) (usually within Y1,Y2 of Drawing Area)
22-23  Not used (zero)
24-31  Command  (E5h)
*/
func (gpu *GPU) GP0DrawingOffsetSet(data uint32) {
	x := uint16(GetRange(data, 0, 11))
	y := uint16(GetRange(data, 11, 11))

	gpu.drawingXOffset = int(ForceSignExtension16(x, 11))
	gpu.drawingYOffset = int(ForceSignExtension16(y, 11))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e6h-mask-bit-setting

0     Set mask while drawing (0=TextureBit15, 1=ForceBit15=1)   ;GPUSTAT.11
1     Check mask before draw (0=Draw Always, 1=Draw if Bit15=0) ;GPUSTAT.12
2-23  Not used (zero)
24-31 Command  (E6h)
*/
func (gpu *GPU) GP0MaskBitSetup(data uint32) {
	gpu.setMaskBit = TestBit(data, 0)
	gpu.drawUnmaskedPixels = TestBit(data, 1)
}

func (gpu *GPU) GP0RenderPrimitive() {
	switch gpu.shape {
	case PRIMITIVE_POLYGON:
		gpu.ProcessPolygonCommand()
	case PRIMITIVE_RECTANGLE:
		gpu.ProcessRectangleCommand()
	}

	gpu.mode = MODE_NORMAL
}

func (gpu *GPU) GP0DoTransferToVRAM() {
	// 2nd  Destination Coord (YyyyXxxxh)  ;Xpos counted in halfwords
	gpu.startX = int(gpu.fifo.buffer[1] & 0xffff)
	gpu.startY = int(gpu.fifo.buffer[1] >> 16)

	// 3rd  Width+Height      (YsizXsizh)  ;Xsiz counted in halfwords
	resolution := gpu.fifo.buffer[2]
	gpu.imgWidth = int(resolution & 0xffff)
	gpu.imgHeight = int(resolution >> 16)
	size := gpu.imgWidth * gpu.imgHeight

	if size%2 == 1 {
		// must be even otherwise round up with 16 bit padding since cpu transfer 32 bit data
		size += 1
	}

	// each pixel in vram is 16 bit in size and each word is 32 bit in size so divide by 2
	gpu.wordsLeft = size / 2

	gpu.imgX = 0
	gpu.imgY = 0

	gpu.mode = MODE_CPUtoVRamBlit
}

func (gpu *GPU) GP0DoTransferFromVRAM() {
	// 2nd  Source Coord      (YyyyXxxxh) ; write to GP0 port (as usually)
	gpu.startX = int(gpu.fifo.buffer[1] & 0xffff)
	gpu.startY = int(gpu.fifo.buffer[1] >> 16)

	// 3rd  Width+Height      (YsizXsizh) ;/
	resolution := gpu.fifo.buffer[2]
	gpu.imgWidth = int(resolution & 0xffff)
	gpu.imgHeight = int(resolution >> 16)
	size := gpu.imgWidth * gpu.imgHeight

	if size%2 == 1 {
		// must be even otherwise round up with 16 bit padding since cpu transfer 32 bit data
		size += 1
	}

	// each pixel in vram is 16 bit in size and each word is 32 bit in size so divide by 2
	gpu.wordsLeft = size / 2

	gpu.imgX = 0
	gpu.imgY = 0

	gpu.mode = MODE_VramtoCPUBlit
}

func (gpu *GPU) GP0FillVRam() {
	var colour uint32 = 0

	r := GetRange(gpu.fifo.buffer[0], 0, 8)
	g := GetRange(gpu.fifo.buffer[0], 8, 8)
	b := GetRange(gpu.fifo.buffer[0], 16, 8)

	PackRange(&colour, 0, r>>3, 5)
	PackRange(&colour, 5, g>>3, 5)
	PackRange(&colour, 10, b>>3, 5)

	x := int(gpu.fifo.buffer[1] & 0xffff)
	y := int(gpu.fifo.buffer[1] >> 16)

	resolution := gpu.fifo.buffer[2]
	w := int(resolution & 0xffff)
	h := int(resolution >> 16)

	// position and dimensions must be within vram boundaries; also x and width are in 16-pixel (32-bytes) units (steps of 10h)
	startX := x & 0x3f0
	startY := y & 0x1ff

	width := (w + 0xf) & 0x3f0 // round up to a multiple of 0x10
	height := h & 0x1ff

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			xpos := Modulo(startX+x, VRAM_WIDTH)
			ypos := Modulo(startY+y, VRAM_HEIGHT)
			gpu.vram.Write16(xpos, ypos, uint16(colour))
		}
	}

	gpu.mode = MODE_NORMAL
}
