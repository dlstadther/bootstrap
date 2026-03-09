.PHONY: build

build:
	cd ansible && mise install && uv sync
