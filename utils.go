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

	e.g. GetRange(0b10010001, 4, 4) = 0b1001)
*/
func GetRange(n uint32, pos int, len int) uint32 {
	return (n >> pos) & ((1 << len) - 1)
}

func PackRange(n *uint32, pos int, data uint32, size uint32) {
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
func ForceSignExtension16(n uint16, len int) int16 {
	shift := 16 - len
	return int16(n<<shift) >> shift
}

func MinOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

func MaxOf(vars ...int) int {
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}

	return max
}

func Clamp8(v int) int {
	if v > 255 {
		return 255
	}

	if v < 0 {
		return 0
	}

	return v
}

func Modulo(x, y int) int {
	return ((x % y) + y) % y
}

func IsTopLeft(v1, v2 *Vertex) bool {
	// is edge top (perfectly horizontal and points to right) or left (leans to left side)?
	return (v1.y == v2.y && v1.x < v2.x) || (v1.y > v2.y)
}

// (x, y) is a test point
func Edge(x1, y1, x2, y2, x, y int) int {
	// computes a cross product (area of paralleogram) between two vectors: <x-x1,y-y1,0> (edge origin to test point) and <x2-x1,y2-y1,0> (edge)
	// returns: 0 (on edge), negative (left of edge), positive (right of edge)
	return (x-x1)*(y2-y1) - (y-y1)*(x2-x1)
}
