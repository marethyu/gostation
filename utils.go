package main

func TestBit(n uint32, pos int) bool {
	return n&(1<<pos) != 0
}

func ModifyBit(n *uint32, pos int, test bool) {
	if test {
		*n |= (1 << pos)
	} else {
		*n &= ^(1 << pos)
	}
}

/*
Get value from n starting from bit position pos with length len

	e.g. GetValue(0b10010001, 4, 4) = 0b1001)
*/
func GetValue(n uint32, pos int, len int) uint32 {
	return (n >> pos) & ((1 << len) - 1)
}

func PackValue(n *uint32, pos int, data uint32, size uint32) {
	*n |= (data & ((1 << size) - 1)) << pos
}

func SignExtendedByte(n uint8) uint32 {
	return uint32(int8(n))
}

func SignExtendedHWord(n uint16) uint32 {
	return uint32(int16(n))
}

func SignExtendedWord(n uint32) uint32 {
	return uint32(int16(n))
}

/*
force sign extension on values whose bit size less than 16
in order to get successful small value sign extension, the value must be shifted 16-bitlen bits left (making them 16 bit unsigned)
of course, shift them back 16-bitlen bits right to get bitlen bit signed which is what we want
*/
func ForceSignExtension16(n uint16, len int) uint16 {
	shift := 16 - len
	return uint16(int16(n<<shift) >> shift)
}

func MinOf(vars ...int16) int16 {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

func MaxOf(vars ...int16) int16 {
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}

	return max
}
