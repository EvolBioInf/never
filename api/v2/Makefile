nw = $(shell which noweb)

all : api

api: api.go
	go build api.go
api.go: api.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../../scripts/preTangle.awk api.org | bash ../../scripts/org2nw | notangle -Rapi.go | gofmt > api.go;\
	fi
tangle: api.go api_test.go
.PHONY: test
test: api_test.go api.go
	go test -v
api_test.go: api.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../../scripts/preTangle.awk api.org | bash ../../scripts/org2nw | notangle -Rapi_test.go | gofmt > api_test.go;\
	fi

clean:
	rm -f *.go
