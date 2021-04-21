# go-fmi

Golang wrapper for building FMI compatible shared libraries in Go.

WIP - not ready for external presentation yet.

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