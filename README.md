# go-fmi

Golang wrapper for building FMI compatible shared libraries in Go.

WIP - not ready for external presentation yet.

## FMI Implementation

A copy of the FMI documentation, model description schema and C headers are in `./third_party/fmi`.

As this will generate a shared object file the `FMI2_FUNCTION_PREFIX` is not set.
A tool will dynamically load this library and manually export function symbols.