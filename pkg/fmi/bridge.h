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

// various typedefs to handle const pointers in cgo
typedef const fmi2StatusKind fmi2StatusKind_t;
typedef const fmi2CallbackFunctions* fmi2CallbackFunctions_t;
typedef const fmi2String* strings_t;
typedef const fmi2ValueReference* valueReferences_t;
typedef const fmi2Real* fmi2Reals_t;
typedef const fmi2Integer* fmi2Integers_t;
typedef const fmi2Boolean* fmi2Booleans_t;
typedef const fmi2String* fmi2Strings_t;
typedef const fmi2Byte* serializedState_t;

void bridge_fmi2CallbackLogger(fmi2CallbackLogger f,
    fmi2ComponentEnvironment componentEnvironment,
    fmi2String instanceName,
    fmi2Status status,
    fmi2String category,
    fmi2String message);

#endif  /* bridge_h */