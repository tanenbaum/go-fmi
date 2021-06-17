package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"

	"github.com/tanenbaum/go-fmi/pkg/fmi"
)

const (
	guid            = "{2d5ad039-5b33-4b1a-9405-e2455d930aed}"
	name            = "BouncingBall"
	fixedSolverStep = 1e-3
	v_min           = 0.1
)

const (
	vr_h = iota + 1
	vr_der_h
	vr_v
	vr_der_v
	vr_g
	vr_e
	vr_v_min
)

func init() {
	fmi.RegisterModel(model{})
}

type model struct{}

func (m model) Description() fmi.ModelDescription {
	return fmi.ModelDescription{
		GUID: guid,
		Name: name,
	}
}

func (m model) Instantiate(l fmi.Logger) (fmi.ModelInstance, error) {
	return &bouncingBall{
		Logger: l,
		data:   initialState(),
		z:      make([]float64, 1),
		prez:   make([]float64, 1),
	}, nil
}

type bouncingBall struct {
	fmi.Logger
	*data
	terminateSimulation  bool
	nextEventTimeDefined bool
	nextEventTime        float64
	nSteps               uint
	z                    []float64
	prez                 []float64
}

type data struct {
	H float64
	V float64
	G float64
	E float64
}

func initialState() *data {
	return &data{
		1,
		0,
		-9.81,
		0.7,
	}
}

func (b bouncingBall) SetupExperiment(toleranceDefined bool, tolerance float64,
	startTime float64, stopTimeDefined bool, stopTime float64) error {
	return nil
}

func (b bouncingBall) EnterInitializationMode() error {
	return nil
}

func (b bouncingBall) ExitInitializationMode() error {
	return nil
}

func (b bouncingBall) Terminate() error {
	return nil
}

func (b *bouncingBall) Reset() error {
	b.data = initialState()
	return nil
}

func (b *bouncingBall) Encode() ([]byte, error) {
	bs := &bytes.Buffer{}
	enc := gob.NewEncoder(bs)
	if err := enc.Encode(b.data); err != nil {
		return nil, err
	}
	return bs.Bytes(), nil
}

func (b *bouncingBall) Decode(rs []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(rs))
	d := &data{}
	if err := dec.Decode(d); err != nil {
		return err
	}
	b.data = d
	return nil
}

func (b *bouncingBall) DoStep(
	currentCommunicationPoint, communicationStepSize float64,
	noSetFMUStatePriorToCurrentPoint bool) (fmi.StepResult, error) {

	time := math.Abs(currentCommunicationPoint)
	tNext := currentCommunicationPoint + communicationStepSize

	epsilon := (1 + time) * math.Nextafter(0, 1)

	for time+fixedSolverStep < tNext+epsilon {

		x := b.getContinuousStates()
		dx := b.getDerivatives()

		// forward Euler step
		for i, d := range dx {
			x[i] += fixedSolverStep * d
		}

		b.setContinuousStates(x...)

		stateEvent := false

		b.z = b.getEventIndicators()

		// check for zero-crossings
		for i := range b.z {
			stateEvent = stateEvent || b.prez[i] < 0 && b.z[i] >= 0
			stateEvent = stateEvent || b.prez[i] > 0 && b.z[i] <= 0
		}

		// remember the current event indicators
		temp := b.z
		b.z = b.prez
		b.prez = temp

		// check for time event
		timeEvent := b.nextEventTimeDefined && (time+fixedSolverStep*1e-2) >= b.nextEventTime

		// log events
		if timeEvent {
			b.Event(fmt.Sprintf("Time event detected at t=%f s.", time))
		}
		if stateEvent {
			b.Event(fmt.Sprintf("State event detected at t=%f s.", time))
		}

		if stateEvent || timeEvent {
			b.eventUpdate()

			// update previous event indicators
			b.prez = b.getEventIndicators()
		}

		// terminate simulation, if requested by the model in the previous step
		if b.terminateSimulation {
			return fmi.StepResultPartial, nil
		}

		b.nSteps++
		time = fixedSolverStep * float64(b.nSteps)
	}

	return fmi.StepResultSuccess, nil
}

func (d *data) GetReal(vrs fmi.ValueReference) ([]float64, error) {
	fs := make([]float64, len(vrs))
	for i, vr := range vrs {
		switch vr {
		case vr_h:
			fs[i] = d.H
		case vr_der_h:
		case vr_v:
			fs[i] = d.V
		case vr_der_v:
		case vr_g:
			fs[i] = d.G
		case vr_e:
			fs[i] = d.E
		case vr_v_min:
			fs[i] = v_min
		default:
			return nil, fmt.Errorf("Value reference %d not recognised", vr)
		}
	}
	return fs, nil
}

func (d *data) GetInteger(fmi.ValueReference) ([]int32, error) {
	return nil, errors.New("GetInteger not allowed")
}

func (d *data) GetBoolean(fmi.ValueReference) ([]bool, error) {
	return nil, errors.New("GetBoolean not allowed")
}

func (d *data) GetString(fmi.ValueReference) ([]string, error) {
	return nil, errors.New("GetString not allowed")
}

func (d *data) SetReal(vrs fmi.ValueReference, fs []float64) error {
	for i, vr := range vrs {
		switch vr {
		case vr_h:
			d.H = fs[i]
		case vr_v:
			d.V = fs[i]
		case vr_g:
			d.G = fs[i]
		case vr_e:
			d.E = fs[i]
		case vr_v_min:
			return errors.New("Variable v_min is constant and cannot be set.")
		default:
			return fmt.Errorf("Unexpected value reference: %d", vr)
		}
	}
	return nil
}

func (d *data) SetInteger(fmi.ValueReference, []int32) error {
	return errors.New("SetInteger not allowed")
}

func (d *data) SetBoolean(fmi.ValueReference, []bool) error {
	return errors.New("SetBoolean not allowed")
}

func (d *data) SetString(fmi.ValueReference, []string) error {
	return errors.New("SetString not allowed")
}

func (d *data) eventUpdate() {
	if d.H <= 0 && d.V < 0 {
		d.H = 0
		d.V = -d.V * d.E

		if d.V < v_min {
			// stop bouncing
			d.V = 0
			d.G = 0
		}
	}
}

func (d *data) setContinuousStates(x ...float64) {
	d.H = x[0]
	d.V = x[1]
}

func (d *data) getContinuousStates() []float64 {
	return []float64{d.H, d.V}
}

func (d *data) getDerivatives() []float64 {
	return []float64{d.V, d.G}
}

func (d *data) getEventIndicators() []float64 {
	if d.H == 0 && d.V == 0 {
		return []float64{1}
	} else {
		return []float64{d.H}
	}
}

func main() {
}
