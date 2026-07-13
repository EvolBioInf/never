SUBDIRS := testing
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
tangle: never.go $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: test

test:
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir test; \
	done

clean:
	rm -f *.go static/templates.html rm -f
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir clean; \
	done