GO_TEST = go test -coverprofile=coverage.out
REPO = github.com/jbgo/sftbot

build: *.go command/*.go
	go build -o sftbot *.go

coverage:
	go tool cover -html=coverage.out

test: test/trading

test/trading:
	$(GO_TEST) $(REPO)/trading

test/db:
	$(GO_TEST) $(REPO)/db
