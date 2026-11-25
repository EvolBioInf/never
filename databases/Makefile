all: README
README: README.html
README.html: README.org
	pandoc -o README.html README.org
clean:
	rm -f README.html
