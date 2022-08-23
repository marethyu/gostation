package main

import (
	"fmt"
)

/*
https://psx-spx.consoledev.net/cpuspecifications/#cop0-exception-handling

	00h INT     Interrupt
	01h MOD     Tlb modification (none such in PSX)
	02h TLBL    Tlb load         (none such in PSX)
	03h TLBS    Tlb store        (none such in PSX)
	04h AdEL    Address error, Data load or Instruction fetch
	05h AdES    Address error, Data store
	            The address errors occur when attempting to read
	            outside of KUseg in user mode and when the address
	            is misaligned. (See also: BadVaddr register)
	06h IBE     Bus error on Instruction fetch
	07h DBE     Bus error on Data load/store
	08h Syscall Generated unconditionally by syscall instruction
	09h BP      Breakpoint - break instruction
	0Ah RI      Reserved instruction
	0Bh CpU     Coprocessor unusable
	0Ch Ov      Arithmetic overflow
	0Dh-1Fh     Not used
*/
const (
	EXC_ADDR_ERROR_LOAD  = 0x4
	EXC_ADDR_ERROR_STORE = 0x5
	EXC_SYSCALL          = 0x8
	EXC_BREAK            = 0x9
	EXC_RESERVED_INS     = 0xa
	EXC_COP_UNUSABLE     = 0xb
	EXC_OVERFLOW         = 0xc
)

type Coprocessor0 struct {
	cpu *CPU

	/* https://psx-spx.consoledev.net/cpuspecifications/#cop0-register-summary */
	r3  uint32 /* BPC - Breakpoint on execute (R/W) */
	r5  uint32 /* BDA - Breakpoint on data access (R/W) */
	r6  uint32 /* JUMPDEST - Randomly memorized jump address (R) */
	r7  uint32 /* DCIC - Breakpoint control (R/W) */
	r8  uint32 /* BadVaddr - Bad Virtual Address (R) */
	r9  uint32 /* BDAM - Data Access breakpoint mask (R/W) */
	r11 uint32 /* BPCM - Execute breakpoint mask (R/W) */

	/* for exception handling */
	sr    uint32 /* cop0r12 - SR - System status register (R/W) */
	cause uint32 /* cop0r13 - CAUSE - (R)  Describes the most recently recognised exception */
	epc   uint32 /* cop0r14 - EPC - Return Address from Trap (R) */
}

func (cop0 *Coprocessor0) CacheIsolated() bool {
	return TestBit(cop0.sr, 16)
}

func NewCoprocessor0(cpu *CPU) *Coprocessor0 {
	return &Coprocessor0{
		cpu,
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

func (cop0 *Coprocessor0) GetRegister(reg uint32) uint32 {
	switch reg {
	case 3:
		return cop0.r3
	case 5:
		return cop0.r5
	case 6:
		return 0 // TODO
	case 7:
		return 0 // TODO
	case 8:
		return cop0.r8
	case 9:
		return 0 // TODO
	case 11:
		return cop0.r11
	case 12:
		return cop0.sr
	case 13:
		return cop0.cause // TODO
	case 14:
		return cop0.epc
	case 15:
		return 0x00000002
	default:
		panic(fmt.Sprintf("[Coprocessor0:::GetRegister] tried to read unsupported COP0 register: %x", reg))
	}
}

func (cop0 *Coprocessor0) ModifyRegister(reg uint32, v uint32) {
	switch reg {
	case 3:
		cop0.r3 = v
	case 5:
		cop0.r5 = v
	case 6:
		cop0.r6 = v
	case 7:
		cop0.r7 = v
	case 9:
		cop0.r9 = v
	case 11:
		cop0.r11 = v
	case 12:
		cop0.sr = v
	case 13:
		cop0.cause = v
	default:
		panic(fmt.Sprintf("[Coprocessor0:::ModifyRegister] tried to write to unsupported COP0 register: %x", reg))
	}
}

func (cop0 *Coprocessor0) EnterException(cause uint32, msg string) {
	fmt.Printf("[Coprocessor0::EnterException] %s\n", msg)

	var vector uint32
	if TestBit(cop0.sr, 22) {
		// 1=ROM/KSEG1
		vector = 0xbfc00180
	} else {
		// 0=RAM/KSEG0
		vector = 0x80000080
	}

	// shift mode bits in sr 2 positions left (bits 6-7 are always zero)
	var mask uint32 = 0b111111
	mode := cop0.sr & mask
	cop0.sr &= ^mask
	cop0.sr |= (mode << 2) & mask

	cop0.cause = cause << 2 // bits 0-1 are unused so

	if cop0.cpu.isDelaySlot {
		cop0.epc = cop0.cpu.current_pc - 4
		ModifyBit(&cop0.cause, 31, true)
	} else {
		cop0.epc = cop0.cpu.current_pc
	}

	cop0.cpu.pc = vector
	cop0.cpu.next_pc = vector + 4
}

func (cop0 *Coprocessor0) LeaveException() {
	var mask uint32 = 0b111111
	mode := cop0.sr & mask
	cop0.sr &= ^mask
	cop0.sr |= (mode & 0b110000) | (mode >> 2) // bits 4-5 are unchanged!!
}
