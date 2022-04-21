FILES = main.go
SRC = $(addprefix src/, $(FILES))
EXEC = huff
GOC = gccgo

default:
	$(GOC) $(SRC) -o $(EXEC)

.PHONY: clean

clean:
	rm $(EXEC)
