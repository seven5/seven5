#
# TODO: figure out how to express this such that you don't need to "touch" the
# TODO: before make becomes aware of it.  Arg!

all: $(wildcard *.html)

%.html: %.md 
	markdown -css="css/base.css" $< $@
