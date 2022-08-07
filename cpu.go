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
	next_pc uint32

	/* COP0 system status */
	sr uint32
}

func NewCPU(core *GoStationCore) *CPU {
	return &CPU{
		core,
		[32]uint32{0},
		0xbfc00000,
		0,
		0,
		0xbfc00004,
		0,
	}
}

func (cpu *CPU) Step() {
	opcode := cpu.Core.Bus.Read32(cpu.pc)

	cpu.pc = cpu.next_pc
	cpu.next_pc += 4

	cpu.ExecutePrimaryOpcode(opcode)
}

/* Refer to this page https://psx-spx.consoledev.net/cpuspecifications/#cpu-opcode-encoding for all opcodes and its encodings */
func (cpu *CPU) ExecutePrimaryOpcode(opcode uint32) {
	op := GetValue(opcode, 26, 6)

	switch op {
	case 0x00:
		cpu.ExecuteSecondaryOpcode(opcode)
	case 0x02:
		cpu.OpJump(opcode)
	case 0x05:
		cpu.OpBNE(opcode)
	case 0x08:
		cpu.OpADDI(opcode)
	case 0x09:
		cpu.OpADDIU(opcode)
	case 0x0d:
		cpu.OpORI(opcode)
	case 0x0f:
		cpu.OpLUI(opcode)
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
	case 0x25:
		cpu.OpOR(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecuteSecondaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) ExecuteCop0Opcode(opcode uint32) {
	op := GetValue(opcode, 21, 5)

	switch op {
	case 0b00100:
		cpu.OpMTC0(opcode)
	default:
		panic(fmt.Sprintf("[CPU::ExecuteCop0Opcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) modifyReg(i int, v uint32) {
	cpu.r[i] = v
	cpu.r[0] = 0 // R0 is always zero
}

func (cpu *CPU) reg(i int) uint32 {
	return cpu.r[i]
}

func (cpu *CPU) branch(imm16 uint32) {
	cpu.next_pc = cpu.pc + (imm16 << 2)
}
