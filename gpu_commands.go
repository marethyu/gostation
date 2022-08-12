package main

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e1h-draw-mode-setting-aka-texpage

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
	gpu.txBase = GetValue(data, 0, 4)
	gpu.tyBase = GetValue(data, 4, 1)
	gpu.semiTransparency = uint8(GetValue(data, 5, 2))
	gpu.textureFormat = uint8(GetValue(data, 7, 2))
	gpu.dilthering = TestBit(data, 9)
	gpu.drawToDisplay = TestBit(data, 10)
	gpu.textureDisable = TestBit(data, 11)
	gpu.rectTextureXFlip = TestBit(data, 12)
	gpu.rectTextureYFlip = TestBit(data, 13)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e2h-texture-window-setting

0-4    Texture window Mask X   (in 8 pixel steps)
5-9    Texture window Mask Y   (in 8 pixel steps)
10-14  Texture window Offset X (in 8 pixel steps)
15-19  Texture window Offset Y (in 8 pixel steps)
20-23  Not used (zero)
24-31  Command  (E2h)

*/
func (gpu *GPU) GP0TextureWindowSetup(data uint32) {
	gpu.texWindowMaskX = GetValue(data, 0, 5)
	gpu.texWindowMaskY = GetValue(data, 5, 5)
	gpu.texWindowOffsetX = GetValue(data, 10, 5)
	gpu.texWindowOffsetY = GetValue(data, 15, 5)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e3h-set-drawing-area-top-left-x1y1

0-9    X-coordinate (0..1023)
10-18  Y-coordinate (0..511)   ;\on Old 160pin GPU (max 1MB VRAM)
19-23  Not used (zero)         ;/
10-19  Y-coordinate (0..1023)  ;\on New 208pin GPU (max 2MB VRAM)
20-23  Not used (zero)         ;/(retail consoles have only 1MB though)
24-31  Command  (Exh)

*/
func (gpu *GPU) GP0DrawingAreaTopLeftSet(data uint32) {
	gpu.drawingAreaX1 = GetValue(data, 0, 10)
	gpu.drawingAreaY1 = GetValue(data, 10, 10)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e4h-set-drawing-area-bottom-right-x2y2

0-9    X-coordinate (0..1023)
10-18  Y-coordinate (0..511)   ;\on Old 160pin GPU (max 1MB VRAM)
19-23  Not used (zero)         ;/
10-19  Y-coordinate (0..1023)  ;\on New 208pin GPU (max 2MB VRAM)
20-23  Not used (zero)         ;/(retail consoles have only 1MB though)
24-31  Command  (Exh)

*/
func (gpu *GPU) GP0DrawingAreaBottomRightSet(data uint32) {
	gpu.drawingAreaX2 = GetValue(data, 0, 10)
	gpu.drawingAreaY2 = GetValue(data, 10, 10)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e5h-set-drawing-offset-xy

0-10   X-offset (-1024..+1023) (usually within X1,X2 of Drawing Area)
11-21  Y-offset (-1024..+1023) (usually within Y1,Y2 of Drawing Area)
22-23  Not used (zero)
24-31  Command  (E5h)

*/
func (gpu *GPU) GP0DrawingOffsetSet(data uint32) {
	x := uint16(GetValue(data, 0, 11))
	y := uint16(GetValue(data, 11, 11))

	// x and y are 11 bit unsigned values so
	// in order to get successful sign extension, x and y must be shifted 5 bits left (making them 16 bit unsigned)
	// of course, shift them 5 bits right to get 11 bit signed which is what we want
	gpu.drawingXOffset = int32(int16(x<<5) >> 5)
	gpu.drawingYOffset = int32(int16(y<<5) >> 5)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp0e6h-mask-bit-setting

0     Set mask while drawing (0=TextureBit15, 1=ForceBit15=1)   ;GPUSTAT.11
1     Check mask before draw (0=Draw Always, 1=Draw if Bit15=0) ;GPUSTAT.12
2-23  Not used (zero)
24-31 Command  (E6h)

*/
func (gpu *GPU) GP0MaskBitSetup(data uint32) {
	gpu.setMaskBit = TestBit(data, 0)
	gpu.drawUnmaskedPixels = TestBit(data, 1)
}

/*
GP1 commands here
*/

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp100h-reset-gpu

GP1(01h)      ;clear fifo
GP1(02h)      ;ack irq (0)
GP1(03h)      ;display off (1)
GP1(04h)      ;dma off (0)
GP1(05h)      ;display address (0)
GP1(06h)      ;display x1,x2 (x1=200h, x2=200h+256*10)
GP1(07h)      ;display y1,y2 (y1=010h, y2=010h+240)
GP1(08h)      ;display mode 320x200 NTSC (0)
GP0(E1h..E6h) ;rendering attributes (0)

*/
func (gpu *GPU) GP1Reset() {
	// GP1(01h)      ;clear fifo
	// TODO
	// GP1(02h)      ;ack irq (0)
	gpu.irq = false
	// GP1(03h)      ;display off (1)
	gpu.displayDisable = true
	// GP1(04h)      ;dma off (0)
	gpu.GP1DMADirectionSet(DMA_DIR_OFF)
	// GP1(05h)      ;display address (0)
	gpu.GP1DisplayVRamStartSet(0)
	// GP1(06h)      ;display x1,x2 (x1=200h, x2=200h+256*10)
	gpu.GP1HorizDisplayRangeSet(0xc00200)
	// GP1(07h)      ;display y1,y2 (y1=010h, y2=010h+240)
	gpu.GP1VertDisplayRangeSet(0x40010)
	// GP1(08h)      ;display mode 320x200 NTSC (0)
	gpu.GP1DisplayModeSet(0)

	/* GP0(E1h..E6h) ;rendering attributes (0) */

	// GP0(E1h)
	gpu.GP0DrawModeSet(0)
	// GP0(E2h)
	gpu.GP0TextureWindowSetup(0)
	// GP0(E3h)
	gpu.GP0DrawingAreaTopLeftSet(0)
	// GP0(E4h)
	gpu.GP0DrawingAreaBottomRightSet(0)
	// GP0(E5h)
	gpu.GP0DrawingOffsetSet(0)
	// GP0(E6h)
	gpu.GP0MaskBitSetup(0)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp104h-dma-direction-data-request

0-1  DMA Direction (0=Off, 1=FIFO, 2=CPUtoGP0, 3=GPUREADtoCPU) ;GPUSTAT.29-30
2-23 Not used (zero)

*/
func (gpu *GPU) GP1DMADirectionSet(data uint32) {
	gpu.dmaDirection = uint8(GetValue(data, 0, 2))
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp105h-start-of-display-area-in-vram

0-9   X (0-1023)    (halfword address in VRAM)  (relative to begin of VRAM)
10-18 Y (0-511)     (scanline number in VRAM)   (relative to begin of VRAM)
19-23 Not used (zero)

*/
func (gpu *GPU) GP1DisplayVRamStartSet(data uint32) {
	gpu.displayVramStartX = GetValue(data, 0, 10) & 0b1111111110 // ignore the LSB to align with 16 bit pixels
	gpu.displayVramStartY = GetValue(data, 10, 9)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp106h-horizontal-display-range-on-screen

0-11   X1 (260h+0)       ;12bit       ;\counted in video clock units,
12-23  X2 (260h+320*8)   ;12bit       ;/relative to HSYNC

*/
func (gpu *GPU) GP1HorizDisplayRangeSet(data uint32) {
	gpu.displayHorizX1 = GetValue(data, 0, 12)
	gpu.displayHorizX2 = GetValue(data, 12, 12)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp107h-vertical-display-range-on-screen

0-9   Y1 (NTSC=88h-(240/2), (PAL=A3h-(288/2))  ;\scanline numbers on screen,
10-19 Y2 (NTSC=88h+(240/2), (PAL=A3h+(288/2))  ;/relative to VSYNC
20-23 Not used (zero)

*/
func (gpu *GPU) GP1VertDisplayRangeSet(data uint32) {
	gpu.displayVertY1 = GetValue(data, 0, 10)
	gpu.displayVertY2 = GetValue(data, 10, 10)
}

/* https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp108h-display-mode

0-1   Horizontal Resolution 1     (0=256, 1=320, 2=512, 3=640) ;GPUSTAT.17-18
2     Vertical Resolution         (0=240, 1=480, when Bit5=1)  ;GPUSTAT.19
3     Video Mode                  (0=NTSC/60Hz, 1=PAL/50Hz)    ;GPUSTAT.20
4     Display Area Color Depth    (0=15bit, 1=24bit)           ;GPUSTAT.21
5     Vertical Interlace          (0=Off, 1=On)                ;GPUSTAT.22
6     Horizontal Resolution 2     (0=256/320/512/640, 1=368)   ;GPUSTAT.16
7     "Reverseflag"               (0=Normal, 1=Distorted)      ;GPUSTAT.14
8-23  Not used (zero)

*/
func (gpu *GPU) GP1DisplayModeSet(data uint32) {
	gpu.hr1 = uint8(GetValue(data, 0, 2))
	gpu.vertRes = TestBit(data, 2)
	gpu.PALMode = TestBit(data, 3)
	gpu.displayColourDepth = TestBit(data, 4)
	gpu.verticalInterlace = TestBit(data, 5)
	gpu.hr2 = TestBit(data, 6)
	gpu.reverseFlag = TestBit(data, 7)

	if gpu.verticalInterlace {
		// force set vertRes to 1 if verticalInterlace=true
		gpu.vertRes = true
	}

	if gpu.vertRes {
		gpu.vertResolution = 480
	} else {
		gpu.vertResolution = 240
	}

	if gpu.hr2 {
		gpu.horizResolution = 368
	} else {
		switch gpu.hr1 {
		case HORIZ_RES_256:
			gpu.horizResolution = 256
		case HORIZ_RES_320:
			gpu.horizResolution = 320
		case HORIZ_RES_512:
			gpu.horizResolution = 512
		case HORIZ_RES_640:
			gpu.horizResolution = 640
		}
	}

	if !gpu.verticalInterlace {
		gpu.interlace = true
	}
}