package fmi

// #include <stdlib.h>
// #include "./c/fmi2Functions.h"
// typedef const fmi2CallbackFunctions* fmi2CallbackFunctions_t;
// typedef const fmi2String* strings_t;
// typedef const fmi2ValueReference* valueReferences_t;
// typedef const fmi2Real* fmi2Reals_t;
// typedef const fmi2Integer* fmi2Integers_t;
// typedef const fmi2Boolean* fmi2Booleans_t;
// typedef const fmi2String* fmi2Strings_t;
// typedef const fmi2Byte* serializedState_t;
// typedef const fmi2StatusKind fmi2StatusKind_t;
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

const (
	FMUTypeModelExchange FMUType = iota
	FMUTypeCoSimulation
)

var (
	fmiVersion       = C.CString(C.fmi2Version)
	fmiTypesPlatform = C.CString(C.fmi2TypesPlatform)
	fmus             = map[FMUID]*FMU{}
)

type FMUType uint

type FMU struct {
	Name             string
	Typee            FMUType
	Guid             string
	ResourceLocation string
	Visible          bool
	LoggingOn        bool
}

type FMUID uintptr

func (f FMUID) asFMI2Component() C.fmi2Component {
	return C.fmi2Component(f)
}

//export fmi2GetVersion
func fmi2GetVersion() C.fmi2String {
	return fmiVersion
}

// GetVersion is wrapper for fmi2GetVersion
func GetVersion() string {
	return C.GoString(fmi2GetVersion())
}

//export fmi2GetTypesPlatform
func fmi2GetTypesPlatform() C.fmi2String {
	return fmiTypesPlatform
}

// GetTypesPlatform is wrapper for fmi2GetTypesPlatform
func GetTypesPlatform() string {
	return C.GoString(fmi2GetTypesPlatform())
}

func fmuBool(b C.fmi2Boolean) bool {
	return b == C.fmi2True
}

func boolFMU(b bool) C.fmi2Boolean {
	if b {
		return C.fmi2True
	}
	return C.fmi2False
}

//export fmi2Instantiate
func fmi2Instantiate(instanceName C.fmi2String, fmuType C.fmi2Type, fmuGUID C.fmi2String,
	fmuResourceLocation C.fmi2String, functions C.fmi2CallbackFunctions_t,
	visible C.fmi2Boolean, loggingOn C.fmi2Boolean) C.fmi2Component {
	id := FMUID(C.malloc(1))
	fmus[id] = &FMU{
		Name:             C.GoString(instanceName),
		Typee:            FMUType(fmuType),
		Guid:             C.GoString(fmuGUID),
		ResourceLocation: C.GoString(fmuResourceLocation),
		Visible:          fmuBool(visible),
		LoggingOn:        fmuBool(loggingOn),
	}
	return C.fmi2Component(id)
}

// Instantiate is Go wrapper for fmi2Instantiate
// callback functions use a default value
func Instantiate(instanceName string, fmuType FMUType, fmuGUID string,
	fmuResourceLocation string, visible bool, loggingOn bool) FMUID {
	n := C.CString(instanceName)
	g := C.CString(fmuGUID)
	r := C.CString(fmuResourceLocation)
	comp := fmi2Instantiate(n, C.fmi2Type(fmuType), g, r, &C.fmi2CallbackFunctions{}, boolFMU(visible), boolFMU(loggingOn))
	C.free(unsafe.Pointer(n))
	C.free(unsafe.Pointer(g))
	C.free(unsafe.Pointer(r))
	return FMUID(comp)
}

//export fmi2FreeInstance
func fmi2FreeInstance(c C.fmi2Component) {
	if c == nil {
		return
	}

	id, _, err := getFMU(c)
	if err != nil {
		return
	}

	delete(fmus, id)
	C.free(unsafe.Pointer(id))
}

// FreeInstance is a wrapper for fmi2FreeInstance
func FreeInstance(id FMUID) {
	fmi2FreeInstance(C.fmi2Component(id))
}

//export fmi2SetDebugLogging
func fmi2SetDebugLogging(c C.fmi2Component, logginOn C.fmi2Boolean,
	nCategories C.size_t, categories C.strings_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetupExperiment
func fmi2SetupExperiment(c C.fmi2Component, toleranceDefined C.fmi2Boolean,
	tolerance C.fmi2Real, startTime C.fmi2Real, stopTimeDefined C.fmi2Boolean,
	stopTime C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2EnterInitializationMode
func fmi2EnterInitializationMode(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2ExitInitializationMode
func fmi2ExitInitializationMode(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2Terminate
func fmi2Terminate(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2Reset
func fmi2Reset(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetReal
func fmi2GetReal(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Real) C.fmi2Status {
	// TODO: implement
	var vs []C.fmi2Real
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&vs), int(nvr))
	for i := 0; i < int(nvr); i++ {
		vs[i] = 1.0
	}
	return C.fmi2OK
}

//export fmi2GetInteger
func fmi2GetInteger(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Integer) C.fmi2Status {
	// TODO: implement
	var vs []C.fmi2Integer
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&vs), int(nvr))
	for i := 0; i < int(nvr); i++ {
		vs[i] = 1
	}
	return C.fmi2OK
}

//export fmi2GetBoolean
func fmi2GetBoolean(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Boolean) C.fmi2Status {
	// TODO: implement
	var vs []C.fmi2Boolean
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&vs), int(nvr))
	for i := 0; i < int(nvr); i++ {
		vs[i] = 1
	}
	return C.fmi2OK
}

//export fmi2GetString
func fmi2GetString(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2String) C.fmi2Status {
	// TODO: implement
	var vs []C.fmi2String
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&vs), int(nvr))
	for i := 0; i < int(nvr); i++ {
		vs[i] = C.CString("foo")
	}
	return C.fmi2OK
}

//export fmi2SetReal
func fmi2SetReal(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Reals_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetInteger
func fmi2SetInteger(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Integers_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetBoolean
func fmi2SetBoolean(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Booleans_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetString
func fmi2SetString(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Strings_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetFMUstate
func fmi2GetFMUstate(c C.fmi2Component, FMUstate *C.fmi2FMUstate) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetFMUstate
func fmi2SetFMUstate(c C.fmi2Component, FMUState C.fmi2FMUstate) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
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

//export fmi2GetDirectionalDerivative
func fmi2GetDirectionalDerivative(c C.fmi2Component, vUnknown_ref C.valueReferences_t, nUnknown C.size_t,
	vKnown_ref C.valueReferences_t, nKnown C.size_t,
	dvKnown C.fmi2Reals_t, dvUnknown *C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2EnterEventMode
func fmi2EnterEventMode(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2NewDiscreteStates
func fmi2NewDiscreteStates(c C.fmi2Component, fmi2eventInfo *C.fmi2EventInfo) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2EnterContinuousTimeMode
func fmi2EnterContinuousTimeMode(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2CompletedIntegratorStep
func fmi2CompletedIntegratorStep(c C.fmi2Component, noSetFMUStatePriorToCurrentPoint C.fmi2Boolean, enterEventMode, terminateSimulation *C.fmi2Boolean) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetTime
func fmi2SetTime(c C.fmi2Component, time C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetContinuousStates
func fmi2SetContinuousStates(c C.fmi2Component, x C.fmi2Reals_t, nx C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetDerivatives
func fmi2GetDerivatives(c C.fmi2Component, derivatives *C.fmi2Real, nx C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetEventIndicators
func fmi2GetEventIndicators(c C.fmi2Component, eventIndicators *C.fmi2Real, ni C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetContinuousStates
func fmi2GetContinuousStates(c C.fmi2Component, x *C.fmi2Real, nx C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetNominalsOfContinuousStates
func fmi2GetNominalsOfContinuousStates(c C.fmi2Component, x_nominal *C.fmi2Real, nx C.size_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2SetRealInputDerivatives
func fmi2SetRealInputDerivatives(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, order C.fmi2Integers_t, value C.fmi2Reals_t) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetRealOutputDerivatives
func fmi2GetRealOutputDerivatives(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, order C.fmi2Integers_t, value *C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2DoStep
func fmi2DoStep(c C.fmi2Component, currentCommunicationPoint, communicationStepSize C.fmi2Real, noSetFMUStatePriorToCurrentPoint C.fmi2Boolean) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2CancelStep
func fmi2CancelStep(c C.fmi2Component) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetStatus
func fmi2GetStatus(c C.fmi2Component, s C.fmi2StatusKind, value *C.fmi2Status) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetRealStatus
func fmi2GetRealStatus(c C.fmi2Component, s C.fmi2StatusKind, value *C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetIntegerStatus
func fmi2GetIntegerStatus(c C.fmi2Component, s C.fmi2StatusKind, value *C.fmi2Integer) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetBooleanStatus
func fmi2GetBooleanStatus(c C.fmi2Component, s C.fmi2StatusKind, value *C.fmi2Boolean) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

//export fmi2GetStringStatus
func fmi2GetStringStatus(c C.fmi2Component, s C.fmi2StatusKind, value *C.fmi2String) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

func getFMU(c C.fmi2Component) (id FMUID, fmu *FMU, err error) {
	id = FMUID(c)
	fmu, err = GetFMU(id)
	return
}

func GetFMU(id FMUID) (*FMU, error) {
	fmu, ok := fmus[id]
	if !ok {
		return nil, fmt.Errorf("FMU %v not found", id)
	}
	return fmu, nil
}

func carrayToSlice(carray unsafe.Pointer, slice unsafe.Pointer, len int) {
	sliceHeader := (*reflect.SliceHeader)(slice)
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(carray)
}
