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
	gpu.dmaDirection = DMA_DIR_OFF
	// GP1(05h)      ;display address (0)
	gpu.displayVramStartX = 0
	gpu.displayVramStartY = 0
	// GP1(06h)      ;display x1,x2 (x1=200h, x2=200h+256*10)
	gpu.displayHorizX1 = 0x200
	gpu.displayHorizX2 = 0x200 + 256*10
	// GP1(07h)      ;display y1,y2 (y1=010h, y2=010h+240)
	gpu.displayVertY1 = 0x010
	gpu.displayVertY2 = 0x010 + 240
	// GP1(08h)      ;display mode 320x200 NTSC (0)
	gpu.hr1 = 0
	gpu.vertRes = false
	gpu.vertResolution = 240
	gpu.PALMode = false
	gpu.displayColourDepth = false
	gpu.verticalInterlace = false
	gpu.interlace = true
	gpu.hr2 = false
	gpu.horizResolution = 256
	gpu.reverseFlag = false

	/* GP0(E1h..E6h) ;rendering attributes (0) */

	// GP0(E1h)
	gpu.txBase = 0
	gpu.tyBase = 0
	gpu.semiTransparency = 0
	gpu.textureFormat = 0
	gpu.dilthering = false
	gpu.drawToDisplay = false
	gpu.textureDisable = false
	gpu.rectTextureXFlip = false
	gpu.rectTextureYFlip = false

	// GP0(E2h)
	gpu.texWindowMaskX = 0
	gpu.texWindowMaskY = 0
	gpu.texWindowOffsetX = 0
	gpu.texWindowOffsetY = 0

	// GP0(E3h)
	gpu.drawingAreaX1 = 0
	gpu.drawingAreaY1 = 0

	// GP0(E4h)
	gpu.drawingAreaX2 = 0
	gpu.drawingAreaY2 = 0

	// GP0(E5h)
	gpu.drawingXOffset = 0
	gpu.drawingYOffset = 0

	// GP0(E6h)
	gpu.setMaskBit = false
	gpu.drawUnmaskedPixels = false
}
