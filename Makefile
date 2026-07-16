SUBDIRS := api/v1/never api/v2 docs/v2 fetch util
TESTDIRS := api/v1/never api/v2 fetch
version = $(shell bash scripts/getVersion.sh)
date = $(shell bash scripts/getDate.sh)
nw = $(shell which noweb)

all: main $(SUBDIRS)

.PHONY: all clean $(SUBDIRS) test test_db

main: main.go
	go build -ldflags "-X github.com/evolbioinf/never/util.version=$(version) -X github.com/evolbioinf/never/util.date=$(date)" main.go && \
	mkdir -p bin && \
	cp main bin/never 

main.go: main.org $(SUBDIRS)
	if [ "$(nw)" != "" ]; then\
		bash scripts/org2nw main.org | notangle -Rmain.go | gofmt > main.go;\
	fi

api/v2: fetch util
api/v1/never: fetch util
docs/v2: util
fetch: util

tangle: main.go

$(SUBDIRS):
	$(MAKE) -C $@

clean:
	for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir clean; \
	done ; \
	rm -rf main.go main bin testing/testdb ; \

test: test_db main
	./main -d testing/testdb/ -p 8008 --no-rate-limit & \
	SERVER_PID=$$! ; \
	trap "kill $$SERVER_PID" EXIT ; \
	sleep 1 ; \
	for dir in $(TESTDIRS); do \
		$(MAKE) -C $$dir test; \
	done

test_db: testing/testdb

testing/testdb: testing/testdb_small_dump.sql testing/testdb_large_dump.sql
	cd testing && \
	mkdir -p testdb && \
	sqlite3 testdb/small < testdb_small_dump.sql && \
	sqlite3 testdb/large < testdb_large_dump.sql
