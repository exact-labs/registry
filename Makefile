clean:
	rm just_registry
build:
	go build registry/main.go && mv main just_registry
run:
	go run registry/main.go serve
