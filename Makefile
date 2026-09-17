.PHONY: all web binary test help

GOOS ?= $(shell go env GOOS)
ifeq ($(GOOS),windows)
EXE := .exe
else
EXE :=
endif

OUT ?= mesosphere$(EXE)

all: binary

help:
	@echo "make web     - build frontend into web/dist and copy to pkg/server/dist"
	@echo "make binary  - web + go build -o $(OUT)"
	@echo "make test    - go test ./..."

web:
	cd web && npm run build
	rm -rf pkg/server/dist
	cp -R web/dist pkg/server/dist

binary: web
	go build -o $(OUT) .

test:
	go test ./...
