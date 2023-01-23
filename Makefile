clean:
	rm registry
build:
	go build . && mv just registry
debug:
	go build .
	mv just debug/bin
run:
	cd debug && go run ../ serve
