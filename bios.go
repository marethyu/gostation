package main

import (
	"fmt"
	"strings"
)

func IdentifySystemCall(r4 uint32) string {
	switch r4 {
	case 0x0:
		return "unknown (r4=0)"
	case 0x1:
		return "BIOS::EnterCriticalSection()"
	case 0x2:
		return "BIOS::ExitCriticalSection()"
	case 0x3:
		return "BIOS::ChangeThreadSubFunction(addr)"
	default:
		return "BIOS::DeliverEvent(F0000010h,4000h)"
	}
}

/*
Only TTY functions are supported

https://psx-spx.consoledev.net/kernelbios/#bios-tty-console-std_io
*/

func BIOSAFunction(gostation *GoStation, r9 uint32, log bool) {
	if log {
		fmt.Printf("[BIOSAFunction] BIOS A(%02Xh)\n", r9)
	}

	switch r9 {
	case 0x3c:
		BIOSPutchar(gostation)
	case 0x3e:
		BIOSPuts(gostation)
	case 0x3f:
		BIOSPrintf(gostation)
	}
}

func BIOSBFunction(gostation *GoStation, r9 uint32, log bool) {
	if log {
		fmt.Printf("[BIOSBFunction] BIOS B(%02Xh)\n", r9)
	}

	switch r9 {
	case 0x49:
		// TODO
	case 0x3d:
		BIOSPutchar(gostation)
	case 0x3f:
		BIOSPuts(gostation)
	}
}

/*
https://psx-spx.consoledev.net/kernelbios/#a3ch-or-b3dh-putcharchar-write-character-to-tty
*/
func BIOSPutchar(gostation *GoStation) {
	fmt.Print(string(uint8(BIOSFunctionArgument(gostation, 0))))
}

/*
https://psx-spx.consoledev.net/kernelbios/#a3eh-or-b3fh-putssrc-write-string-to-tty
*/
func BIOSPuts(gostation *GoStation) {
	var sb strings.Builder
	addr := BIOSFunctionArgument(gostation, 0)

	for {
		ch := gostation.CPU.Read8(addr)
		if ch == 0 {
			break
		}
		sb.WriteByte(ch)
		addr += 1
	}

	fmt.Println(sb.String())
}

/*
https://psx-spx.consoledev.net/kernelbios/#a3fh-printftxtparam1param2etc-print-string-to-console
*/
func BIOSPrintf(gostation *GoStation) {
	var sb strings.Builder
	addr := BIOSFunctionArgument(gostation, 0)

	for {
		ch := gostation.CPU.Read8(addr)
		if ch == 0 {
			break
		}
		sb.WriteByte(ch)
		addr += 1
	}

	txt := sb.String()
	nargs := strings.Count(txt, "%")
	var args []any

	for i := 1; i <= nargs; i += 1 {
		args = append(args, BIOSFunctionArgument(gostation, i))
	}

	fmt.Printf(sb.String(), args...)
}

/* Argument(s) are passed in R4,R5,R6,R7,[SP+10h],[SP+14h],etc. */
func BIOSFunctionArgument(gostation *GoStation, arg int) uint32 {
	if arg < 4 {
		return gostation.CPU.reg(arg + 4)
	}

	return gostation.CPU.Read32(gostation.CPU.reg(29) + uint32(0x10+arg*0x4))
}
