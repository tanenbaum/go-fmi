package fmi

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

// Status is return status of functions
type Status uint

// ModelState represents state machine of model
type ModelState uint

// FMU represents an active FMU instance
type FMU struct {
	Name             string
	Typee            FMUType
	GUID             string
	ResourceLocation string
	State            ModelState

	logger   Logger
	instance ModelInstance
}

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
}
