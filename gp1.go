package main

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp100h-reset-gpu

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
	gpu.GP1ResetCommandBuffer()
	// GP1(02h)      ;ack irq (0)
	gpu.GP1AcknowledgeInterrupt()
	// GP1(03h)      ;display off (1)
	gpu.GP1DisplayEnableSet(1)
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

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp101h-reset-command-buffer

0-23  Not used (zero)
*/
func (gpu *GPU) GP1ResetCommandBuffer() {
	gpu.fifo.Reset(16)
	gpu.mode = MODE_NORMAL
	// TODO clut cache
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp102h-acknowledge-gpu-interrupt-irq1

0-23  Not used (zero)                                        ;GPUSTAT.24
*/
func (gpu *GPU) GP1AcknowledgeInterrupt() {
	gpu.irq = false // reset
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp103h-display-enable

0     Display On/Off   (0=On, 1=Off)                         ;GPUSTAT.23
1-23  Not used (zero)
*/
func (gpu *GPU) GP1DisplayEnableSet(data uint32) {
	gpu.displayDisable = TestBit(data, 0)
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp104h-dma-direction-data-request

0-1  DMA Direction (0=Off, 1=FIFO, 2=CPUtoGP0, 3=GPUREADtoCPU) ;GPUSTAT.29-30
2-23 Not used (zero)
*/
func (gpu *GPU) GP1DMADirectionSet(data uint32) {
	gpu.dmaDirection = int(GetRange(data, 0, 2))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp105h-start-of-display-area-in-vram

0-9   X (0-1023)    (halfword address in VRAM)  (relative to begin of VRAM)
10-18 Y (0-511)     (scanline number in VRAM)   (relative to begin of VRAM)
19-23 Not used (zero)
*/
func (gpu *GPU) GP1DisplayVRamStartSet(data uint32) {
	gpu.displayVramStartX = int(GetRange(data, 0, 10) & 0b1111111110) // ignore the LSB to align with 16 bit pixels
	gpu.displayVramStartY = int(GetRange(data, 10, 9))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp106h-horizontal-display-range-on-screen

0-11   X1 (260h+0)       ;12bit       ;\counted in video clock units,
12-23  X2 (260h+320*8)   ;12bit       ;/relative to HSYNC
*/
func (gpu *GPU) GP1HorizDisplayRangeSet(data uint32) {
	gpu.displayHorizX1 = int(GetRange(data, 0, 12))
	gpu.displayHorizX2 = int(GetRange(data, 12, 12))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp107h-vertical-display-range-on-screen

0-9   Y1 (NTSC=88h-(240/2), (PAL=A3h-(288/2))  ;\scanline numbers on screen,
10-19 Y2 (NTSC=88h+(240/2), (PAL=A3h+(288/2))  ;/relative to VSYNC
20-23 Not used (zero)
*/
func (gpu *GPU) GP1VertDisplayRangeSet(data uint32) {
	gpu.displayVertY1 = int(GetRange(data, 0, 10))
	gpu.displayVertY2 = int(GetRange(data, 10, 10))
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp108h-display-mode

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
	gpu.hr1 = uint8(GetRange(data, 0, 2))
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

	if gpu.PALMode {
		gpu.videoCyclesPerScanline = VCYCLES_PER_SCANLINE_PAL
		gpu.scanlinesPerFrame = SCANLINES_PER_FRAME_PAL
		gpu.Core.cyclesPerFrame = CPU_CYCLES_PER_SEC / 50
	} else {
		gpu.videoCyclesPerScanline = VCYCLES_PER_SCANLINE_NTSC
		gpu.scanlinesPerFrame = SCANLINES_PER_FRAME_NTSC
		gpu.Core.cyclesPerFrame = CPU_CYCLES_PER_SEC / 60
	}
}

/*
	https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gp110h-get-gpu-info

00h-01h = Returns Nothing (old value in GPUREAD remains unchanged)
02h     = Read Texture Window setting  ;GP0(E2h) ;20bit/MSBs=Nothing
03h     = Read Draw area top left      ;GP0(E3h) ;20bit/MSBs=Nothing
04h     = Read Draw area bottom right  ;GP0(E4h) ;20bit/MSBs=Nothing
05h     = Read Draw offset             ;GP0(E5h) ;22bit
06h     = Returns Nothing (old value in GPUREAD remains unchanged)
07h     = Read GPU Type (usually 2)    ;see "GPU Versions" chapter
08h     = Unknown (Returns 00000000h) (lightgun on some GPUs?)
09h-0Fh = Returns Nothing (old value in GPUREAD remains unchanged)
10h-FFFFFFh = Mirrors of 00h..0Fh
*/
func (gpu *GPU) GP1GPUInfo(data uint32) {
	switch data & 0xf {
	case 0x2:
		PackRange(&gpu.gpuReadVal, 0, uint32(gpu.texWindowMaskX), 5)
		PackRange(&gpu.gpuReadVal, 5, uint32(gpu.texWindowMaskY), 5)
		PackRange(&gpu.gpuReadVal, 10, uint32(gpu.texWindowOffsetX), 5)
		PackRange(&gpu.gpuReadVal, 15, uint32(gpu.texWindowOffsetY), 5)
		PackRange(&gpu.gpuReadVal, 20, 0, 12)
	case 0x3:
		PackRange(&gpu.gpuReadVal, 0, uint32(gpu.drawingAreaX1), 10)
		PackRange(&gpu.gpuReadVal, 10, uint32(gpu.drawingAreaY1), 10)
		PackRange(&gpu.gpuReadVal, 20, 0, 12)
	case 0x4:
		PackRange(&gpu.gpuReadVal, 0, uint32(gpu.drawingAreaX2), 10)
		PackRange(&gpu.gpuReadVal, 10, uint32(gpu.drawingAreaY2), 10)
		PackRange(&gpu.gpuReadVal, 20, 0, 12)
	case 0x5:
		PackRange(&gpu.gpuReadVal, 0, uint32(gpu.drawingXOffset), 11)
		PackRange(&gpu.gpuReadVal, 11, uint32(gpu.drawingYOffset), 11)
	case 0x7:
		gpu.gpuReadVal = 2
	case 0x8:
		gpu.gpuReadVal = 0
	}
}
