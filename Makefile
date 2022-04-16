chibicc: main.go
	rm -f *.s
	go build -o chibicc .

test: chibicc
	./test.sh

clean:
	rm -f chibicc *.o *~ tmp* *.s

.PHONY: test clean
