package main

import (
	"fmt"
	"math"
)

/*
Nice references:
https://psx-spx.consoledev.net/cpuspecifications/
https://ffhacktics.com/wiki/PSX_instruction_set
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00001x | <---------immediate26bit---------> | j/jal
// j      dest        pc=(pc and F0000000h)+(imm26bit*4)
func (cpu *CPU) OpJump(opcode uint32) {
	imm26 := GetValue(opcode, 0, 26)

	cpu.next_pc = (cpu.pc & 0xf0000000) | (imm26 << 2)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00010x | rs   | rt   | <--immediate16bit--> | beq/bne
// bne    rs,rt,dest  if rs<>rt then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) OpBNE(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := cpu.reg(rs) != cpu.reg(rt)
	if test {
		cpu.branch(imm16)
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addi  rt,rs,imm        rt=rs+(-8000h..+7FFFh) (with ov.trap)
func (cpu *CPU) OpADDI(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	a := int32(imm16)
	b := int32(cpu.reg(rs))

	if a > 0 && b > math.MaxInt32-a {
		panic("[CPU::OpADDI] Signed integer overflow!!!")
	}

	val := uint32(a + b)
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addiu rt,rs,imm        rt=rs+(-8000h..+7FFFh)
func (cpu *CPU) OpADDIU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) + imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// ori  rt,rs,imm        rt = rs OR  (0000h..FFFFh)
func (cpu *CPU) OpORI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5)) /* the low 16 bits of rt which is assumed to be zero will be filled with imm16 */

	val := cpu.reg(rs) | imm16
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001111 | N/A  | rt   | <--immediate16bit--> | lui-imm
// lui  rt,imm            rt = (0000h..FFFFh) SHL 16
func (cpu *CPU) OpLUI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16) /* this value will be placed in the high 16 bits of a 32 bit value */
	rt := int(GetValue(opcode, 16, 5))

	val := imm16 << 16 /* the low 16 bits of a 32 bit value is set to zero */
	cpu.modifyReg(rt, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lw  rt,imm(rs)    rt=[imm+rs]  ;word
func (cpu *CPU) OpLoadWord(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.Core.Bus.Read32(addr)

	// Loads have delay
	cpu.pending_load = true
	cpu.pending_r = rt
	cpu.pending_val = val
	cpu.load_countdown = 2
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sw  rt,imm(rs)    [imm+rs]=rt             ;store 32bit
func (cpu *CPU) OpStoreWord(opcode uint32) {
	if TestBit(cpu.sr, 16) {
		// Ignore write when cache is isolated
		return
	}

	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	addr := cpu.reg(rs) + imm16
	val := cpu.reg(rt)
	cpu.Core.Bus.Write32(addr, val)
}

/*
Secondary opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// sll  rd,rt,imm         rd = rt SHL (00h..1Fh)
func (cpu *CPU) OpSLL(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	val := cpu.reg(rt) << imm5
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// addu  rd,rs,rt         rd=rs+rt
func (cpu *CPU) OpADDU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) + cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// or   rd,rs,rt         rd = rs OR  rt
func (cpu *CPU) OpOR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	val := cpu.reg(rs) | cpu.reg(rt)
	cpu.modifyReg(rd, val)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setb  sltu  rd,rs,rt  if rs<rt then rd=1 else rd=0 (unsigned)
func (cpu *CPU) OpSLTU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	test := cpu.reg(rs) < cpu.reg(rt)
	if test {
		cpu.modifyReg(rd, 1)
	} else {
		cpu.modifyReg(rd, 0)
	}
}

/*
COP0 opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0100| rt   | rd   | N/A  | 000000 | MTCn rt,rd_dat  ;dat = rt
// mtc# rt,rd       ;cop#datRd = rt ;data regs
func (cpu *CPU) OpMTC0(opcode uint32) {
	rd := GetValue(opcode, 11, 5)
	rt := int(GetValue(opcode, 16, 5))

	val := cpu.reg(rt)
	switch rd {
	case 3, 5, 6, 7, 9, 11:
		if val != 0 {
			panic("[CPU::OpMTC0] The breakpoint registers are implemented yet!")
		}
	case 12:
		cpu.sr = val
	case 13:
		if val != 0 {
			panic("[CPU::OpMTC0] The cause register is not implemented yet!")
		}
	default:
		panic(fmt.Sprintf("[CPU::OpMTC0] Unsupported COP0 register: %x", rd))
	}
}
