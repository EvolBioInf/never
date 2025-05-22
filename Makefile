all: never

never: bin/never
bin/never: never/never.org
	test -d bin || mkdir bin
	make -C never
	cp never/never bin
doc:
	make -C doc

PHONY.: clean

clean:
	make clean -C never
	make clean -C doc
