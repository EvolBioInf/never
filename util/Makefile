nw = $(shell which noweb)
SRCS = $(wildcard *.org)
GOFILES = $(patsubst %.org, %.go, $(SRCS))

all: $(GOFILES)

%.go: %.org
	if [ "$(nw)" != "" ]; then\
		awk -f ../scripts/preTangle.awk $< | bash ../scripts/org2nw | notangle -R$@ | gofmt > $@;\
	fi
tangle: $(GOFILES)

clean:
	rm -f *.go
