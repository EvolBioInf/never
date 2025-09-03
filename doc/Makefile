text = neverDoc
SRC1 = fetch.tex never.tex testNever.tex util.tex
SRC2 = header.tex intro.tex neverDoc.tex ../never/index.tex

date = $(shell git log | grep -m 1 Date | sed -r 's/Date: +[A-Z][a-z]+ ([A-Z][a-z]+) ([0-9]+) [^ ]+ ([0-9]+) .+/\2_\1_\3/')
version = $(shell git describe)

all: $(text).pdf $(SRC1) $(SRC2)
%.tex:	../*/%.org
	bash ../scripts/org2nw $< | awk -f ../scripts/preWeave.awk | noweave -n -x | sed 's/_test/\\_test/g' > $@
$(text).pdf: $(text).tex $(SRC1) $(SRC2)
	echo $(date) | tr '_' ' ' > date.txt
	echo $(version) | tr '-' ' ' | awk '{printf "%s", $$1; if ($$2) printf "-%s", $$2; printf "\n"}' > version.txt
	latex $(text).tex
	bibtex $(text)
	latex $(text).tex
	latex $(text).tex
	dvips $(text) -o -q
	ps2pdf -dALLOWPSTRANSPARENCY $(text).ps
publish: $(text).pdf
	if [ -d ~/WinHome/ownCloud/ ]; then \
		cp $(text).pdf ~/WinHome/ownCloud/docs/; \
	fi

clean:
	rm -f $(text).pdf $(text).ps $(text).dvi $(text).aux $(SRC1) 
