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

func initCodes(head *node, codes *[256][24]int8, tempCodes *[24]int8, depth int) {
	for p := int8(0); p < 2; p++ {
		if head.isLeaf[p] {
			tempCodes[depth] = p
			tempCodes[depth+1] = -1
			i := 0
			for i = 0; tempCodes[i] != -1; i++ {
				codes[uint8(head.char[p])][i] = tempCodes[i]
			}
			codes[uint8(head.char[p])][i] = -1

		} else {
			tempCodes[depth] = p
			initCodes(head.next[p], codes, tempCodes, depth+1)
		}
	}
}

func writeBit(b, i byte) byte {
	return (b << 7) >> i
}

func readBit(c, i byte) byte {
	return (c & ((1 << 7) >> i)) >> (7 - i)
}

func checkWrite(w io.ByteWriter, c, cindex *byte) {
	if *cindex >= 8 {
		w.WriteByte(*c)
		*cindex = 0
		*c = 0
	}
}

func writeTree(head *node, w io.ByteWriter, c, cindex *byte) {
	for p := 0; p < 2; p++ {
		checkWrite(w, c, cindex)

		if head.isLeaf[p] == true {
			*c |= writeBit(1, *cindex)
			*cindex++
			for i := byte(0); i < 8; i++ {
				checkWrite(w, c, cindex)
				*c |= writeBit(readBit(head.char[p], i), *cindex)
				*cindex++
			}
			checkWrite(w, c, cindex)
		} else {
			*cindex++ //leaves the bit we were at at 0
			writeTree(head.next[p], w, c, cindex)
		}
	}
}

func compress(head *node, r io.ByteReader, w io.ByteWriter, amount int) {
	var codes [256][24]int8
	var tempCodes [24]int8
	var c, cindex byte

	writeTree(head, w, &c, &cindex)
	initCodes(head, &codes, &tempCodes, 0)

	for {
		ch, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// TODO: Handle errors.
		}
		tmp := uint8(ch)
		for i := 0; codes[tmp][i] != -1; i++ {
			checkWrite(w, &c, &cindex)

			c |= writeBit(byte(codes[tmp][i]), cindex)
			cindex++
		}
		checkWrite(w, &c, &cindex)
	}

	if cindex > 0 {
		w.WriteByte(c)
	}
}

func decompress(head *node, r io.ByteReader, w io.ByteWriter, c, cindex *byte) {
	parseNode := head

	for {
		if *cindex >= 8 {
			ch, err := r.ReadByte()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				// TODO: Handle errors.
			}
			*c = ch
			*cindex = 0
		}

		nextStep := readBit(*c, *cindex)
		*cindex++

		if parseNode.isLeaf[nextStep] == true {
			w.WriteByte(parseNode.char[nextStep])
			parseNode = head
		} else {
			parseNode = parseNode.next[nextStep]
		}
	}
}

func readTree(r io.ByteReader, c, cindex *byte) *node {
	var node node

	for p := 0; p < 2; p++ {
		if *cindex >= 8 {
			ch, err := r.ReadByte()
			if err != nil {
				log.Fatal(err)
			}

			*c = ch
			*cindex = 0
		}

		if readBit(*c, *cindex) != 1 {
			*cindex++
			node.isLeaf[p] = false
			node.next[p] = readTree(r, c, cindex)
			continue
		}

		node.isLeaf[p] = true
		node.next[p] = nil
		*cindex++
		for i := byte(0); i < 8; i++ {
			if *cindex >= 8 {
				ch, err := r.ReadByte()
				if err != nil {
					log.Fatal(err)
				}
				*c = ch
				*cindex = 0
			}

			node.char[p] |= writeBit(readBit(*c, *cindex), i) //copy bits over
			*cindex++
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
	var finName, foutName string
	mode := CompressMode
	var doPrintTree bool
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
			foutName = args[i+1]
			i++

		case "-p":
			doPrintTree = true
		default:
			finName = args[i]
		}
	}

	fin := os.Stdin
	if finName != "" {
		f, err := os.Open(finName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fin = f
	}
	r := bufio.NewReader(fin)

	fout := os.Stdout
	if foutName != "" {
		f, err := os.Create(foutName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fout = f
	}
	w := bufio.NewWriter(fout)
	defer w.Flush()

	switch mode {
	case CompressMode:
		countLetters(r)
		fin.Seek(0, io.SeekStart) // BUG: This won't work with stdin.
		r = bufio.NewReader(fin)  // Discard existing buffer.

		start := sortLetters()
		head := createTree(start)

		if doPrintTree {
			printTree(head, 1, 0)
		}

		compress(head, r, w, start)
	case DecompressMode:
		var c, cindex byte = 0, 8
		head := readTree(r, &c, &cindex)

		if doPrintTree {
			printTree(head, 1, 0)
		}

		decompress(head, r, w, &c, &cindex)
	}
}
