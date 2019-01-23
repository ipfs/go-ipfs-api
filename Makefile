all: deps
gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go
deps: gx
	gx --verbose install --global
	gx-go rewrite
.PHONY: all gx deps

# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	( cd testenv ; make testenv )
	@echo "===================          done           ==================="
