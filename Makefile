EXAMPLE_DIRS = $(dir $(wildcard examples/*/.))
EXAMPLES = $(foreach dir,$(EXAMPLE_DIRS),$(shell basename $(dir)))
EXAMPLES_SO = $(EXAMPLES:%=%.so)
EXAMPLES_FMU = $(EXAMPLES:%=%.fmu)

.PHONY: $(EXAMPLES_SO)
$(EXAMPLES_SO):
	mkdir -p out/build
	go build -buildmode c-shared -o ./out/build/$@ ./examples/$(basename $@)

.PHONY: $(EXAMPLES_FMU)
$(EXAMPLES_FMU): name = $(basename $@)
$(EXAMPLES_FMU): temp = out/temp/$(name)
$(EXAMPLES_FMU): build-examples
	mkdir -p out/fmus $(temp)/binaries/linux64
	cp examples/$(name)/modelDescription.xml $(temp)
	cp out/build/$(name).so $(temp)/binaries/linux64
	cd $(temp) && zip -r ../../fmus/$@ * && cd -

.PHONY: build-examples
build-examples: $(EXAMPLES_SO)

.PHONY: example-fmus
example-fmus: $(EXAMPLES_FMU)

.PHONY: test
test:
	go test -tags=none -race -cover ./...

.PHONY: integration-test
integration-test: export TEST_FMU = ./out/fmus/BouncingBall.fmu
integration-test:
	python -m unittest -v -f test/integration_test.py