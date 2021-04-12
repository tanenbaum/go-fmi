package fmi

// #include <stdlib.h>
// #include "./c/fmi2Functions.h"
// #include "bridge.h"
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
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	fmiVersion       = C.CString(C.fmi2Version)
	fmiTypesPlatform = C.CString(C.fmi2TypesPlatform)
	// fmus stores all active FMUs at runtime
	fmus = map[FMUID]*FMU{}
	// models stores registered models
	models = map[string]Model{}
)

// FMUID holds a simple pointer that can be shared from this library to the calling system
// The id is mapped internally to the actual FMU stored in Go memory
type FMUID uintptr

func (f FMUID) asFMI2Component() C.fmi2Component {
	return C.fmi2Component(f)
}

// RegisterModel registers a model implementation and description with this FMI implementation.
// Multiple separate models can be registered, as long as they have different GUIDs.
// When Instantiated, the model will be looked up by GUID in the generated modelDescription.xml file in the FMI.
func RegisterModel(model Model) error {
	desc := model.Description()
	if desc.GUID == "" {
		return errors.New("Model description GUID cannot be empty")
	}

	if _, got := models[desc.GUID]; got {
		return fmt.Errorf("Model for GUID %s already registered", desc.GUID)
	}

	models[desc.GUID] = model
	return nil
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
	_ C.fmi2Boolean, loggingOn C.fmi2Boolean) C.fmi2Component {
	name := C.GoString(instanceName)

	logger := func(status Status, category, message string) {
		n := C.CString(name)
		c := C.CString(category)
		m := C.CString(message)
		defer C.free(unsafe.Pointer(n))
		defer C.free(unsafe.Pointer(c))
		defer C.free(unsafe.Pointer(m))
		C.bridge_fmi2CallbackLogger(functions.logger, functions.componentEnvironment, n, C.fmi2Status(status), c, m)
	}
	return Instantiate(
		name,
		FMUType(fmuType),
		C.GoString(fmuGUID),
		C.GoString(fmuResourceLocation),
		fmuBool(loggingOn), logger)
}

/*
Instantiate returns a new instance of an FMU. If a null pointer is returned, then instantiation
failed. In that case, `functions->logger` is called with detailed information about the
reason. An FMU can be instantiated many times (provided capability flag canBeInstantiatedOnlyOncePerProcess = false).

This function must be called successfully before any of the following functions can be called.
For co-simulation, this function call has to perform all actions of a slave which are necessary
before a simulation run starts (for example, loading the model file, compilation...).

Argument `instanceName` is a unique identifier for the FMU instance. It is used to name the
instance, for example, in error or information messages generated by one of the fmi2XXXFunctional Mock-up Interface 2.0.2
functions. It is not allowed to provide a null pointer and this string must be non-empty (in
other words, must have at least one character that is no white space). [If only one FMU is
simulated, as instanceName attribute modelName or <ModelExchange/CoSimulation modelIdentifier=”..”> from the XML schema fmiModelDescription might be used.]

Argument `fmuType` defines the type of the FMU:

- fmi2ModelExchange : FMU with initialization and events; between events simulation
of continuous systems is performed with external integrators from the environment

- fmi2CoSimulation : Black box interface for co-simulation.

Argument `fmuGUID` is used to check that the modelDescription.xml file is compatible with the C code of the FMU.
It is a vendor specific globally unique identifier of the
XML file (for example, it is a “fingerprint” of the relevant information stored in the XML file). It
is stored in the XML file as attribute “guid” and has to be passed to the
fmi2Instantiate function via argument fmuGUID. It must be identical to the one stored
inside the fmi2Instantiate function; otherwise the C code and the XML file of the FMU
are not consistent with each other. This argument cannot be null.

Argument fmuResourceLocation is a URI according to the IETF RFC3986 syntax to
indicate the location to the "resources" directory of the unzipped FMU archive.
[Function fmi2Instantiate is then able to read all
needed resources from this directory, for example maps or tables used by the FMU.]

Argument `functions` provides callback functions to be used from the FMU functions to
utilize resources from the environment. Only logging is implemented here.
Memory management callbacks will be removed in FMI v3.0.

Argument visible = fmi2False defines that the interaction with the user should be
reduced to a minimum (no application window, no plotting, no animation, etc.). In other
words, the FMU is executed in batch mode. If visible = fmi2True , the FMU is executed
in interactive mode, and the FMU might require to explicitly acknowledge start of simulation /
instantiation / initialization (acknowledgment is non-blocking).
`visible` is ignored by this implementation.

If loggingOn = fmi2True , debug logging is enabled. If loggingOn = fmi2False , debug
logging is disabled. [The FMU enable/disables LogCategories which are useful for
debugging according to this argument. Which LogCategories the FMU sets is unspecified.]
*/
func Instantiate(instanceName string, fmuType FMUType, fmuGUID string,
	fmuResourceLocation string, loggingOn bool, logFn LoggerCallback) C.fmi2Component {
	id := FMUID(C.malloc(1))
	fmu := &FMU{
		Name:             instanceName,
		Typee:            fmuType,
		GUID:             fmuGUID,
		ResourceLocation: fmuResourceLocation,
		State:            ModelStateInstantiated,
	}
	// log errors by default
	loggingMask := loggerCategoryError
	// loggingOn means log events
	if loggingOn {
		loggingMask |= loggerCategoryEvents
	}
	fmu.logger = &logger{
		mask:              loggingMask,
		fmiCallbackLogger: logFn,
	}

	if fmu.Name == "" {
		fmu.logger.Error(errors.New("Missing instance name"))
		return nil
	}

	if fmu.GUID == "" {
		fmu.logger.Error(errors.New("Missing GUID"))
		return nil
	}

	model, ok := models[fmu.GUID]
	if !ok {
		fmu.logger.Error(fmt.Errorf("GUID %s does not match any registered model", fmu.GUID))
		return nil
	}

	instance, err := model.Instantiate(fmu.logger)
	if err != nil {
		fmu.logger.Error(fmt.Errorf("Error instantiating model: %w", err))
		return nil
	}
	fmu.instance = instance

	fmus[id] = fmu

	return C.fmi2Component(id)
}

//export fmi2FreeInstance
/*
fmi2FreeInstance disposes the given instance, unloads the loaded model, and frees all the allocated memory
and other resources that have been allocated by the functions of the FMU interface. If a null
pointer is provided for `c`, the function call is ignored (does not have an effect).
*/
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
func fmi2SetDebugLogging(c C.fmi2Component, loggingOn C.fmi2Boolean,
	nCategories C.size_t, categories C.strings_t) C.fmi2Status {
	var cs []C.fmi2String
	carrayToSlice(unsafe.Pointer(categories), unsafe.Pointer(&cs), int(nCategories))
	cats := make([]string, len(cs))
	for i, c := range cs {
		cats[i] = C.GoString(c)
	}

	return C.fmi2Status(SetDebugLogging(FMUID(c), fmuBool(loggingOn), cats))
}

/*
SetDebugLogging controls debug logging that is output via the logger function callback.
If loggingOn = fmi2True, debug logging is enabled, otherwise it is switched off.
If loggingOn = fmi2True and nCategories = 0, then all debug messages shall be output.
If loggingOn=fmi2True and nCategories > 0, then only debug messages according to
the categories argument shall be output. Vector categories has
nCategories elements. The allowed values of categories are defined by the modeling
environment that generated the FMU. Depending on the generating modeling environment,
none, some or all allowed values for categories for this FMU are defined in the
modelDescription.xml file via element `fmiModelDescription.LogCategories `.
Supported log categories are in `logger.go`.
*/
func SetDebugLogging(id FMUID, loggingOn bool, categories []string) Status {
	const expected = ModelStateInstantiated | ModelStateInitializationMode |
		ModelStateEventMode | ModelStateContinuousTimeMode |
		ModelStateStepComplete | ModelStateStepInProgress | ModelStateStepFailed | ModelStateStepCanceled |
		ModelStateTerminated | ModelStateError
	fmu, ok := allowedState(id, "SetDebugLogging", expected)
	if !ok {
		return StatusError
	}
	if !loggingOn {
		fmu.logger.setMask(loggerCategoryNone)
		return StatusOK
	}
	if len(categories) == 0 {
		fmu.logger.setMask(loggerCategoryAll)
		return StatusOK
	}

	mask := loggerCategoryNone
	for _, cat := range categories {
		m, err := loggerCategoryFromString(cat)
		if err != nil {
			fmu.logger.Error(fmt.Errorf("Log category %s was not recognized", cat))
			return StatusError
		}
		mask |= m
	}
	fmu.logger.setMask(mask)
	return StatusOK
}

//export fmi2SetupExperiment
func fmi2SetupExperiment(c C.fmi2Component, toleranceDefined C.fmi2Boolean,
	tolerance C.fmi2Real, startTime C.fmi2Real, stopTimeDefined C.fmi2Boolean,
	stopTime C.fmi2Real) C.fmi2Status {

	return C.fmi2Status(SetupExperiment(FMUID(c),
		fmuBool(toleranceDefined), float64(tolerance),
		float64(startTime), fmuBool(stopTimeDefined), float64(stopTime)))
}

/*
SetupExperiment informs the FMU to setup the experiment. This function must be called after
fmi2Instantiate and before fmi2EnterInitializationMode is called. Arguments
toleranceDefined and tolerance depend on the FMU type:

fmuType = fmi2ModelExchange:
If `toleranceDefined = fmi2True`, then the model is called with a numerical
integration scheme where the step size is controlled by using `tolerance` for error
estimation (usually as relative tolerance). In such a case, all numerical algorithms used
inside the model (for example, to solve non-linear algebraic equations) should also
operate with an error estimation of an appropriate smaller relative tolerance.

fmuType = fmi2CoSimulation:
If `toleranceDefined = fmi2True`, then the communication interval of the slave is
controlled by error estimation. In case the slave utilizes a numerical integrator with
variable step size and error estimation, it is suggested to use `tolerance` for the error
estimation of the internal integrator (usually as relative tolerance).
An FMU for Co-Simulation might ignore this argument.

The arguments startTime and stopTime can be used to check whether the model is valid
within the given boundaries or to allocate memory which is necessary for storing results.
Argument startTime is the fixed initial value of the independent variable 5 [if the
independent variable is `time`, startTime is the starting time of initializaton]. If
`stopTimeDefined = fmi2True`, then `stopTime` is the defined final value of the
independent variable [if the independent variable is `time`, stopTime is the stop time of
the simulation] and if the environment tries to compute past stopTime the FMU has to
return `fmi2Status = fmi2Error`. If `stopTimeDefined = fmi2False`, then no final value
of the independent variable is defined and argument stopTime is meaningless.
*/
func SetupExperiment(id FMUID, toleranceDefined bool, tolerance float64,
	startTime float64, stopTimeDefined bool, stopTime float64) Status {
	const expected = ModelStateInstantiated
	fmu, ok := allowedState(id, "SetupExperiment", expected)
	if !ok {
		return StatusError
	}

	if err := fmu.instance.SetupExperiment(
		toleranceDefined, tolerance, startTime, stopTimeDefined, stopTime); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling SetupExperiment: %w", err))
		return StatusError
	}
	return StatusOK
}

//export fmi2EnterInitializationMode
func fmi2EnterInitializationMode(c C.fmi2Component) C.fmi2Status {
	return C.fmi2Status(EnterInitializationMode(FMUID(c)))
}

/*
EnterInitializationMode informs the FMU to enter Initialization Mode. Before calling this function, all variables with
attribute < ScalarVariable initial = "exact" or "approx"> can be set with the
`fmi2SetXXX` functions (the ScalarVariable attributes are defined in the Model Description File).
Setting other variables is not allowed. Furthermore, fmi2SetupExperiment must be called at least once before calling
fmi2EnterInitializationMode, in order that startTime is defined.
*/
func EnterInitializationMode(id FMUID) Status {
	const expected = ModelStateInstantiated
	fmu, ok := allowedState(id, "EnterInitializationMode", expected)
	if !ok {
		return StatusError
	}

	if err := fmu.instance.EnterInitializationMode(); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling EnterInitializationMode: %w", err))
		return StatusError
	}
	fmu.State = ModelStateInitializationMode

	return StatusOK
}

//export fmi2ExitInitializationMode
func fmi2ExitInitializationMode(c C.fmi2Component) C.fmi2Status {
	return C.fmi2Status(ExitInitializationMode(FMUID(c)))
}

/*
ExitInitializationMode informs the FMU to exit Initialization Mode.
For fmuType = fmi2ModelExchange , this function switches off all initialization equations,
and the FMU enters Event Mode implicitly; that is, all continuous-time and active discrete-
time equations are available.
*/
func ExitInitializationMode(id FMUID) Status {
	const expected = ModelStateInitializationMode
	fmu, ok := allowedState(id, "ExitInitializationMode", expected)
	if !ok {
		return StatusError
	}

	if err := fmu.instance.ExitInitializationMode(); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling ExitInitializationMode: %w", err))
		return StatusError
	}

	if fmu.Typee == FMUTypeModelExchange {
		fmu.State = ModelStateEventMode
	} else {
		fmu.State = ModelStateStepComplete
	}

	return StatusOK
}

//export fmi2Terminate
func fmi2Terminate(c C.fmi2Component) C.fmi2Status {
	return C.fmi2Status(Terminate(FMUID(c)))
}

/*
Terminate informs the FMU that the simulation run is terminated. After calling this function, the final
values of all variables can be inquired with the fmi2GetXXX(..) functions. It is not allowed
to call this function after one of the functions returned with a status flag of fmi2Error or
fmi2Fatal .
*/
func Terminate(id FMUID) Status {
	const expected = ModelStateEventMode | ModelStateContinuousTimeMode |
		ModelStateStepComplete | ModelStateStepFailed
	fmu, ok := allowedState(id, "Terminate", expected)
	if !ok {
		return StatusError
	}

	if err := fmu.instance.Terminate(); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling Terminate: %w", err))
		return StatusError
	}

	fmu.State = ModelStateTerminated
	return StatusOK
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

func allowedState(id FMUID, name string, expected ModelState) (*FMU, bool) {
	fmu, err := GetFMU(id)
	if err != nil {
		return nil, false
	}

	if fmu.State&expected == 0 {
		fmu.logger.Error(fmt.Errorf("Illegal call sequence at %s", name))
		return nil, false
	}
	return fmu, true
}

func carrayToSlice(carray unsafe.Pointer, slice unsafe.Pointer, len int) {
	sliceHeader := (*reflect.SliceHeader)(slice)
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(carray)
}
