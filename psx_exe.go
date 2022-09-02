package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"reflect"
)

/*
PSX executable header format:

	000h-007h ASCII ID "PS-X EXE"
	008h-00Fh Zerofilled
	010h      Initial PC                   (usually 80010000h, or higher)
	014h      Initial GP/R28               (usually 0)
	018h      Destination Address in RAM   (usually 80010000h, or higher)
	01Ch      Filesize (must be N*800h)    (excluding 800h-byte header)
	020h      Data section Start Address   (usually 0)
	024h      Data Section Size in bytes   (usually 0)
	028h      BSS section Start Address    (usually 0) (when below Size=None)
	02Ch      BSS section Size in bytes    (usually 0) (0=None)
	030h      Initial SP/R29 & FP/R30 Base (usually 801FFFF0h) (or 0=None)
	034h      Initial SP/R29 & FP/R30 Offs (usually 0, added to above Base)
	038h-04Bh Reserved for A(43h) Function (should be zerofilled in exefile)
	04Ch-xxxh ASCII marker
	           "Sony Computer Entertainment Inc. for Japan area"
	           "Sony Computer Entertainment Inc. for Europe area"
	           "Sony Computer Entertainment Inc. for North America area"
	           (or often zerofilled in some homebrew files)
	           (the BIOS doesn't verify this string, and boots fine without it)
	xxxh-7FFh Zerofilled
	800h...   Code/Data                  (loaded to entry[018h] and up)

See:
- https://web.archive.org/web/20210606105823/http://www.emulatronia.com/doctec/consolas/psx/exeheader.txt
- https://psx-spx.consoledev.net/cdromdrive/#filenameexe-general-purpose-executable
*/
type PSXExeHeader struct {
	Magic   [8]byte
	Text    uint32 /* SCE only */
	Data    uint32 /* SCE only */
	PC0     uint32
	GP0     uint32 /* SCE only */
	TAddr   uint32
	TSize   uint32
	DAddr   uint32 /* SCE only */
	DSize   uint32 /* SCE only */
	BAddr   uint32 /* SCE only */
	BSize   uint32 /* SCE only */
	SAddr   uint32
	SSize   uint32
	SavedSP uint32
	SavedFP uint32
	SavedGP uint32
	SavedRA uint32
	SavedS0 uint32
}

type PSXExecutable struct {
	Header PSXExeHeader
	Data   []byte
}

func NewPSXExe(pathToExe string) *PSXExecutable {
	exe, err := os.Open(pathToExe)
	if err != nil {
		log.Fatal("Unable to open PSX executable: ", err)
	}
	defer exe.Close()

	header := PSXExeHeader{}

	binary.Read(exe, binary.LittleEndian, &header)

	magic := [8]byte{0x50, 0x53, 0x2d, 0x58, 0x20, 0x45, 0x58, 0x45} /* PS-X EXE */

	if !reflect.DeepEqual(header.Magic, magic) {
		panic("[NewPSXExe] The PSX executable does not begin with 'PSX-X EXE'")
	}

	stats, statsErr := exe.Stat()
	if statsErr != nil {
		log.Fatal("[NewPSXExe] Stat on PSX executable failed: ", statsErr)
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	tsize := uint32(size - 0x800)
	if tsize < header.TSize {
		fmt.Println("[NewPSXExe] WARNING: header.TSize does not agree with actual size?")
		header.TSize = tsize
	}

	bufr := bufio.NewReader(exe)
	_, err = bufr.Read(bytes)

	bytes = bytes[0x800-76:] // remove the header (we already parsed the header data so we need to remove zeroes)

	return &PSXExecutable{
		header,
		bytes,
	}
}
