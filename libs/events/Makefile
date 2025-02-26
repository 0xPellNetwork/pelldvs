.PHONY: docs
REPO:=github.com/0xPellNetwork/pelldvs/libs/events

docs:
	@go get github.com/davecheney/godoc2md
	godoc2md $(REPO) > README.md

test:
	go test -v ./...
