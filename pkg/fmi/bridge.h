/*
    Bridge functions for fmi2Functions and cgo
*/

#ifndef bridge_h
#define bridge_h

#include "./c/fmi2Functions.h"

// ModelState is used to encode model state
typedef struct {
    size_t size;
    char data[1];
} ModelState;

void bridge_fmi2CallbackLogger(fmi2CallbackLogger f,
    fmi2ComponentEnvironment componentEnvironment,
    fmi2String instanceName,
    fmi2Status status,
    fmi2String category,
    fmi2String message);

#endif  /* bridge_h */