
build:
	go build cmd/cfpx.go

install:
	go install cmd/cfpx.go

test:
	go test -v ./...

clean:
	rm -f cfpx
