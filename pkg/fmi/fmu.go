package fmi

import (
	"errors"
	"fmt"
)

const (
	FMUTypeModelExchange FMUType = iota
	FMUTypeCoSimulation
)

const (
	StatusOK Status = iota
	StatusWarning
	StatusDiscard
	StatusError
	StatusFatal
	StatusPending
)

const (
	// StepResultSuccess means DoStep worked fine
	StepResultSuccess StepResult = iota
	// StepResultPartial means only a partial result was computed
	StepResultPartial
	// StepResultAsync means the slave has returned and is running async
	StepResultAsync
)

const (
	ModelStateStartAndEnd ModelState = 1 << iota
	ModelStateInstantiated
	ModelStateInitializationMode
	// ME states
	ModelStateEventMode
	ModelStateContinuousTimeMode

	// CS states
	ModelStateStepComplete
	ModelStateStepInProgress
	ModelStateStepFailed
	ModelStateStepCanceled

	ModelStateTerminated
	ModelStateError
	ModelStateFatal
)

// FMUType is type of FMU
type FMUType uint

// ModelState represents state machine of model
type ModelState uint

// FMU represents an active FMU instance
type FMU struct {
	Name             string
	Typee            FMUType
	GUID             string
	ResourceLocation string
	State            ModelState

	logger    Logger
	instance  ModelInstance
	startTime float64
}

// Status is return status of functions
type Status uint

// ValueReference is list of indexes to model values
type ValueReference []uint

// Model represents an FMU model to be executed in model-exchange or co-simulation
type Model interface {
	// Description provides XML compatible model description
	// used to generated `modelDescription.xml` as well as set defaults for initialisation
	Description() ModelDescription

	// Instantiate returns a new model instance.
	// ModelInstance should be a new thread-safe instance.
	// Can return an error if the implementation needs to.
	Instantiate(Logger) (ModelInstance, error)
}

// ModelInstance represents a live FMU that is being simulated through FMI interface
type ModelInstance interface {
	// SetupExperiment called from fmi2SetupExperiment.
	// error can be returned if there are issues.
	SetupExperiment(toleranceDefined bool, tolerance float64,
		startTime float64, stopTimeDefined bool, stopTime float64) error

	// EnterInitializationMode called from fmi2EnterInitializationMode
	EnterInitializationMode() error

	// ExitInitializationMode called from fmi2ExitInitializationMode
	ExitInitializationMode() error

	// Terminate called from fmi2Terminate
	Terminate() error

	// Reset called from fmi2Reset
	Reset() error
}

// ValueGetter retrieves values from the model
type ValueGetter interface {
	// GetReal called from fmi2GetReal
	GetReal(ValueReference) ([]float64, error)

	// GetInteger called from fmi2GetInteger
	GetInteger(ValueReference) ([]int32, error)

	// GetBoolean called from fmi2GetBoolean
	GetBoolean(ValueReference) ([]bool, error)

	// GetString called from fmi2GetString
	GetString(ValueReference) ([]string, error)
}

// ValueSetter sets values in the model
type ValueSetter interface {
	// SetReal called from fmi2SetReal
	SetReal(ValueReference, []float64) error

	// SetInteger called from fmi2SetInteger
	SetInteger(ValueReference, []int32) error

	// SetBoolean called from fmi2SetBoolean
	SetBoolean(ValueReference, []bool) error

	// SetString called from fmi2SetString
	SetString(ValueReference, []string) error
}

// StepResult is returned from cosim DoStep
type StepResult uint

func (r StepResult) Status() Status {
	switch r {
	case StepResultSuccess:
		return StatusOK
	case StepResultPartial:
		return StatusDiscard
	case StepResultAsync:
		return StatusPending
	}
	return StatusWarning
}

// CoSimulator implements methods for co-simulation
type CoSimulator interface {
	ValueGetterSetter

	// DoStep is called by fmi2DoStep
	DoStep(
		currentCommunicationPoint, communicationStepSize float64,
		noSetFMUStatePriorToCurrentPoint bool) (StepResult, error)
}

// ModelExchanger implements methods for model exchange
type ModelExchanger interface {
	ValueGetterSetter
}

type ValueGetterSetter interface {
	ValueGetter
	ValueSetter
}

// StateEncoder to be implemented by models that support state serialization.
// Used by fmi2GetFMUstate, fmi2SerializedFMUstateSize and fmi2SerializeFMUstate.
type StateEncoder interface {
	// Encode internal state into byte slice or return error
	Encode() ([]byte, error)
}

// StateDecoder to be implemented by models that support state serialization
// Used by fmi2SetFMUstate and fmi2DeSerializeFMUstate.
type StateDecoder interface {
	// Decode state byte slice and replace internal state or return error
	Decode([]byte) error
}

func (f *FMU) ValueGetter() (ValueGetter, error) {
	return f.valueGetterSetter()
}

func (f *FMU) ValueSetter() (ValueSetter, error) {
	return f.valueGetterSetter()
}

// CoSimulator gets cosimulator instance for the FMU.
// Returns an error if the fmu type is not cosimulation and implements the interface.
func (f *FMU) CoSimulator() (CoSimulator, error) {
	if f.Typee != FMUTypeCoSimulation {
		return nil, errors.New("FMU type is not set to cosimulation")
	}

	return f.cosimulator()
}

func (f *FMU) cosimulator() (CoSimulator, error) {
	cosim, ok := f.instance.(CoSimulator)
	if !ok {
		return nil, errors.New("FMU model instance does not implement cosimulation interface")
	}
	return cosim, nil
}

func (f *FMU) modelExchanger() (ModelExchanger, error) {
	mexch, ok := f.instance.(ModelExchanger)
	if !ok {
		return nil, errors.New("FMU model instance does not implement model exchange interface")
	}
	return mexch, nil
}

func (f *FMU) valueGetterSetter() (ValueGetterSetter, error) {
	switch f.Typee {
	case FMUTypeCoSimulation:
		return f.cosimulator()
	case FMUTypeModelExchange:
		return f.modelExchanger()
	}
	return nil, fmt.Errorf("Unknown FMU type %v", f.Typee)
}

func (f *FMU) StateEncoder() (StateEncoder, error) {
	se, ok := f.instance.(StateEncoder)
	if !ok {
		return nil, errors.New("FMU model instance does not implement state encoding for getting FMU state")
	}
	return se, nil
}

func (f *FMU) StateDecoder() (StateDecoder, error) {
	sd, ok := f.instance.(StateDecoder)
	if !ok {
		return nil, errors.New("FMU model instance does not implement state decoding for setting FMU state")
	}
	return sd, nil
}
