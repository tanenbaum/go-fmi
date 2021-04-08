/*
    Bridge functions for fmi2Functions and cgo
*/

#include "./c/fmi2Functions.h"

void bridge_fmi2CallbackLogger(fmi2CallbackLogger f,
    fmi2ComponentEnvironment componentEnvironment,
    fmi2String instanceName,
    fmi2Status status,
    fmi2String category,
    fmi2String message);