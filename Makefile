GO_TEST = go test -coverprofile=coverage.out
REPO = github.com/jbgo/sftbot

default: build test coverage

build: *.go command/*.go
	go build -o sftbot *.go

coverage: coverage/db coverage/trading

coverage/db:
	go tool cover -html=tmp/db.coverage -o tmp/db.coverage.html

coverage/trading:
	go tool cover -html=tmp/trading.coverage -o tmp/trading.coverage.html

deploy:
	/bin/bash -c 'echo TODO: aws s3 cp ./sftbot s3://SOMEWHERE/sftbot/sftbot-`./sftbot -version 2>&1`'

setup:
	cp -n .creds.json.example .creds.json || true
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

test: test/trading test/db

test/db:
	go test -coverprofile=tmp/db.coverage $(REPO)/db

test/trading:
	go test -coverprofile=tmp/trading.coverage $(REPO)/trading
