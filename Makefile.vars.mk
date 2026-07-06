IMG_TAG ?= latest

CURDIR ?= $(shell pwd)
BIN_FILENAME ?= transmuter

# Image URL to use all building/pushing image targets
GHCR_IMG ?= ghcr.io/helmetica-framework/transmuter:$(IMG_TAG)
