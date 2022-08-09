package main

import (
	"fmt"
)

type CPU struct {
	Core *GoStationCore

	/* All registers */
	r  [32]uint32 /* R0 to R31 */
	pc uint32
	hi uint32
	lo uint32

	/* necessary for branch delay slot */
	next_pc     uint32
	isBranch    bool /* true if branch or jump occured */
	isDelaySlot bool /* (if branch is success) true if instruction is branch delay slot */

	/* necessary for load delay slot */
	pending_load   bool
	pending_r      int
	pending_val    uint32
	load_countdown int

	/* cop0r12 - SR - System status register (R/W) */
	sr uint32
	/* cop0r13 - CAUSE - (R)  Describes the most recently recognised exception */
	cause uint32
	/* cop0r14 - EPC - Return Address from Trap */
	epc uint32

	current_pc uint32
}

func NewCPU(core *GoStationCore) *CPU {
	return &CPU{
		core,
		[32]uint32{0},
		0xbfc00000,
		0,
		0,
		0xbfc00004,
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
	}
}

func (cpu *CPU) Step() {
	cpu.current_pc = cpu.pc
	if cpu.current_pc%4 != 0 {
		cpu.enterException(EXC_ADDR_ERROR_LOAD)
	}
	opcode := cpu.Core.Bus.Read32(cpu.pc)

	cpu.pc = cpu.next_pc
	cpu.next_pc += 4

	cpu.isDelaySlot = cpu.isBranch
	cpu.isBranch = false

	cpu.ExecutePrimaryOpcode(opcode)

	if cpu.pending_load {
		cpu.load_countdown -= 1
		if cpu.load_countdown == 0 {
			cpu.modifyReg(cpu.pending_r, cpu.pending_val)
			cpu.pending_load = false
		}
	}
}

func (cpu *CPU) Log(logRegisters bool) {
	if logRegisters {
		fmt.Println()
		for i := 0; i < 8; i++ {
			r := i * 4
			fmt.Printf("r%d=%08x r%d=%08x r%d=%08x r%d=%08x\n", r, cpu.r[r], r+1, cpu.r[r+1], r+2, cpu.r[r+2], r+3, cpu.r[r+3])
		}
		fmt.Printf("hi=%08x lo=%08x\n", cpu.hi, cpu.lo)
	}

	fmt.Printf("[%08x]    ", cpu.pc)
	cpu.DisassemblePrimaryOpcode(cpu.Core.Bus.Read32(cpu.pc))
	fmt.Println()
}

/* Refer to this page https://psx-spx.consoledev.net/cpuspecifications/#cpu-opcode-encoding for all opcodes and its encodings */
func (cpu *CPU) ExecutePrimaryOpcode(opcode uint32) {
	op := GetValue(opcode, 26, 6)

	switch op {
	case 0x00:
		cpu.ExecuteSecondaryOpcode(opcode)
	case 0x01:
		cpu.OpBcondZ(opcode)
	case 0x02:
		cpu.OpJump(opcode)
	case 0x03:
		cpu.OpJAL(opcode)
	case 0x04:
		cpu.OpBEQ(opcode)
	case 0x05:
		cpu.OpBNE(opcode)
	case 0x06:
		cpu.OpBLEZ(opcode)
	case 0x07:
		cpu.OpBGTZ(opcode)
	case 0x08:
		cpu.OpADDI(opcode)
	case 0x09:
		cpu.OpADDIU(opcode)
	case 0x0a:
		cpu.OpSLTI(opcode)
	case 0x0b:
		cpu.OpSLTIU(opcode)
	case 0x0c:
		cpu.OpANDI(opcode)
	case 0x0d:
		cpu.OpORI(opcode)
	case 0x0f:
		cpu.OpLUI(opcode)
	case 0x20:
		cpu.OpLoadByte(opcode)
	case 0x21:
		cpu.OpLoadHWord(opcode)
	case 0x23:
		cpu.OpLoadWord(opcode)
	case 0x24:
		cpu.OpLoadByteU(opcode)
	case 0x25:
		cpu.OpLoadHWordU(opcode)
	case 0x28:
		cpu.OpStoreByte(opcode)
	case 0x29:
		cpu.OpStoreHWord(opcode)
	case 0x2b:
		cpu.OpStoreWord(opcode)
	case 0b010000:
		cpu.ExecuteCop0Opcode(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecutePrimaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) ExecuteSecondaryOpcode(opcode uint32) {
	op := GetValue(opcode, 0, 6)

	switch op {
	case 0x00:
		cpu.OpSLL(opcode)
	case 0x02:
		cpu.OpSRL(opcode)
	case 0x03:
		cpu.OpSRA(opcode)
	case 0x04:
		cpu.OpSLLV(opcode)
	case 0x06:
		cpu.OpSRLV(opcode)
	case 0x07:
		cpu.OpSRAV(opcode)
	case 0x08:
		cpu.OpJR(opcode)
	case 0x09:
		cpu.OpJALR(opcode)
	case 0x0c:
		cpu.OpSYS(opcode)
	case 0x10:
		cpu.OpMFHI(opcode)
	case 0x11:
		cpu.OpMTHI(opcode)
	case 0x12:
		cpu.OpMFLO(opcode)
	case 0x13:
		cpu.OpMTLO(opcode)
	case 0x19:
		cpu.OpMULTU(opcode)
	case 0x1a:
		cpu.OpDIV(opcode)
	case 0x1b:
		cpu.OpDIVU(opcode)
	case 0x20:
		cpu.OpADD(opcode)
	case 0x21:
		cpu.OpADDU(opcode)
	case 0x23:
		cpu.OpSUBU(opcode)
	case 0x24:
		cpu.OpAND(opcode)
	case 0x25:
		cpu.OpOR(opcode)
	case 0x26:
		cpu.OpXOR(opcode)
	case 0x27:
		cpu.OpNOR(opcode)
	case 0x2a:
		cpu.OpSLT(opcode)
	case 0x2b:
		cpu.OpSLTU(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecuteSecondaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) ExecuteCop0Opcode(opcode uint32) {
	op := GetValue(opcode, 21, 5)

	switch op {
	case 0b00000:
		cpu.OpMFC0(opcode)
	case 0b00100:
		cpu.OpMTC0(opcode)
	case 0b10000:
		cpu.OpRFE(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecuteCop0Opcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) modifyReg(i int, v uint32) {
	cpu.r[i] = v
	cpu.r[0] = 0 // R0 is always zero

	if cpu.pending_load && cpu.pending_r == i {
		// Uh well... never mind the load
		cpu.pending_load = false
	}
}

func (cpu *CPU) reg(i int) uint32 {
	return cpu.r[i]
}

func (cpu *CPU) loadDelaySlotInit(i int, v uint32) {
	if cpu.pending_load {
		cpu.modifyReg(cpu.pending_r, cpu.pending_val)
	}

	cpu.pending_load = true
	cpu.pending_r = i
	cpu.pending_val = v
	cpu.load_countdown = 2
}

func (cpu *CPU) branch(imm16 uint32) {
	cpu.next_pc = cpu.pc + (imm16 << 2)
	cpu.isBranch = true
}
