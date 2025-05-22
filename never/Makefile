nw = $(shell which noweb)

all: never

never: never.go
	go build never.go
never.go: never.org
	if [ "$(nw)" != "" ]; then\
		bash ../scripts/org2nw never.org | notangle -Rnever.go | gofmt > never.go;\
	fi
tangle: never.go never_test.go
.PHONY: test
test: never_test.go never.go
	go test -v
never_test.go: never.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../scripts/preTangle.awk never.org | bash ../scripts/org2nw | notangle -Rnver_test.go | gofmt > never_test.go;\
	fi

clean:
	rm -f *.go
