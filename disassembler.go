package main

import (
	"fmt"
)

func (cpu *CPU) DisassemblePrimaryOpcode(opcode uint32) {
	op := GetValue(opcode, 26, 6)

	switch op {
	case 0x00:
		cpu.DisassembleSecondaryOpcode(opcode)
	case 0x01:
		cpu.DisOpBcondZ(opcode)
	case 0x02:
		cpu.DisOpJump(opcode)
	case 0x03:
		cpu.DisOpJAL(opcode)
	case 0x04:
		cpu.DisOpBEQ(opcode)
	case 0x05:
		cpu.DisOpBNE(opcode)
	case 0x06:
		cpu.DisOpBLEZ(opcode)
	case 0x07:
		cpu.DisOpBGTZ(opcode)
	case 0x08:
		cpu.DisOpADDI(opcode)
	case 0x09:
		cpu.DisOpADDIU(opcode)
	case 0x0a:
		cpu.DisOpSLTI(opcode)
	case 0x0b:
		cpu.DisOpSLTIU(opcode)
	case 0x0c:
		cpu.DisOpANDI(opcode)
	case 0x0d:
		cpu.DisOpORI(opcode)
	case 0x0f:
		cpu.DisOpLUI(opcode)
	case 0x20:
		cpu.DisOpLoadByte(opcode)
	case 0x21:
		cpu.DisOpLoadHWord(opcode)
	case 0x23:
		cpu.DisOpLoadWord(opcode)
	case 0x24:
		cpu.DisOpLoadByteU(opcode)
	case 0x25:
		cpu.DisOpLoadHWordU(opcode)
	case 0x28:
		cpu.DisOpStoreByte(opcode)
	case 0x29:
		cpu.DisOpStoreHWord(opcode)
	case 0x2b:
		cpu.DisOpStoreWord(opcode)
	case 0b010000:
		cpu.DisassembleCop0Opcode(opcode)
	default:
		panic(fmt.Sprintf("[CPU::DisassemblePrimaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) DisassembleSecondaryOpcode(opcode uint32) {
	op := GetValue(opcode, 0, 6)

	switch op {
	case 0x00:
		cpu.DisOpSLL(opcode)
	case 0x02:
		cpu.DisOpSRL(opcode)
	case 0x03:
		cpu.DisOpSRA(opcode)
	case 0x04:
		cpu.DisOpSLLV(opcode)
	case 0x06:
		cpu.DisOpSRLV(opcode)
	case 0x07:
		cpu.DisOpSRAV(opcode)
	case 0x08:
		cpu.DisOpJR(opcode)
	case 0x09:
		cpu.DisOpJALR(opcode)
	case 0x0c:
		cpu.DisOpSYS(opcode)
	case 0x0d:
		cpu.DisOpBRK(opcode)
	case 0x10:
		cpu.DisOpMFHI(opcode)
	case 0x11:
		cpu.DisOpMTHI(opcode)
	case 0x12:
		cpu.DisOpMFLO(opcode)
	case 0x13:
		cpu.DisOpMTLO(opcode)
	case 0x19:
		cpu.DisOpMULTU(opcode)
	case 0x1a:
		cpu.DisOpDIV(opcode)
	case 0x1b:
		cpu.DisOpDIVU(opcode)
	case 0x20:
		cpu.DisOpADD(opcode)
	case 0x21:
		cpu.DisOpADDU(opcode)
	case 0x23:
		cpu.DisOpSUBU(opcode)
	case 0x24:
		cpu.DisOpAND(opcode)
	case 0x25:
		cpu.DisOpOR(opcode)
	case 0x26:
		cpu.DisOpXOR(opcode)
	case 0x27:
		cpu.DisOpNOR(opcode)
	case 0x2a:
		cpu.DisOpSLT(opcode)
	case 0x2b:
		cpu.DisOpSLTU(opcode)
	default:
		panic(fmt.Sprintf("[CPU::DisassembleSecondaryOpcode] Unknown Opcode: %x", opcode))
	}
}

func (cpu *CPU) DisassembleCop0Opcode(opcode uint32) {
	op := GetValue(opcode, 21, 5)

	switch op {
	case 0b00000:
		cpu.DisOpMFC0(opcode)
	case 0b00100:
		cpu.DisOpMTC0(opcode)
	case 0b10000:
		cpu.DisOpRFE(opcode)
	default:
		panic(fmt.Sprintf("[CPU::DisassembleCop0Opcode] Unknown Opcode: %x", opcode))
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000001 | rs   | 00000| <--immediate16bit--> | bltz
// 000001 | rs   | 00001| <--immediate16bit--> | bgez
// 000001 | rs   | 10000| <--immediate16bit--> | bltzal
// 000001 | rs   | 10001| <--immediate16bit--> | bgezal
// bltz   rs,dest     if rs<0   then pc=$+4+(-8000h..+7FFFh)*4
// bgez   rs,dest     if rs>=0  then pc=$+4+(-8000h..+7FFFh)*4
// bltzal rs,dest     if rs<0   then pc=$+4+(..)*4, ra=$+8
// bgezal rs,dest     if rs>=0  then pc=$+4+(..)*4, ra=$+8
func (cpu *CPU) DisOpBcondZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	cond := GetValue(opcode, 16, 5)
	rs := int(GetValue(opcode, 21, 5))

	switch cond {
	case 0b00000:
		fmt.Printf("%-7s r%d,%08x", "bltz", rs, imm16)
	case 0b00001:
		fmt.Printf("%-7s r%d,%08x", "bgez", rs, imm16)
	case 0b10000:
		fmt.Printf("%-7s r%d,%08x", "bltzal", rs, imm16)
	case 0b10001:
		fmt.Printf("%-7s r%d,%08x", "bgezal", rs, imm16)
	default:
		panic(fmt.Sprintf("[CPU::DisOpBcondZ] Unknown condition: %x", cond))
	}
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00001x | <---------immediate26bit---------> | j/jal
// j      dest        pc=(pc and F0000000h)+(imm26bit*4)
func (cpu *CPU) DisOpJump(opcode uint32) {
	imm26 := GetValue(opcode, 0, 26)

	fmt.Printf("%-7s %08x", "j", imm26)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00001x | <---------immediate26bit---------> | j/jal
// jal    dest        pc=(pc and F0000000h)+(imm26bit*4),ra=$+8
func (cpu *CPU) DisOpJAL(opcode uint32) {
	imm26 := GetValue(opcode, 0, 26)

	fmt.Printf("%-7s %08x", "jal", imm26)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00010x | rs   | rt   | <--immediate16bit--> | beq/bne
// beq    rs,rt,dest  if rs=rt  then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) DisOpBEQ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "beq", rs, rt, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00010x | rs   | rt   | <--immediate16bit--> | beq/bne
// bne    rs,rt,dest  if rs<>rt then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) DisOpBNE(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "bne", rs, rt, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00011x | rs   | N/A  | <--immediate16bit--> | blez/bgtz
// blez   rs,dest     if rs<=0  then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) DisOpBLEZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x", "blez", rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 00011x | rs   | N/A  | <--immediate16bit--> | blez/bgtz
// bgtz   rs,dest     if rs>0   then pc=$+4+(-8000h..+7FFFh)*4
func (cpu *CPU) DisOpBGTZ(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x", "bgtz", rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addi  rt,rs,imm        rt=rs+(-8000h..+7FFFh) (with ov.trap)
func (cpu *CPU) DisOpADDI(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "addi", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// addiu rt,rs,imm        rt=rs+(-8000h..+7FFFh)
func (cpu *CPU) DisOpADDIU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "addiu", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// setlt slti  rt,rs,imm if rs<(-8000h..+7FFFh)  then rt=1 else rt=0 (signed)
func (cpu *CPU) DisOpSLTI(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "slti", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// setb  sltiu rt,rs,imm if rs<(FFFF8000h..7FFFh) then rt=1 else rt=0(unsigned)
func (cpu *CPU) DisOpSLTIU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "sltiu", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// andi rt,rs,imm        rt = rs AND (0000h..FFFFh)
func (cpu *CPU) DisOpANDI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "andi", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001xxx | rs   | rt   | <--immediate16bit--> | alu-imm
// ori  rt,rs,imm        rt = rs OR  (0000h..FFFFh)
func (cpu *CPU) DisOpORI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "ori", rt, rs, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 001111 | N/A  | rt   | <--immediate16bit--> | lui-imm
// lui  rt,imm            rt = (0000h..FFFFh) SHL 16
func (cpu *CPU) DisOpLUI(opcode uint32) {
	imm16 := GetValue(opcode, 0, 16)
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,%08x", "lui", rt, imm16)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lb  rt,imm(rs)    rt=[imm+rs]  ;byte sign-extended
func (cpu *CPU) DisOpLoadByte(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "lb", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lh  rt,imm(rs)    rt=[imm+rs]  ;halfword sign-extended
func (cpu *CPU) DisOpLoadHWord(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "lh", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lw  rt,imm(rs)    rt=[imm+rs]  ;word
func (cpu *CPU) DisOpLoadWord(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "lw", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lbu rt,imm(rs)    rt=[imm+rs]  ;byte zero-extended
func (cpu *CPU) DisOpLoadByteU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "lbu", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 100xxx | rs   | rt   | <--immediate16bit--> | load rt,[rs+imm]
// lhu rt,imm(rs)    rt=[imm+rs]  ;halfword zero-extended
func (cpu *CPU) DisOpLoadHWordU(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "lhu", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sb  rt,imm(rs)    [imm+rs]=(rt AND FFh)   ;store 8bit
func (cpu *CPU) DisOpStoreByte(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "sb", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sh  rt,imm(rs)    [imm+rs]=(rt AND FFFFh) ;store 16bit
func (cpu *CPU) DisOpStoreHWord(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "sh", rt, imm16, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 101xxx | rs   | rt   | <--immediate16bit--> | store rt,[rs+imm]
// sw  rt,imm(rs)    [imm+rs]=rt             ;store 32bit
func (cpu *CPU) DisOpStoreWord(opcode uint32) {
	imm16 := SignExtendedWord(GetValue(opcode, 0, 16))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,%08x(r%d)", "sw", rt, imm16, rs)
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
func (cpu *CPU) DisOpSLL(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "sll", rd, rt, imm5)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// srl  rd,rt,imm         rd = rt SHR (00h..1Fh)
func (cpu *CPU) DisOpSRL(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "srl", rd, rt, imm5)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | rt   | rd   | imm5 | 0000xx | shift-imm
// sra  rd,rt,imm         rd = rt SAR (00h..1Fh)
func (cpu *CPU) DisOpSRA(opcode uint32) {
	imm5 := GetValue(opcode, 6, 5)
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,r%d,%08x", "sra", rd, rt, imm5)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// sllv rd,rt,rs          rd = rt SHL (rs AND 1Fh)
func (cpu *CPU) DisOpSLLV(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "sllv", rd, rt, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// srlv rd,rt,rs          rd = rt SHR (rs AND 1Fh)
func (cpu *CPU) DisOpSRLV(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "srlv", rd, rt, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 0001xx | shift-reg
// srav rd,rt,rs          rd = rt SAR (rs AND 1Fh)
func (cpu *CPU) DisOpSRAV(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "srav", rd, rt, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 001000 | jr
// jr     rs          pc=rs
func (cpu *CPU) DisOpJR(opcode uint32) {
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d", "jr", rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | rd   | N/A  | 001001 | jalr
// jalr (rd,)rs(,rd)  pc=rs, rd=$+8
func (cpu *CPU) DisOpJALR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d", "jalr", rd, rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | <-----comment20bit------> | 00110x | sys/brk
// syscall  imm20        generates a system call exception
func (cpu *CPU) DisOpSYS(opcode uint32) {
	comment := GetValue(opcode, 6, 20)

	fmt.Printf("%-7s %x", "syscall", comment)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | <-----comment20bit------> | 00110x | sys/brk
// break    imm20        generates a breakpoint exception
func (cpu *CPU) DisOpBRK(opcode uint32) {
	comment := GetValue(opcode, 6, 20)

	fmt.Printf("%-7s %x", "break", comment)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mfhi   rd              rd=hi  ;move from hi
func (cpu *CPU) DisOpMFHI(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))

	fmt.Printf("%-7s r%d", "mfhi", rd)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 0100x1 | mthi/mtlo
// mthi   rs              hi=rs  ;move to hi
func (cpu *CPU) DisOpMTHI(opcode uint32) {
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d", "mthi", rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | N/A  | N/A  | rd   | N/A  | 0100x0 | mfhi/mflo
// mflo   rd              rd=lo  ;move from lo
func (cpu *CPU) DisOpMFLO(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))

	fmt.Printf("%-7s r%d", "mflo", rd)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | N/A  | N/A  | N/A  | 0100x1 | mthi/mtlo
// mtlo   rs              lo=rs  ;move to lo
func (cpu *CPU) DisOpMTLO(opcode uint32) {
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d", "mtlo", rs)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// multu  rs,rt           hi:lo = rs*rt (unsigned)
func (cpu *CPU) DisOpMULTU(opcode uint32) {
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d", "multu", rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// div    rs,rt           lo = rs/rt, hi=rs mod rt (signed)
// TODO timing
func (cpu *CPU) DisOpDIV(opcode uint32) {
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d", "div", rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | N/A  | N/A  | 0110xx | mul/div
// divu   rs,rt           lo = rs/rt, hi=rs mod rt (unsigned)
// TODO timing
func (cpu *CPU) DisOpDIVU(opcode uint32) {
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d", "divu", rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// add   rd,rs,rt         rd=rs+rt (with overflow trap)
func (cpu *CPU) DisOpADD(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "add", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// addu  rd,rs,rt         rd=rs+rt
func (cpu *CPU) DisOpADDU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "addu", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// subu  rd,rs,rt         rd=rs-rt
func (cpu *CPU) DisOpSUBU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "subu", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// and  rd,rs,rt         rd = rs AND rt
func (cpu *CPU) DisOpAND(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "and", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// or   rd,rs,rt         rd = rs OR  rt
func (cpu *CPU) DisOpOR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "or", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// xor  rd,rs,rt         rd = rs XOR rt
func (cpu *CPU) DisOpXOR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "xor", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// nor  rd,rs,rt         rd = FFFFFFFFh XOR (rs OR rt)
func (cpu *CPU) DisOpNOR(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "nor", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setlt slt   rd,rs,rt  if rs<rt then rd=1 else rd=0 (signed)
func (cpu *CPU) DisOpSLT(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "slt", rd, rs, rt)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 000000 | rs   | rt   | rd   | N/A  | 10xxxx | alu-reg
// setb  sltu  rd,rs,rt  if rs<rt then rd=1 else rd=0 (unsigned)
func (cpu *CPU) DisOpSLTU(opcode uint32) {
	rd := int(GetValue(opcode, 11, 5))
	rt := int(GetValue(opcode, 16, 5))
	rs := int(GetValue(opcode, 21, 5))

	fmt.Printf("%-7s r%d,r%d,r%d", "sltu", rd, rs, rt)
}

/*
CDisOp0 opcodes are implemented here
*/

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0000| rt   | rd   | N/A  | 000000 | MFCn rt,rd_dat  ;rt = dat
// mfc# rt,rd       ;rt = cDisOp#datRd ;data regs
func (cpu *CPU) DisOpMFC0(opcode uint32) {
	rd := GetValue(opcode, 11, 5)
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,r%d", "mfc0", rt, rd)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 0100nn |0|0100| rt   | rd   | N/A  | 000000 | MTCn rt,rd_dat  ;dat = rt
// mtc# rt,rd       ;cDisOp#datRd = rt ;data regs
func (cpu *CPU) DisOpMTC0(opcode uint32) {
	rd := GetValue(opcode, 11, 5)
	rt := int(GetValue(opcode, 16, 5))

	fmt.Printf("%-7s r%d,r%d", "mtc0", rt, rd)
}

// 31..26 |25..21|20..16|15..11|10..6 |  5..0  |
//
//	6bit  | 5bit | 5bit | 5bit | 5bit |  6bit  |
//
// 010000 |1|0000| N/A  | N/A  | N/A  | 010000 | COP0 10h  ;=RFE
// rfe
func (cpu *CPU) DisOpRFE(opcode uint32) {
	if GetValue(opcode, 0, 6) != 0b010000 {
		panic(fmt.Sprintf("[CPU::OpRFE] Unknown opcode: %x", opcode))
	}

	fmt.Printf("rfe")
}
