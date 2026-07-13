nw = $(shell which noweb)
all: never_test.go

.PHONY: test
test: never_test.go
	go test -v
never_test.go: testNever.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../../../../scripts/preTangle.awk testNever.org | bash ../../../../scripts/org2nw | notangle -Rnever_test.go | gofmt > never_test.go;\
	fi

clean:
	rm -f *.go
