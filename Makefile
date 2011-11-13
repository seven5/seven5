
all: $(wildcard *.html)


%.html: %.md 
	markdown -css="http://kevinburke.bitbucket.org/markdowncss/markdown.css" $< $@