nw = $(shell which noweb)

all: fetch

fetch: fetch.go
	go build fetch.go
fetch.go: fetch.org
	if [ "$(nw)" != "" ]; then\
		bash ../scripts/org2nw fetch.org | notangle -Rfetch.go | gofmt > fetch.go;\
	fi
tangle: fetch.go fetch_test.go
.PHONY: test
test: fetch_test.go fetch.go
	go test -v
fetch_test.go: fetch.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../scripts/preTangle.awk fetch.org | bash ../scripts/org2nw | notangle -Rfetch_test.go | gofmt > fetch_test.go;\
	fi

clean:
	rm -f *.go
