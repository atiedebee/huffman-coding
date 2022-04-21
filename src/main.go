package main

import (
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

var letter_info [256]letters

type MODE uint8

const (
	COMPRESS MODE = iota
	DECOMPRESS
)

func count_letters(file *os.File) {
	ch := make([]byte, 1)

	n, _ := file.Read(ch)
	for n == 1 {
		letter_info[ch[0]].freq += 1
		letter_info[ch[0]].char = ch[0]
		n, _ = file.Read(ch)
	}
}

func sort_letters() int {
	var i int
	for i = 0; i < len(letter_info); i++ {
		for j := i + 1; j < len(letter_info); j++ {
			if letter_info[i].freq < letter_info[j].freq {
				char := letter_info[i].char
				freq := letter_info[i].freq
				letter_info[i].freq = letter_info[j].freq
				letter_info[i].char = letter_info[j].char
				letter_info[j].freq = freq
				letter_info[j].char = char
			}
		}
		if letter_info[i].freq == 0 {
			break
		}
	}

	return i - 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init_codes(head *node, codes *[256][24]int8, temp_codes *[24]int8, depth int) {
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
			init_codes(head.next[p], codes, temp_codes, depth+1)
		}
	}
}

func write_bit(b, i byte) byte {
	return (b << 7) >> i
}

func read_bit(c, i byte) byte {
	return (c & ((1 << 7) >> i)) >> (7 - i)
}

func check_write(f_out *os.File, c, c_index *byte) {

	if *c_index >= 8 {
		f_out.Write([]byte{*c})
		*c_index = 0
		*c = 0
	}
}

func write_tree(head *node, f_out *os.File, c, c_index *byte) {

	for p := 0; p < 2; p++ {
		check_write(f_out, c, c_index)

		if head.isLeaf[p] == true {
			*c |= write_bit(1, *c_index)
			*c_index++
			for i := byte(0); i < 8; i++ {
				check_write(f_out, c, c_index)
				*c |= write_bit(read_bit(head.char[p], i), *c_index)
				*c_index++
			}
			check_write(f_out, c, c_index)
		} else {
			*c_index++ //leaves the bit we were at at 0
			write_tree(head.next[p], f_out, c, c_index)
		}
	}
}

func compress(head *node, f_in, f_out *os.File, amount int) {
	var codes [256][24]int8
	var temp_codes [24]int8
	var c, c_index byte = 0, 0
	ch_in := make([]byte, 1)

	write_tree(head, f_out, &c, &c_index)
	init_codes(head, &codes, &temp_codes, 0)

	read_size, _ := f_in.Read(ch_in)
	for read_size == 1 {
		tmp := uint8(ch_in[0])
		for i := 0; codes[tmp][i] != -1; i++ {
			check_write(f_out, &c, &c_index)

			c |= write_bit(byte(codes[tmp][i]), c_index)
			c_index++
		}
		check_write(f_out, &c, &c_index)
		read_size, _ = f_in.Read(ch_in)
	}

	if c_index > 0 {
		f_out.Write([]byte{c})
	}
}

func decompress(head *node, f_in, f_out *os.File, c, c_index *byte) {
	var parse_node = head
	var n int = 1

	if *c_index >= 8 {
		ch := make([]byte, 1)
		n, _ = f_in.Read(ch)
		*c = ch[0]
		*c_index = 0
	}

	for n == 1 {
		next_step := read_bit(*c, *c_index)

		if parse_node.isLeaf[next_step] == true {
			f_out.Write([]byte{parse_node.char[next_step]})
			parse_node = head

		} else {
			parse_node = parse_node.next[next_step]
		}

		*c_index++
		if *c_index >= 8 {
			ch := make([]byte, 1)

			n, _ = f_in.Read(ch)

			*c = ch[0]
			*c_index = 0
		}
	}
}

func read_tree(f_in *os.File, c, c_index *byte) *node {
	node := new(node)

	for p := 0; p < 2; p++ {

		if *c_index >= 8 {
			ch := make([]byte, 1)

			n, err := f_in.Read(ch)
			if n != 1 {
				log.Fatal(err)
			}

			*c = ch[0]
			*c_index = 0
		}

		if read_bit(*c, *c_index) == 1 {
			node.isLeaf[p] = true
			node.next[p] = nil
			*c_index++
			for i := byte(0); i < 8; i++ {
				if *c_index >= 8 {
					ch := make([]byte, 1)

					n, err := f_in.Read(ch)
					if n != 1 {
						log.Fatal(err)
					}
					*c = ch[0]
					*c_index = 0
				}

				node.char[p] |= write_bit(read_bit(*c, *c_index), i) //copy bits over
				*c_index++
			}
		} else {
			*c_index++
			node.isLeaf[p] = false
			node.next[p] = read_tree(f_in, c, c_index)
		}
	}

	return node
}

func padd(depth int) {
	const padding string = "    "
	for i := 0; i < depth; i++ {
		fmt.Printf(padding)
	}
}

func print_tree(head *node, depth int, isabove int) {
	if head.isLeaf[0] == false {
		print_tree(head.next[0], depth+1, 1)
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
		print_tree(head.next[1], depth+1, -1)
	} else {
		padd(depth)
		fmt.Printf("\\--%q\n", head.char[1])
	}

}

func sort_tree(nodes *[]*node, length int) {
	for i := 0; i < length; i++ {

		a := letter_info[i].freq
		if a == -1 {
			a = (*nodes)[i].sum + 1
		}

		for j := i + 1; j < length; j++ {

			b := letter_info[j].freq
			if b == -1 {
				b = (*nodes)[j].sum + 1
			}

			if a < b {
				tmp1 := letter_info[j]
				letter_info[j] = letter_info[i]
				letter_info[i] = tmp1
				tmp2 := (*nodes)[i]
				(*nodes)[i] = (*nodes)[j]
				(*nodes)[j] = tmp2
			}
		}
	}
}

func create_tree(start int) *node {
	var temp *node = nil
	nodes := make([](*node), 256)

	for i := start; i > 0; i-- {
		j := i - 1

		temp = new(node)
		temp.sum = 0

		if letter_info[i].freq == -1 {
			temp.sum += nodes[i].sum
			temp.next[0] = nodes[i]
			temp.isLeaf[0] = false
		} else {
			temp.sum += letter_info[i].freq
			temp.char[0] = letter_info[i].char
			temp.isLeaf[0] = true
		}

		if letter_info[j].freq == -1 {
			temp.sum += nodes[j].sum
			temp.next[1] = nodes[j]
			temp.isLeaf[1] = false
		} else {
			temp.sum += letter_info[j].freq
			temp.char[1] = letter_info[j].char
			temp.isLeaf[1] = true
		}
		nodes[j] = temp
		letter_info[i].freq = -1
		letter_info[j].freq = -1

		sort_tree(&nodes, i)
	}

	return nodes[0]
}

func main() {
	var f_in_name, f_out_name *string = nil, nil
	var mode MODE = COMPRESS
	var err error = nil
	var print_tree_bool bool = false
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c":
			mode = COMPRESS
		case "-d":
			mode = DECOMPRESS

		case "-o":
			if i+1 >= len(args) {
				log.Fatal("-o must be followed by an output file")
			}
			f_out_name = &args[i+1]
			i++

		case "-p":
			print_tree_bool = true
		default:
			f_in_name = &args[i]
		}
	}

	f_in := os.Stdin
	if f_in_name != nil {
		f_in, err = os.Open(*f_in_name)
		if err != nil {
			log.Fatal(err)
		}
		defer f_in.Close()
	}

	f_out := os.Stdout
	if f_out_name != nil {
		f_out, err = os.Create(*f_out_name)
		if err != nil {
			log.Fatal(err)
		}
		defer f_out.Close()
	}

	switch mode {
	case COMPRESS:
		count_letters(f_in)
		f_in.Seek(0, io.SeekStart)

		start := sort_letters()
		head := create_tree(start)

		if print_tree_bool {
			print_tree(head, 1, 0)
		}

		compress(head, f_in, f_out, start)
	case DECOMPRESS:
		var c, c_index byte = 0, 8
		head := read_tree(f_in, &c, &c_index)

		if print_tree_bool {
			print_tree(head, 1, 0)
		}

		decompress(head, f_in, f_out, &c, &c_index)
	default:
		return
	}
}
