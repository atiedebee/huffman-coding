package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type letters struct {
	freq int32
	char byte
}

type node struct {
	next [2](*node)

	isLeaf [2]bool
	sum    int32
	char   [2]byte
}

var letterInfo [256]letters

type Mode uint8

const (
	CompressMode Mode = iota
	DecompressMode
)

func countLetters(file io.ByteReader) {
	for {
		ch, err := file.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			// TODO: Handle error.
		}

		letterInfo[ch].freq += 1
		letterInfo[ch].char = ch
	}
}

func sortLetters() int {
	var i int
	for i = 0; i < len(letterInfo); i++ {
		for j := i + 1; j < len(letterInfo); j++ {
			if letterInfo[i].freq < letterInfo[j].freq {
				char := letterInfo[i].char
				freq := letterInfo[i].freq
				letterInfo[i].freq = letterInfo[j].freq
				letterInfo[i].char = letterInfo[j].char
				letterInfo[j].freq = freq
				letterInfo[j].char = char
			}
		}
		if letterInfo[i].freq == 0 {
			break
		}
	}

	return i - 1
}

func initCodes(head *node, codes *[256][24]int8, temp_codes *[24]int8, depth int) {
	for p := int8(0); p < 2; p++ {
		if head.isLeaf[p] {
			temp_codes[depth] = p
			temp_codes[depth+1] = -1
			i := 0
			for i = 0; temp_codes[i] != -1; i++ {
				codes[uint8(head.char[p])][i] = temp_codes[i]
			}
			codes[uint8(head.char[p])][i] = -1

		} else {
			temp_codes[depth] = p
			initCodes(head.next[p], codes, temp_codes, depth+1)
		}
	}
}

func writeBit(b, i byte) byte {
	return (b << 7) >> i
}

func readBit(c, i byte) byte {
	return (c & ((1 << 7) >> i)) >> (7 - i)
}

func checkWrite(f_out io.ByteWriter, c, c_index *byte) {
	if *c_index >= 8 {
		f_out.WriteByte(*c)
		*c_index = 0
		*c = 0
	}
}

func writeTree(head *node, f_out io.ByteWriter, c, c_index *byte) {
	for p := 0; p < 2; p++ {
		checkWrite(f_out, c, c_index)

		if head.isLeaf[p] == true {
			*c |= writeBit(1, *c_index)
			*c_index++
			for i := byte(0); i < 8; i++ {
				checkWrite(f_out, c, c_index)
				*c |= writeBit(readBit(head.char[p], i), *c_index)
				*c_index++
			}
			checkWrite(f_out, c, c_index)
		} else {
			*c_index++ //leaves the bit we were at at 0
			writeTree(head.next[p], f_out, c, c_index)
		}
	}
}

func compress(head *node, f_in io.ByteReader, f_out io.ByteWriter, amount int) {
	var codes [256][24]int8
	var temp_codes [24]int8
	var c, c_index byte

	writeTree(head, f_out, &c, &c_index)
	initCodes(head, &codes, &temp_codes, 0)

	for {
		ch_in, err := f_in.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// TODO: Handle errors.
		}
		tmp := uint8(ch_in)
		for i := 0; codes[tmp][i] != -1; i++ {
			checkWrite(f_out, &c, &c_index)

			c |= writeBit(byte(codes[tmp][i]), c_index)
			c_index++
		}
		checkWrite(f_out, &c, &c_index)
	}

	if c_index > 0 {
		f_out.WriteByte(c)
	}
}

func decompress(head *node, f_in io.ByteReader, f_out io.ByteWriter, c, c_index *byte) {
	parse_node := head

	for *c_index >= 8 {
		ch, err := f_in.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			// TODO: Handle errors.
		}
		*c = ch
		*c_index = 0

		next_step := readBit(*c, *c_index)

		if parse_node.isLeaf[next_step] == true {
			f_out.WriteByte(parse_node.char[next_step])
			parse_node = head

		} else {
			parse_node = parse_node.next[next_step]
		}

		*c_index++
	}
}

func readTree(f_in io.ByteReader, c, c_index *byte) *node {
	var node node

	for p := 0; p < 2; p++ {
		if *c_index >= 8 {
			ch, err := f_in.ReadByte()
			if err != nil {
				log.Fatal(err)
			}

			*c = ch
			*c_index = 0
		}

		if readBit(*c, *c_index) != 1 {
			*c_index++
			node.isLeaf[p] = false
			node.next[p] = readTree(f_in, c, c_index)
			continue
		}

		node.isLeaf[p] = true
		node.next[p] = nil
		*c_index++
		for i := byte(0); i < 8; i++ {
			if *c_index >= 8 {
				ch, err := f_in.ReadByte()
				if err != nil {
					log.Fatal(err)
				}
				*c = ch
				*c_index = 0
			}

			node.char[p] |= writeBit(readBit(*c, *c_index), i) //copy bits over
			*c_index++
		}
	}

	return &node
}

func padd(depth int) {
	const padding string = "    "
	for i := 0; i < depth; i++ {
		fmt.Printf(padding)
	}
}

func printTree(head *node, depth int, isabove int) {
	if head.isLeaf[0] == false {
		printTree(head.next[0], depth+1, 1)
	} else {
		padd(depth)
		fmt.Printf("/--%q\n", head.char[0])
	}

	padd(depth - 1)
	if isabove == 1 {
		fmt.Printf("/--<\n")
	} else if isabove == 0 {
		fmt.Printf("---<\n")
	} else {
		fmt.Printf("\\--<\n")
	}

	if head.isLeaf[1] == false {
		printTree(head.next[1], depth+1, -1)
	} else {
		padd(depth)
		fmt.Printf("\\--%q\n", head.char[1])
	}

}

func sortTree(nodes *[]*node, length int) {
	for i := 0; i < length; i++ {
		a := letterInfo[i].freq
		if a == -1 {
			a = (*nodes)[i].sum + 1
		}

		for j := i + 1; j < length; j++ {
			b := letterInfo[j].freq
			if b == -1 {
				b = (*nodes)[j].sum + 1
			}

			if a < b {
				tmp1 := letterInfo[j]
				letterInfo[j] = letterInfo[i]
				letterInfo[i] = tmp1
				tmp2 := (*nodes)[i]
				(*nodes)[i] = (*nodes)[j]
				(*nodes)[j] = tmp2
			}
		}
	}
}

func createTree(start int) *node {
	nodes := make([](*node), 256)

	for i := start; i > 0; i-- {
		j := i - 1

		var temp node
		temp.sum = 0

		if letterInfo[i].freq == -1 {
			temp.sum += nodes[i].sum
			temp.next[0] = nodes[i]
			temp.isLeaf[0] = false
		} else {
			temp.sum += letterInfo[i].freq
			temp.char[0] = letterInfo[i].char
			temp.isLeaf[0] = true
		}

		if letterInfo[j].freq == -1 {
			temp.sum += nodes[j].sum
			temp.next[1] = nodes[j]
			temp.isLeaf[1] = false
		} else {
			temp.sum += letterInfo[j].freq
			temp.char[1] = letterInfo[j].char
			temp.isLeaf[1] = true
		}
		nodes[j] = &temp
		letterInfo[i].freq = -1
		letterInfo[j].freq = -1

		sortTree(&nodes, i)
	}

	return nodes[0]
}

func main() {
	var f_in_name, f_out_name string
	mode := CompressMode
	var print_tree_bool bool
	args := os.Args[1:]

	for i := range args {
		switch args[i] {
		case "-c":
			mode = CompressMode
		case "-d":
			mode = DecompressMode

		case "-o":
			if i+1 >= len(args) {
				log.Fatal("-o must be followed by an output file")
			}
			f_out_name = args[i+1]
			i++

		case "-p":
			print_tree_bool = true
		default:
			f_in_name = args[i]
		}
	}

	f_in := os.Stdin
	if f_in_name != "" {
		f, err := os.Open(f_in_name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f_in = f
	}
	r := bufio.NewReader(f_in)

	f_out := os.Stdout
	if f_out_name != "" {
		f, err := os.Create(f_out_name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f_out = f
	}
	w := bufio.NewWriter(f_out)
	defer w.Flush()

	switch mode {
	case CompressMode:
		countLetters(r)
		f_in.Seek(0, io.SeekStart) // BUG: This won't work with stdin.
		r = bufio.NewReader(f_in)  // Discard existing buffer.

		start := sortLetters()
		head := createTree(start)

		if print_tree_bool {
			printTree(head, 1, 0)
		}

		compress(head, r, w, start)
	case DecompressMode:
		var c, c_index byte = 0, 8
		head := readTree(r, &c, &c_index)

		if print_tree_bool {
			printTree(head, 1, 0)
		}

		decompress(head, r, w, &c, &c_index)
	}
}
