nw = $(shell which noweb)

all: never.go static/templates.html

never.go: never.org
	if [ "$(nw)" != "" ]; then\
		bash ../../../scripts/org2nw never.org | notangle -Rnever.go | gofmt > never.go;\
	fi
static/templates.html: never.org
	if [ "$(nw)" != "" ]; then\
		bash ../../../scripts/org2nw never.org | notangle -Rtemplates.html > static/templates.html;\
	fi
tangle: never.go never_test.go
.PHONY: 1test
1test: never_test.go never.go
	go test -v
never_test.go: testNever.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../../../scripts/preTangle.awk testNever.org | bash ../../../scripts/org2nw | notangle -Rnever_test.go | gofmt > never_test.go;\
	fi

clean:
	rm -f *.go static/templates.html
