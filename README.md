# go-fmi

*WIP - not quite ready for use yet. I'll tag an initial release and update the README/docs when this is stable.*

Golang wrapper for building FMI compatible shared libraries in Go.

Initially supporting FMI 2.0. If warranted, I'd be happy to add FMI 1.0 compatibility too.

Will look at FMI 3.0 once it has stabilised.

## Platforms

The steps to generate the FMU zips in the Makefile are designed specifically for Linux amd64 shared libraries.

In theory, it should be a simple to modify this and create DLLs using Go, targetting Windows amd64.

## FMI Implementation

A copy of the FMI documentation, model description schema and C headers are in `./third_party/fmi`.

As this will generate a shared object file the `FMI2_FUNCTION_PREFIX` is not set.
A tool will dynamically load this library and manually export function symbols.

## Integration Tests

Integration tests use the Python 3.x [fmpy](https://github.com/CATIA-Systems/FMPy) library.

Install these dependencies before validating the FMU.

Run `make integration-test` to execute the integration tests.