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

type GPU struct {
	Core *GoStationCore

	/* 1F801814h - GPUSTAT - GPU Status Register (R) -- whut the forking fock why itz long */
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

	rectTextureXFlip bool /* mirror texture rectangle along the x-axis */
	rectTextureYFlip bool /* mirror texture rectangle along the y-axis */
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
		0,
		false,
		0,
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

func (gpu *GPU) Read32(address uint32) uint32 {
	switch address {
	case 0x1f801810:
		panic("[GPU::Read32] GPUREAD is not implemented yet!")
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
	default:
		panic(fmt.Sprintf("[GPU::WriteGP0] Unknown Opcode: %x\n", data))
	}
}

func (gpu *GPU) Write32(address uint32, data uint32) {
	switch address {
	case 0x1f801810:
		gpu.WriteGP0(data)
	case 0x1f801814:
		panic("[GPU::Write32] GP1 is not implemented yet!")
	}
}
