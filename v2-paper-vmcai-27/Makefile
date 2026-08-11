# Makefile for Spectre v2 VMCAI 2027 paper
# Requires: pdflatex + bibtex (or latexmk)
# LaTeX class: uses ../paper-vmcai-2027/llncs.cls (relative path from this dir)

PAPER = main
LATEX = pdflatex
BIBTEX = bibtex

.PHONY: all clean view watch

all: $(PAPER).pdf

$(PAPER).pdf: $(PAPER).tex references.bib
	$(LATEX) -interaction=nonstopmode $(PAPER)
	$(BIBTEX) $(PAPER)
	$(LATEX) -interaction=nonstopmode $(PAPER)
	$(LATEX) -interaction=nonstopmode $(PAPER)

# Convenience: latexmk handles reruns automatically
latexmk:
	latexmk -pdf -bibtex -interaction=nonstopmode $(PAPER)

view: $(PAPER).pdf
	open $(PAPER).pdf

watch:
	latexmk -pdf -bibtex -pvc $(PAPER)

clean:
	rm -f $(PAPER).pdf $(PAPER).aux $(PAPER).bbl $(PAPER).blg \
	      $(PAPER).log $(PAPER).out $(PAPER).toc $(PAPER).fls \
	      $(PAPER).fdb_latexmk $(PAPER).synctex.gz
