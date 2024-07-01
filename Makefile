GO ?= go

.PHONY: binary

binary: dist FORCE
	$(GO) version
ifeq ($(OS),Windows_NT)
	$(GO) build  -o dist/swaplist.exe .
else
	$(GO) build -o dist/swaplist .
endif

dist:
	mkdir $@

FORCE:
