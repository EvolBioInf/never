version = $(shell bash scripts/getVersion.sh)
date = $(shell bash scripts/getDate.sh)
nw = $(shell which noweb)

all: main 

main: main.go
	go build -ldflags "-X github.com/evolbioinf/never/util.version=$(version) -X github.com/evolbioinf/never/util.date=$(date)" main.go

main.go: main.org
	if [ "$(nw)" != "" ]; then\
		bash scripts/org2nw main.org | notangle -Rmain.go | gofmt > main.go;\
	fi
tangle: main.go
.PHONY: test

