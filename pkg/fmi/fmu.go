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

// FMUType is type of FMU
type FMUType uint

// Status is return status of functions
type Status uint

type FMU struct {
	Name             string
	Typee            FMUType
	Guid             string
	ResourceLocation string

	logger Logger
}
