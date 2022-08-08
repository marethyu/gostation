package main

import (
	"fmt"
	"os"
)

/* https://psx-spx.consoledev.net/cpuspecifications/#cop0-exception-handling

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
	EXC_OVERFLOW         = 0xc
)

/* See also CPU::OpRFE */
func (cpu *CPU) enterException(cause uint32) {
	switch cause {
	case EXC_ADDR_ERROR_LOAD:
		fmt.Fprintln(os.Stderr, "[CPU::enterException] Misaligned address while loading")
	case EXC_ADDR_ERROR_STORE:
		fmt.Fprintln(os.Stderr, "[CPU::enterException] Misaligned address while storing")
	case EXC_SYSCALL:
		fmt.Fprintln(os.Stderr, "[CPU::enterException] Syscall")
	case EXC_OVERFLOW:
		fmt.Fprintln(os.Stderr, "[CPU::enterException] Overflow during addition")
	}

	var vector uint32
	if TestBit(cpu.sr, 22) {
		// 1=ROM/KSEG1
		vector = 0xbfc00180
	} else {
		// 0=RAM/KSEG0
		vector = 0x80000080
	}

	// shift mode bits in sr 2 positions left (bits 6-7 are always zero)
	var mask uint32 = 0b111111
	mode := cpu.sr & mask
	cpu.sr &= ^mask
	cpu.sr |= (mode << 2) & mask

	cpu.cause = cause << 2 // bits 0-1 are unused so

	if cpu.isDelaySlot {
		cpu.epc = cpu.current_pc - 4
		ModifyBit(&cpu.cause, 31, true)
	} else {
		cpu.epc = cpu.current_pc
	}

	cpu.pc = vector
	cpu.next_pc = vector + 4
}
