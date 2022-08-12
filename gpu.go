package main

import (
	"fmt"
)

const (
	GPU_OFFSET = 0x1f801810
	GPU_SIZE   = 8
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

/* whut the forking fock why itz long */
type GPU struct {
	Core *GoStationCore

	/*
		Information for the most of the fields can be found here:
		https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gpu-rendering-attributes
		https://psx-spx.consoledev.net/graphicsprocessingunitgpu/#gpu-display-control-commands-gp1
	*/

	/* 1F801814h - GPUSTAT - GPU Status Register (R) */
	txBase             uint32 /* 0-3   Texture page X Base   (N*64)                              ;GP0(E1h).0-3 */
	tyBase             uint32 /* 4     Texture page Y Base   (N*256) (ie. 0 or 256)              ;GP0(E1h).4 */
	semiTransparency   uint8  /* 5-6   Semi Transparency     (0=B/2+F/2, 1=B+F, 2=B-F, 3=B+F/4)  ;GP0(E1h).5-6 */
	textureFormat      uint8  /* 7-8   Texture page colors   (0=4bit, 1=8bit, 2=15bit, 3=Reserved)GP0(E1h).7-8 */
	dilthering         bool   /* 9     Dither 24bit to 15bit (0=Off/strip LSBs, 1=Dither Enabled);GP0(E1h).9 */
	drawToDisplay      bool   /* 10    Drawing to display area (0=Prohibited, 1=Allowed)         ;GP0(E1h).10 */
	setMaskBit         bool   /* 11    Set Mask-bit when drawing pixels (0=No, 1=Yes/Mask)       ;GP0(E6h).0 */
	drawUnmaskedPixels bool   /* 12    Draw Pixels           (0=Always, 1=Not to Masked areas)   ;GP0(E6h).1 */
	interlace          bool   /* 13    Interlace Field       (or, always 1 when GP1(08h).5=0) */
	reverseFlag        bool   /* 14    "Reverseflag"         (0=Normal, 1=Distorted)             ;GP1(08h).7 */
	textureDisable     bool   /* 15    Texture Disable       (0=Normal, 1=Disable Textures)      ;GP0(E1h).11 */
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
	readyReceiveCmd bool  /* 26    Ready to receive Cmd Word   (0=No, 1=Ready)  ;GP0(...) ;via GP0 */
	readySendVRam   bool  /* 27    Ready to send VRAM to CPU   (0=No, 1=Ready)  ;GP0(C0h) ;via GPUREAD */
	readyReceiveDMA bool  /* 28    Ready to receive DMA Block  (0=No, 1=Ready)  ;GP0(...) ;via GP0 */
	dmaDirection    uint8 /* 29-30 DMA Direction (0=Off, 1=?, 2=CPUtoGP0, 3=GPUREADtoCPU)    ;GP1(04h).0-1 */
	interlaceOdd    bool  /* 31    Drawing even/odd lines in interlace mode (0=Even or Vblank, 1=Odd) */

	/* GP0(E1h) - Draw Mode setting (aka "Texpage") */
	rectTextureXFlip bool /* mirror texture rectangle along the x-axis    ;GP0(E1h).12 */
	rectTextureYFlip bool /* mirror texture rectangle along the y-axis    ;GP0(E1h).13 */

	/* GP0(E2h) - Texture Window setting */
	texWindowMaskX   uint32 /* 0-4    Texture window Mask X   (in 8 pixel steps) */
	texWindowMaskY   uint32 /* 5-9    Texture window Mask Y   (in 8 pixel steps) */
	texWindowOffsetX uint32 /* 10-14  Texture window Offset X (in 8 pixel steps) */
	texWindowOffsetY uint32 /* 15-19  Texture window Offset Y (in 8 pixel steps) */

	/* GP0(E3h) - Set Drawing Area top left (X1,Y1) */
	drawingAreaX1 uint32 /* 0-9    X-coordinate (0..1023) */
	drawingAreaY1 uint32 /* 10-18  Y-coordinate (0..511) or 10-19  Y-coordinate (0..1023)? */

	/* GP0(E4h) - Set Drawing Area bottom right (X2,Y2) */
	drawingAreaX2 uint32
	drawingAreaY2 uint32

	/* GP0(E5h) - Set Drawing Offset (X,Y) */
	drawingXOffset int32 /* 0-10   X-offset (-1024..+1023) (usually within X1,X2 of Drawing Area) */
	drawingYOffset int32 /* 11-21  Y-offset (-1024..+1023) (usually within Y1,Y2 of Drawing Area) */

	/* GP1(05h) - Start of Display area (in VRAM) */
	displayVramStartX uint32 /* 0-9   X (0-1023)    (halfword address in VRAM)  (relative to begin of VRAM) */
	displayVramStartY uint32 /* 10-18 Y (0-511)     (scanline number in VRAM)   (relative to begin of VRAM) */

	/* GP1(06h) - Horizontal Display range (on Screen) */
	displayHorizX1 uint32 /* 0-11   X1 (260h+0)       ;12bit       ;\counted in video clock units, */
	displayHorizX2 uint32 /* 12-23  X2 (260h+320*8)   ;12bit       ;/relative to HSYNC */

	/* GP1(07h) - Vertical Display range (on Screen) */
	displayVertY1 uint32 /* 0-9   Y1 (NTSC=88h-(240/2), (PAL=A3h-(288/2))  ;\scanline numbers on screen, */
	displayVertY2 uint32 /* 10-19 Y2 (NTSC=88h+(240/2), (PAL=A3h+(288/2))  ;/relative to VSYNC */
}

func NewGPU(core *GoStationCore) *GPU {
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
		0,
		0,
		0,
		0,
	}
}

func (gpu *GPU) Contains(address uint32) bool {
	return address >= GPU_OFFSET && address < (GPU_OFFSET+GPU_SIZE)
}

/* some fields are hardcoded for now */
func (gpu *GPU) ReadStatus() uint32 {
	var status uint32 = 0

	PackValue(&status, 0, uint32(gpu.txBase), 4)
	ModifyBit(&status, 4, gpu.tyBase != 0)
	PackValue(&status, 5, uint32(gpu.semiTransparency), 2)
	PackValue(&status, 7, uint32(gpu.textureFormat), 2)
	ModifyBit(&status, 9, gpu.dilthering)
	ModifyBit(&status, 10, gpu.drawToDisplay)
	ModifyBit(&status, 11, gpu.setMaskBit)
	ModifyBit(&status, 12, gpu.drawUnmaskedPixels)
	ModifyBit(&status, 13, gpu.interlace)
	ModifyBit(&status, 14, false)
	ModifyBit(&status, 15, gpu.textureDisable)
	ModifyBit(&status, 16, gpu.hr2)
	PackValue(&status, 17, uint32(gpu.hr1), 2)
	ModifyBit(&status, 19, gpu.vertRes)
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
	PackValue(&status, 29, uint32(gpu.dmaDirection), 2)
	ModifyBit(&status, 31, false)

	return status
}

func (gpu *GPU) ReadGPUREAD() uint32 {
	// TODO
	return 0
}

func (gpu *GPU) Read32(address uint32) uint32 {
	switch address {
	case 0x1f801810:
		return gpu.ReadGPUREAD()
	case 0x1f801814:
		return gpu.ReadStatus()
	}

	return 0
}

func (gpu *GPU) WriteGP0(data uint32) {
	op := data >> 24 // the most significant byte is opcode

	switch op {
	case 0x00:
		// NOP
	case 0xe1:
		gpu.GP0DrawModeSet(data)
	case 0xe2:
		gpu.GP0TextureWindowSetup(data)
	case 0xe3:
		gpu.GP0DrawingAreaTopLeftSet(data)
	case 0xe4:
		gpu.GP0DrawingAreaBottomRightSet(data)
	case 0xe5:
		gpu.GP0DrawingOffsetSet(data)
	case 0xe6:
		gpu.GP0MaskBitSetup(data)
	default:
		panic(fmt.Sprintf("[GPU::WriteGP0] Unknown command: %x\n", data))
	}
}

func (gpu *GPU) WriteGP1(data uint32) {
	op := data >> 24

	switch op {
	case 0x00:
		gpu.GP1Reset()
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
	default:
		panic(fmt.Sprintf("[GPU::WriteGP1] Unknown command: %x\n", data))
	}
}

func (gpu *GPU) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801810:
		gpu.WriteGP0(data)
	case 0x1f801814:
		gpu.WriteGP1(data)
	}
}