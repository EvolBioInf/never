nw = $(shell which noweb)

all : docs

docs: docs.go
	go build docs.go
docs.go: docs.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../scripts/preTangle.awk docs.org | bash ../scripts/org2nw | notangle -Rdocs.go | gofmt > docs.go;\
	fi
tangle: docs.go docs_test.go
.PHONY: test
test: docs_test.go docs.go
	go test -v
docs_test.go: docs.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../scripts/preTangle.awk docs.org | bash ../scripts/org2nw | notangle -Rdocs_test.go | gofmt > docs_test.go;\
	fi

clean:
	rm -f *.go
