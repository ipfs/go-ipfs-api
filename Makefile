GO=env GO111MODULE=on go
GONOMOD=env GO111MODULE=off go

.PHONY: deps
deps:
	$(GO) mod vendor
	$(GO) mod verify
		
# Set up test environment
.PHONY: testenv
WAIT=3
testenv:
	@echo "===================   preparing test env    ==================="
	( cd testenv ; make testenv )
	@echo "===================          done           ==================="

.PHONY: clean
clean:
	@echo "===================   stopping test env    ==================="
	( cd testenv ; make clean )
	@echo "===================          done           ==================="