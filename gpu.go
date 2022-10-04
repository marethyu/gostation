package main

import "fmt"

const (
	GPU_OFFSET = 0x1f801810
	GPU_SIZE   = 8
)

const (
	SEMI_TRANSPARENT_MODE0 = iota /* B/2+F/2 */
	SEMI_TRANSPARENT_MODE1        /* B+F */
	SEMI_TRANSPARENT_MODE2        /* B-F */
	SEMI_TRANSPARENT_MODE3        /* B+F/4 */
)

const (
	TEXTURE_FORMAT_4b = iota
	TEXTURE_FORMAT_8b
	TEXTURE_FORMAT_15b
	TEXTURE_FORMAT_reserved
)

const (
	HORIZ_RES_256 = iota
	HORIZ_RES_320
	HORIZ_RES_512
	HORIZ_RES_640
)

const (
	DMA_DIR_OFF = iota
	DMA_DIR_FIFO
	DMA_DIR_CPUtoGP0
	DMA_DIR_GPUREADtoCPU
)

const (
	MODE_NORMAL = iota
	MODE_RENDERING
	MODE_CPUtoVRamBlit
	MODE_VramtoCPUBlit
	MODE_FillVRam
)

const (
	PRIMITIVE_POLYGON = iota
	PRIMITIVE_RECTANGLE
)

const (
	VCYCLES_PER_SCANLINE_PAL  = 3406
	VCYCLES_PER_SCANLINE_NTSC = 3413
	SCANLINES_PER_FRAME_PAL   = 314
	SCANLINES_PER_FRAME_NTSC  = 263
)

/* whut the forking fock why itz long */
type GPU struct {
	Core *GoStation

	/*
		Information for the most of the fields can be found here:
		https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gpu-rendering-attributes
		https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gpu-display-control-commands-gp1
	*/

	/* 1F801814h - GPUSTAT - GPU Status Register (R) */
	txBase             int  /* 0-3   Texture page X Base   (N*64)                              ;GP0(E1h).0-3 */
	tyBase             int  /* 4     Texture page Y Base   (N*256) (ie. 0 or 256)              ;GP0(E1h).4 */
	semiTransparency   int  /* 5-6   Semi Transparency     (0=B/2+F/2, 1=B+F, 2=B-F, 3=B+F/4)  ;GP0(E1h).5-6 */
	textureFormat      int  /* 7-8   Texture page colors   (0=4bit, 1=8bit, 2=15bit, 3=Reserved)GP0(E1h).7-8 */
	dilthering         bool /* 9     Dither 24bit to 15bit (0=Off/strip LSBs, 1=Dither Enabled);GP0(E1h).9 */
	drawToDisplay      bool /* 10    Drawing to display area (0=Prohibited, 1=Allowed)         ;GP0(E1h).10 */
	setMaskBit         bool /* 11    Set Mask-bit when drawing pixels (0=No, 1=Yes/Mask)       ;GP0(E6h).0 */
	drawUnmaskedPixels bool /* 12    Draw Pixels           (0=Always, 1=Not to Masked areas)   ;GP0(E6h).1 */
	interlace          bool /* 13    Interlace Field       (or, always 1 when GP1(08h).5=0) */
	reverseFlag        bool /* 14    "Reverseflag"         (0=Normal, 1=Distorted)             ;GP1(08h).7 */
	textureDisable     bool /* 15    Texture Disable       (0=Normal, 1=Disable Textures)      ;GP0(E1h).11 */
	/* hr2 16    Horizontal Resolution 2     (0=256/320/512/640, 1=368)    ;GP1(08h).6
	   hr1 17-18 Horizontal Resolution 1     (0=256, 1=320, 2=512, 3=640)  ;GP1(08h).0-1 */
	hr2                bool
	hr1                uint8
	horizResolution    uint32
	vertRes            bool /* 19    Vertical Resolution         (0=240, 1=480, when Bit22=1)  ;GP1(08h).2 */
	vertResolution     uint32
	PALMode            bool /* 20    Video Mode                  (0=NTSC/60Hz, 1=PAL/50Hz)     ;GP1(08h).3 */
	displayColourDepth bool /* 21    Display Area Color Depth    (0=15bit, 1=24bit)            ;GP1(08h).4 */
	verticalInterlace  bool /* 22    Vertical Interlace          (0=Off, 1=On)                 ;GP1(08h).5 */
	displayDisable     bool /* 23    Display Enable              (0=Enabled, 1=Disabled)       ;GP1(03h).0 */
	irq                bool /* 24    Interrupt Request (IRQ1)    (0=Off, 1=IRQ)       ;GP0(1Fh)/GP1(02h) */
	/*
		25    DMA / Data Request, meaning depends on GP1(04h) DMA Direction:
		When GP1(04h)=0 ---> Always zero (0)
		When GP1(04h)=1 ---> FIFO State  (0=Full, 1=Not Full)
		When GP1(04h)=2 ---> Same as GPUSTAT.28
		When GP1(04h)=3 ---> Same as GPUSTAT.27
	*/
	dma             bool
	readyReceiveCmd bool /* 26    Ready to receive Cmd Word   (0=No, 1=Ready)  ;GP0(...) ;via GP0 */
	readySendVRam   bool /* 27    Ready to send VRAM to CPU   (0=No, 1=Ready)  ;GP0(C0h) ;via GPUREAD */
	readyReceiveDMA bool /* 28    Ready to receive DMA Block  (0=No, 1=Ready)  ;GP0(...) ;via GP0 */
	dmaDirection    int  /* 29-30 DMA Direction (0=Off, 1=?, 2=CPUtoGP0, 3=GPUREADtoCPU)    ;GP1(04h).0-1 */
	interlaceOdd    bool /* 31    Drawing even/odd lines in interlace mode (0=Even or Vblank, 1=Odd) */

	/* GP0(E1h) - Draw Mode setting (aka "Texpage") */
	rectTextureXFlip bool /* mirror texture rectangle along the x-axis    ;GP0(E1h).12 */
	rectTextureYFlip bool /* mirror texture rectangle along the y-axis    ;GP0(E1h).13 */

	/* GP0(E2h) - Texture Window setting */
	texWindowMaskX   int /* 0-4    Texture window Mask X   (in 8 pixel steps) */
	texWindowMaskY   int /* 5-9    Texture window Mask Y   (in 8 pixel steps) */
	texWindowOffsetX int /* 10-14  Texture window Offset X (in 8 pixel steps) */
	texWindowOffsetY int /* 15-19  Texture window Offset Y (in 8 pixel steps) */

	/* GP0(E3h) - Set Drawing Area top left (X1,Y1) */
	drawingAreaX1 int /* 0-9    X-coordinate (0..1023) */
	drawingAreaY1 int /* 10-18  Y-coordinate (0..511) or 10-19  Y-coordinate (0..1023)? */

	/* GP0(E4h) - Set Drawing Area bottom right (X2,Y2) */
	drawingAreaX2 int
	drawingAreaY2 int

	/* GP0(E5h) - Set Drawing Offset (X,Y) */
	drawingXOffset int /* 0-10   X-offset (-1024..+1023) (usually within X1,X2 of Drawing Area) */
	drawingYOffset int /* 11-21  Y-offset (-1024..+1023) (usually within Y1,Y2 of Drawing Area) */

	/* GP1(05h) - Start of Display area (in VRAM) */
	displayVramStartX int /* 0-9   X (0-1023)    (halfword address in VRAM)  (relative to begin of VRAM) */
	displayVramStartY int /* 10-18 Y (0-511)     (scanline number in VRAM)   (relative to begin of VRAM) */

	/* GP1(06h) - Horizontal Display range (on Screen) */
	displayHorizX1x7 uint32 /* 0-11   X1 (260h+0)       ;12bit       ;\counted in video clock units, */
	displayHorizX2x7 uint32 /* 12-23  X2 (260h+320*8)   ;12bit       ;/relative to HSYNC (multiplied by 7) */

	/* GP1(07h) - Vertical Display range (on Screen) */
	displayVertY1 uint32 /* 0-9   Y1 (NTSC=88h-(240/2), (PAL=A3h-(288/2))  ;\scanline numbers on screen, */
	displayVertY2 uint32 /* 10-19 Y2 (NTSC=88h+(240/2), (PAL=A3h+(288/2))  ;/relative to VSYNC */

	vram *VRAM

	mode       int
	fifo       *FIFO[uint32]
	fifoActive bool

	/* for rendering commands */
	shape      int    /* what shape to render when FIFO finished collecting all args */
	shape_attr uint32 /* rendering attributes */

	/* for cpu to vram blit */
	startX    int /* destination x (ie. starting position on vram in this context) */
	startY    int /* destination y */
	imgWidth  int /* width of "image" */
	imgHeight int /* height of "image" */
	imgX      int /* x coordinate relative to "image" being transferred */
	imgY      int /* y coordinate relative to "image" being transferred */
	wordsLeft int

	gpuReadVal uint32

	/* timing stuff
	   since the video clock is the cpu clock multiplied by 11/7 and
	   we want to avoid division as much as possible in each gpu tick so we need to multiply video cycles by 7.
	   also it provides more accuracy. */
	videoCyclesx7            uint32
	scanline                 uint32
	videoCyclesPerScanlinex7 uint32 /* 3406*7 (3413*7 in NTSC mode) */
	scanlinesPerFrame        uint32 /* 263 (314 in PAL mode) */
	vblank                   bool   /* currently in vblank? */
}

func NewGPU(core *GoStation) *GPU {
	return &GPU{
		core,
		0,
		0,
		0,
		0,
		false,
		false,
		false,
		false,
		true,
		false,
		false,
		false,
		0,
		256,
		false,
		240,
		false,
		false,
		false,
		true,
		false,
		false,
		false,
		false,
		false,
		0,
		false,
		false,
		false,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		608 * 7,  // 260h
		3168 * 7, // 260h+320*8
		16,       // 88h-(240/2)
		256,      // 88h+(240/2)
		NewVRAM(),
		MODE_NORMAL,
		NewFIFO[uint32](),
		false,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		VCYCLES_PER_SCANLINE_NTSC,
		SCANLINES_PER_FRAME_NTSC,
		false,
	}
}

func (gpu *GPU) Contains(address uint32) bool {
	return address >= GPU_OFFSET && address < (GPU_OFFSET+GPU_SIZE)
}

func (gpu *GPU) InHblank() bool {
	return gpu.videoCyclesx7 < gpu.displayHorizX1x7 || gpu.videoCyclesx7 >= gpu.displayHorizX2x7
}

func (gpu *GPU) InVblank() bool {
	return gpu.scanline < gpu.displayVertY1 || gpu.scanline >= gpu.displayVertY2
}

func (gpu *GPU) Step(cpuCycles uint32) {
	gpu.videoCyclesx7 += cpuCycles * 11

	if gpu.videoCyclesx7 >= gpu.videoCyclesPerScanlinex7 {
		gpu.videoCyclesx7 -= gpu.videoCyclesPerScanlinex7
		gpu.scanline += 1

		inVblank := gpu.InVblank()
		if !gpu.vblank && inVblank {
			// trigger on rising edge
			gpu.Core.Interrupts.Request(IRQ_VBLANK)
		}
		gpu.vblank = inVblank

		if gpu.scanline == gpu.scanlinesPerFrame {
			gpu.scanline = 0
		}
	}
}

/* some fields are hardcoded for now */
func (gpu *GPU) GPUSTATUS() uint32 {
	var status uint32 = 0

	PackRange(&status, 0, uint32(gpu.txBase), 4)
	ModifyBit(&status, 4, gpu.tyBase != 0)
	PackRange(&status, 5, uint32(gpu.semiTransparency), 2)
	PackRange(&status, 7, uint32(gpu.textureFormat), 2)
	ModifyBit(&status, 9, gpu.dilthering)
	ModifyBit(&status, 10, gpu.drawToDisplay)
	ModifyBit(&status, 11, gpu.setMaskBit)
	ModifyBit(&status, 12, gpu.drawUnmaskedPixels)
	ModifyBit(&status, 13, gpu.interlace)
	ModifyBit(&status, 14, false)
	ModifyBit(&status, 15, gpu.textureDisable)
	ModifyBit(&status, 16, gpu.hr2)
	PackRange(&status, 17, uint32(gpu.hr1), 2)
	// Fuck infinite loops
	// ModifyBit(&status, 19, gpu.vertRes)
	ModifyBit(&status, 20, gpu.PALMode)
	ModifyBit(&status, 21, gpu.displayColourDepth)
	ModifyBit(&status, 22, gpu.verticalInterlace)
	ModifyBit(&status, 23, gpu.displayDisable)
	ModifyBit(&status, 24, gpu.irq)

	switch gpu.dmaDirection {
	case DMA_DIR_OFF:
		ModifyBit(&status, 25, false) // Always zero (0)
	case DMA_DIR_FIFO:
		ModifyBit(&status, 25, true) // FIFO State  (0=Full, 1=Not Full)
	case DMA_DIR_CPUtoGP0:
		ModifyBit(&status, 25, true) // Same as GPUSTAT.28
	case DMA_DIR_GPUREADtoCPU:
		ModifyBit(&status, 25, true) // Same as GPUSTAT.27
	}

	ModifyBit(&status, 26, true)
	ModifyBit(&status, 27, true)
	ModifyBit(&status, 28, true)
	PackRange(&status, 29, uint32(gpu.dmaDirection), 2)
	ModifyBit(&status, 31, false)

	return status
}

func (gpu *GPU) GPUREAD() uint32 {
	if gpu.mode == MODE_VramtoCPUBlit {
		lo := uint32(gpu.GP0DoVramToCPUTransfer())
		hi := uint32(gpu.GP0DoVramToCPUTransfer())

		gpu.wordsLeft -= 1

		if gpu.wordsLeft == 0 {
			gpu.mode = MODE_NORMAL
		}

		return (hi << 16) | lo
	}

	return gpu.gpuReadVal
}

func (gpu *GPU) Read32(address uint32) uint32 {
	switch address {
	case 0x1f801810:
		return gpu.GPUREAD()
	case 0x1f801814:
		return gpu.GPUSTATUS()
	default:
		panic(fmt.Sprintf("[GPU::Read32] Invalid address: %x", address))
	}
}

/* Nice summary here: https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gpu-command-summary */
func (gpu *GPU) GP0(data uint32) {
	if gpu.fifoActive {
		gpu.fifo.Push(data)

		if gpu.fifo.Done() {
			switch gpu.mode {
			case MODE_RENDERING:
				gpu.GP0RenderPrimitive()
			case MODE_CPUtoVRamBlit:
				gpu.GP0DoTransferToVRAM()
			case MODE_VramtoCPUBlit:
				gpu.GP0DoTransferFromVRAM()
			case MODE_FillVRam:
				gpu.GP0FillVRam()
			case MODE_NORMAL:
				panic("[GPU::GP0] normal mode???")
			}

			gpu.fifoActive = false
		}

		return
	}

	if gpu.mode == MODE_CPUtoVRamBlit {
		lo := uint16(data & 0xffff)
		hi := uint16(data >> 16)

		gpu.GP0DoCPUToVramTransfer(lo)
		gpu.GP0DoCPUToVramTransfer(hi)

		gpu.wordsLeft -= 1

		if gpu.wordsLeft == 0 {
			gpu.mode = MODE_NORMAL
		}

		return
	}

	op := GetRange(data, 29, 3) // top 3 bits of a command

	switch op {
	case 0b000:
		gpu.GP0ExecuteMiscCommand(data)
	case 0b001:
		gpu.GP0InitRenderPolygonCommand(data)
	case 0b011:
		gpu.GP0InitRenderRectangleCommand(data)
	case 0b101:
		gpu.GP0InitCPUToVRamBlit(data)
	case 0b110:
		gpu.GP0InitVramToCPUBlit(data)
	case 0b111:
		gpu.GP0ExecuteEnvironmentCommand(data)
	default:
		panic(fmt.Sprintf("[GPU::GP0] Unknown command: %x", data))
	}
}

func (gpu *GPU) GP1(data uint32) {
	op := data >> 24 // the most significant byte determines what command

	switch op {
	case 0x00:
		gpu.GP1Reset()
	case 0x01:
		gpu.GP1ResetCommandBuffer()
	case 0x02:
		gpu.GP1AcknowledgeInterrupt()
	case 0x03:
		gpu.GP1DisplayEnableSet(data)
	case 0x04:
		gpu.GP1DMADirectionSet(data)
	case 0x05:
		gpu.GP1DisplayVRamStartSet(data)
	case 0x06:
		gpu.GP1HorizDisplayRangeSet(data)
	case 0x07:
		gpu.GP1VertDisplayRangeSet(data)
	case 0x08:
		gpu.GP1DisplayModeSet(data)
	case 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F:
		gpu.GP1GPUInfo(data)
	default:
		panic(fmt.Sprintf("[GPU::GP1] Unknown command: %x", data))
	}
}

func (gpu *GPU) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801810:
		gpu.GP0(data)
	case 0x1f801814:
		gpu.GP1(data)
	}
}
