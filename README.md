# Huffman coding
A simple program to encode/decode files with huffman coding in go.

## usage
- **[-c]** \[file\] to encode a file (default option). If no file is specified the program will read from stdin.
- **[-d]** \[file\] to decode a file. If no file is specified the program will read from stdin.
- **[-o]** <file\> to specify an output file. If no file is specified thee
- **[-p]** to print out the huffman tree

## examples

> ./huff file.txt | ./huff -d

simple way to test if the application works

>  ./huff file.txt -c -o file.huff
or
> ./huff -c file.txt -o file.huff

encode a file

> ./huff file.huff -d -o file.txt
or
> ./huff file.huff -d -o file.txt

decode a file

## compiling
### Unix-likes
> $ git clone https://github.com/atiedebee/huffman-coding
> $ cd huffman-coding
> $ make

### Windows
> $ git clone https://github.com/atiedebee/huffman-coding
> $ cd huffman-coding/src
> $ go build

or something like that, I don't use windows
