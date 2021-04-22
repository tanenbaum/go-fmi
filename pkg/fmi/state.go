package fmi

// #include <stdlib.h>
// #include <string.h>
// #include "./c/fmi2Functions.h"
// #include "bridge.h"
import "C"

import (
	"fmt"
	"unsafe"
)

//export fmi2GetFMUstate
func fmi2GetFMUstate(c C.fmi2Component, FMUstate *C.fmi2FMUstate) C.fmi2Status {
	bs, s := GetFMUState(FMUID(c))
	if s != StatusOK {
		return C.fmi2Status(s)
	}

	bsize := len(bs)
	ms := (*C.ModelState)(C.malloc(C.ulong(C.sizeof_ModelState + bsize)))
	ms.size = C.ulong(bsize)
	var cs []C.char
	carrayToSlice(unsafe.Pointer(&ms.data[0]), unsafe.Pointer(&cs), bsize)
	for i, b := range bs {
		cs[i] = C.char(b)
	}

	*FMUstate = C.fmi2FMUstate(ms)
	return C.fmi2Status(s)
}

/*
GetFMUstate makes a copy of the internal FMU state and returns a byte array.
(FMUstate). If on entry *FMUstate == NULL, a new allocation is required. If *FMUstate !=
NULL , then *FMUstate points to a previously returned FMUstate that has not been modified
since. In particular, fmi2FreeFMUstate had not been called with this FMUstate as an
argument. [Function fmi2GetFMUstate typically reuses the memory of this FMUstate in this
case and returns the same pointer to it, but with the actual FMUstate .]
*/
func GetFMUState(id FMUID) ([]byte, Status) {
	fmu, ok := allowedSerialize(id, "GetFMUState")
	if !ok {
		return nil, StatusError
	}

	se, err := fmu.StateEncoder()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	bs, err := se.Encode()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	return bs, StatusOK
}

//export fmi2SetFMUstate
func fmi2SetFMUstate(c C.fmi2Component, FMUState C.fmi2FMUstate) C.fmi2Status {
	ms := (*C.ModelState)(FMUState)
	var cs []C.char
	carrayToSlice(unsafe.Pointer(&ms.data[0]), unsafe.Pointer(&cs), int(ms.size))
	bs := make([]byte, int(ms.size))
	for i, c := range cs {
		bs[i] = byte(c)
	}

	return C.fmi2Status(SetFMUState(FMUID(c), bs))
}

/*
SetFMUstate copies the content of the previously copied FMUstate back and uses it as
actual new FMU state. The FMUstate copy still exists.
*/
func SetFMUState(id FMUID, bs []byte) Status {
	fmu, ok := allowedSerialize(id, "GetFMUState")
	if !ok {
		return StatusError
	}

	se, err := fmu.StateDecoder()
	if err != nil {
		fmu.logger.Error(err)
		return StatusError
	}

	if err := se.Decode(bs); err != nil {
		fmu.logger.Error(fmt.Errorf("Error decoding state: %w", err))
		return StatusError
	}

	return StatusOK
}

//export fmi2FreeFMUstate
/*
fmi2FreeFMUstate frees all memory and other resources allocated with the
fmi2GetFMUstate call for this FMUstate . The input argument to this function is the FMUstate
to be freed. If a null pointer is provided, the call is ignored. The function returns a null pointer in
argument FMUstate
*/
func fmi2FreeFMUstate(c C.fmi2Component, FMUState *C.fmi2FMUstate) C.fmi2Status {
	if _, ok := allowedSerialize(FMUID(c), "FreeFMUState"); !ok {
		return C.fmi2Error
	}
	ms := (*C.ModelState)(*FMUState)
	if ms == nil {
		return C.fmi2OK
	}
	C.free(unsafe.Pointer(ms))
	*FMUState = nil
	return C.fmi2OK
}

//export fmi2SerializedFMUstateSize
/*
fmi2SerializedFMUstateSize returns the size of the byte vector, in order that FMUstate
can be stored in it. With this information, the environment has to allocate an fmi2Byte vector of
the required length size.
*/
func fmi2SerializedFMUstateSize(c C.fmi2Component, FMUState C.fmi2FMUstate, size *C.size_t) C.fmi2Status {
	fmu, ok := allowedSerialize(FMUID(c), "SerializedFMUstateSize")
	if !ok {
		return C.fmi2Error
	}
	ms := (*C.ModelState)(FMUState)
	if ms == nil {
		fmu.logger.Error(fmt.Errorf("Invalid argument %s = NULL", "FMUState"))
		return C.fmi2Error
	}
	*size = ms.size
	return C.fmi2OK
}

//export fmi2SerializeFMUstate
/*
fmi2SerializeFMUstate serializes the data which is referenced by pointer FMUstate and
copies this data in to the byte vector serializedState of length size, that must be provided
by the environment.
*/
func fmi2SerializeFMUstate(c C.fmi2Component, FMUstate C.fmi2FMUstate, serializedState *C.fmi2Byte, size C.size_t) C.fmi2Status {
	fmu, ok := allowedSerialize(FMUID(c), "SerializeFMUstate")
	if !ok {
		return C.fmi2Error
	}
	ms := (*C.ModelState)(FMUstate)
	if ms == nil {
		fmu.logger.Error(fmt.Errorf("Invalid argument %s = NULL", "FMUState"))
		return C.fmi2Error
	}

	if ms.size != size {
		fmu.logger.Error(fmt.Errorf("Model state size argument %d does not match %d", size, ms.size))
		return C.fmi2Error
	}

	var bs []byte
	carrayToSlice(unsafe.Pointer(serializedState), unsafe.Pointer(&bs), int(size))
	var cs []C.char
	carrayToSlice(unsafe.Pointer(&ms.data[0]), unsafe.Pointer(&cs), int(ms.size))
	for i, c := range cs {
		bs[i] = byte(c)
	}
	return C.fmi2OK
}

//export fmi2DeSerializeFMUstate
/*
fmi2DeSerializeFMUstate deserializes the byte vector serializedState of length size,
constructs a copy of the FMU state and returns FMUstate, the pointer to this copy.
*/
func fmi2DeSerializeFMUstate(c C.fmi2Component, serializedState C.serializedState_t, size C.size_t, FMUstate *C.fmi2FMUstate) C.fmi2Status {
	_, ok := allowedSerialize(FMUID(c), "DeSerializeFMUstate")
	if !ok {
		return C.fmi2Error
	}
	ms := (*C.ModelState)(C.malloc(C.ulong(C.sizeof_ModelState + size)))
	ms.size = C.ulong(size)
	C.memcpy(unsafe.Pointer(&ms.data[0]), unsafe.Pointer(serializedState), size)
	*FMUstate = C.fmi2FMUstate(ms)
	return C.fmi2OK
}

func allowedSerialize(id FMUID, name string) (*FMU, bool) {
	const expected = ModelStateInstantiated | ModelStateInitializationMode |
		ModelStateEventMode | ModelStateContinuousTimeMode |
		ModelStateStepComplete | ModelStateStepFailed | ModelStateStepCanceled |
		ModelStateTerminated | ModelStateError
	return allowedState(id, name, expected)
}
