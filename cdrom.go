package main

import (
	"fmt"
)

const (
	CDROM_OFFSET = 0x1f801800
	CDROM_SIZE   = 4
)

const (
	RESP_INT0 = iota /* INT0   No response received (no interrupt request) */
	RESP_INT1        /* INT1   Received SECOND (or further) response to ReadS/ReadN (and Play+Report) */
	RESP_INT2        /* INT2   Received SECOND response (to various commands) */
	RESP_INT3        /* INT3   Received FIRST response (to any command) */
	RESP_INT4        /* INT4   DataEnd (when Play/Forward reaches end of disk) (maybe also for Read?) */
	RESP_INT5        /* INT5   Received error-code (in FIRST or SECOND response)
	                           INT5 also occurs on SECOND GetID response, on unlicensed disks
		                       INT5 also occurs when opening the drive door (even if no command
			                   was sent, ie. even if no read-command or other command is active) */
	RESP_INT6 /*        INT6   N/A */
	RESP_INT7 /*        INT7   N/A */
)

type CDROM struct {
	Core *GoStation

	/* 1F801800h - Index/Status Register

		0-1 Index   Port 1F801801h-1F801803h index (0..3 = Index0..Index3)   (R/W)
	    2   ADPBUSY XA-ADPCM fifo empty  (0=Empty) ;set when playing XA-ADPCM sound
	    3   PRMEMPT Parameter fifo empty (1=Empty) ;triggered before writing 1st byte
	    4   PRMWRDY Parameter fifo full  (0=Full)  ;triggered after writing 16 bytes
	    5   RSLRRDY Response fifo empty  (0=Empty) ;triggered after reading LAST byte
	    6   DRQSTS  Data fifo empty      (0=Empty) ;triggered after reading LAST byte
	    7   BUSYSTS Command/parameter transmission busy  (1=Busy)
	*/
	index int  /* 0-1 */
	busy  bool /* 7 */

	paramFIFO *FIFO[uint8]
	respFIFO  *FIFO[uint8]

	/* 1F801803h.Index0 - Request Register

		0-4 0    Not used (should be zero)
	    5   SMEN Want Command Start Interrupt on Next Command (0=No change, 1=Yes)
	    6   BFWR ...
	    7   BFRD Want Data         (0=No/Reset Data Fifo, 1=Yes/Load Data Fifo)
	*/
	irqRequest uint32

	/* 1F801802h.Index1, 1F801803h.Index0, 1F801803h.Index2 - Interrupt Enable Register

	   0-4  Interrupt Enable Bits (usually all set, ie. 1Fh=Enable All IRQs)
	   5-7  Unknown/unused (write: should be zero) (read: usually all bits set)
	*/
	irqEnable uint32

	/* 1F801803h.Index1, 1F801803h.Index3 - Interrupt Flag Register

		0-2   Read: Response Received   Write: 7=Acknowledge   ;INT1..INT7
	    3     Read: Unknown (usually 0) Write: 1=Acknowledge   ;INT8  ;XXX CLRBFEMPT
	    4     Read: Command Start       Write: 1=Acknowledge   ;INT10h;XXX CLRBFWRDY
	    5     Read: Always 1 ;XXX "_"   Write: 1=Unknown              ;XXX SMADPCLR
	    6     Read: Always 1 ;XXX "_"   Write: 1=Reset Parameter Fifo ;XXX CLRPRM
	    7     Read: Always 1 ;XXX "_"   Write: 1=Unknown              ;XXX CHPRST
	*/
	irqFlag uint32

	/* temporary hack to resolve irq timing problem...
	   sometimes cdrom irq accidentally gets set after it gets acknowledged by Interrupts::Write WHILE cpu is busy processing cdrom interrupt */
	irqAcknowledged bool
}

func NewCDROM(core *GoStation) *CDROM {
	return &CDROM{
		core,
		0,
		false,
		NewFIFO[uint8](),
		NewFIFO[uint8](),
		0,
		0,
		0,
		false,
	}
}

func (cdrom *CDROM) Contains(address uint32) bool {
	return address >= CDROM_OFFSET && address < (CDROM_OFFSET+CDROM_SIZE)
}

func (cdrom *CDROM) Step() {
	cdrom.busy = false

	if (cdrom.irqFlag&cdrom.irqEnable&0b111) > 0 && !cdrom.irqAcknowledged {
		cdrom.Core.Interrupts.Request(IRQ_CDROM)
		cdrom.irqAcknowledged = true
	}
}

func (cdrom *CDROM) Read8(address uint32) uint8 {
	fmt.Printf("[CDROM::Read8] addr=%x (index=%d)\n", address, cdrom.index)

	switch address {
	case 0x1f801800: // Index/Status
		var status uint32 = 0

		status |= uint32(cdrom.index)
		ModifyBit(&status, 2, false) // hardcoded for now
		ModifyBit(&status, 3, cdrom.paramFIFO.Empty())
		ModifyBit(&status, 4, !cdrom.paramFIFO.Done())
		ModifyBit(&status, 5, !cdrom.respFIFO.Empty())
		ModifyBit(&status, 6, false) // hardcoded for now
		ModifyBit(&status, 7, cdrom.busy)

		return uint8(status)
	case 0x1f801801: // Response FIFO
		if !cdrom.respFIFO.Empty() {
			return cdrom.respFIFO.Pop()
		}
		return 0
	case 0x1f801802: // Data FIFO
		return 0 // TODO
	case 0x1f801803:
		if cdrom.index%2 == 0 {
			// Interrupt Enable Register
			return uint8(cdrom.irqEnable)
		} else {
			// Interrupt Flag Register
			var flag uint32 = 0b11100000
			PackRange(&flag, 0, uint32(cdrom.irqFlag), 3)
			return uint8(flag)
		}
	default:
		panic(fmt.Sprintf("[CDROM::Read8] Invalid address: %x", address))
	}
}

func (cdrom *CDROM) ProcessCommand(cmd uint8) {
	fmt.Printf("[CDROM::ProcessCommand] %x\n", cmd)

	cdrom.irqAcknowledged = false // reset

	switch cmd {
	case 0x19:
		cdrom.CommandTest()
	default:
		fmt.Printf("[CDROM::ProcessCommand] WARNING: Unknown command: %x\n", cmd)
	}

	cdrom.busy = true
}

func (cdrom *CDROM) Write8(address uint32, data uint8) {
	fmt.Printf("[CDROM::Write8] addr=%x (index=%d)\n", address, cdrom.index)

	switch address {
	case 0x1f801800: // Index/Status
		cdrom.index = int(data & 0b11)
	case 0x1f801801:
		switch cdrom.index {
		case 0: // Command Register
			cdrom.ProcessCommand(data)
		case 1: // Sound Map Data Out
		case 2: // Sound Map Coding Info
		case 3: // Right-CD to Right-SPU Volume
		}
	case 0x1f801802:
		switch cdrom.index {
		case 0: // Parameter FIFO
			cdrom.paramFIFO.Push(data)
		case 1: // Interrupt Enable Register
			cdrom.irqEnable = uint32(data) // TODO bits 5-7
		case 2: // Left-CD to Left-SPU Volume
		case 3: // Right-CD to Left-SPU Volume
		}
	case 0x1f801803:
		switch cdrom.index {
		case 0: // Interrupt Request Register
			cdrom.irqRequest = uint32(data)
		case 1: // Interrupt Flag Register
			if TestBit(uint32(data), 6) {
				// reset parameter fifo
				cdrom.paramFIFO.Reset(16)
			}
			cdrom.irqFlag &= ^uint32(data & 0b11111) // writing 1 will reset irq flags (it is nonsense to use values other than 07h or 1Fh?)
		case 2: // Left-CD to Right-SPU Volume
		case 3: // Audio Volume Apply Changes
		}
	}
}

/*
https://psx-spx.consoledev.net/cdromdrive/#cdrom-test-commands-version-switches-region-chipset-scex
*/
func (cdrom *CDROM) CommandTest() {
	if !cdrom.paramFIFO.Empty() {
		arg := cdrom.paramFIFO.Pop()

		switch arg {
		case 0x20: // 19h,20h --> INT3(yy,mm,dd,ver)
			// 94h,09h,19h,C0h  ;PSX (PU-7)               19 Sep 1994, version vC0 (a)
			cdrom.respFIFO.Reset(4)
			cdrom.respFIFO.Push(0x94)
			cdrom.respFIFO.Push(0x09)
			cdrom.respFIFO.Push(0x19)
			cdrom.respFIFO.Push(0xc0)

			cdrom.irqFlag = RESP_INT3
		default:
			panic(fmt.Sprintf("[CDROM::CommandTest] Unknown argument: %x", arg))
		}
	} else {
		panic("[CDROM::CommandTest] Missing one argument")
	}
}
