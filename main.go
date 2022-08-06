package main

func main() {
	gopsx := NewCore("SCPH1001.BIN")

	for {
		gopsx.Step()
	}
}
