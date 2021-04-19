package fmi

// #include <stdlib.h>
// #include "./c/fmi2Functions.h"
// #include "bridge.h"
// typedef const fmi2Byte* serializedState_t;
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

	p := C.fmi2FMUstate(unsafe.Pointer(ms))
	FMUstate = &p
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
	return C.fmi2OK
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
func fmi2FreeFMUstate(c C.fmi2Component, FMUState *C.fmi2FMUstate) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SerializedFMUstateSize
func fmi2SerializedFMUstateSize(c C.fmi2Component, FMUState C.fmi2FMUstate, size *C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SerializeFMUstate
func fmi2SerializeFMUstate(c C.fmi2Component, FMUstate C.fmi2FMUstate, serializedState *C.fmi2Byte, size C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2DeSerializeFMUstate
func fmi2DeSerializeFMUstate(c C.fmi2Component, serializedState C.serializedState_t, size C.size_t, FMUstate *C.fmi2FMUstate) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}
