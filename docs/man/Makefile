MANPAGES = \
		   sdr.7

.SUFFIXES:
.SUFFIXES: .7 .7.md

all: build

build: $(MANPAGES)

.7.md.7:
	ronn < $< > $@

clean:
	rm -fv $(MANPAGES)
