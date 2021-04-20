package fmi

// #include <stdlib.h>
// #include "./c/fmi2Functions.h"
// #include "bridge.h"
import "C"

import (
	"fmt"
	"unsafe"
)

//export fmi2GetReal
func fmi2GetReal(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Real) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}
	var rs []C.fmi2Real
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&rs), int(nvr))
	fs, s := GetReal(FMUID(c), vs)
	if s != StatusOK {
		return C.fmi2Status(s)
	}
	copyRealArray(fs, rs)
	return C.fmi2OK
}

// GetReal gets real values by value reference
func GetReal(id FMUID, vr ValueReference) ([]float64, Status) {
	fmu, ok := allowedGetValue(id, "GetReal")
	if !ok {
		return nil, StatusError
	}

	vg, err := fmu.ValueGetter()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	fs, err := vg.GetReal(vr)
	if err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling GetReal: %w", err))
		return nil, StatusError
	}
	return fs, StatusOK
}

//export fmi2GetInteger
func fmi2GetInteger(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Integer) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}
	var is []C.fmi2Integer
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&is), int(nvr))
	ints, s := GetInteger(FMUID(c), vs)
	if s != StatusOK {
		return C.fmi2Status(s)
	}
	copyIntegerArray(ints, is)
	return C.fmi2OK
}

// GetInteger gets integer values by value reference
func GetInteger(id FMUID, vr ValueReference) ([]int32, Status) {
	fmu, ok := allowedGetValue(id, "GetInteger")
	if !ok {
		return nil, StatusError
	}

	vg, err := fmu.ValueGetter()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	is, err := vg.GetInteger(vr)
	if err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling GetInteger: %w", err))
		return nil, StatusError
	}
	return is, StatusOK
}

//export fmi2GetBoolean
func fmi2GetBoolean(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2Boolean) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}
	var bs []C.fmi2Boolean
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&bs), int(nvr))
	bools, s := GetBoolean(FMUID(c), vs)
	if s != StatusOK {
		return C.fmi2Status(s)
	}
	copyBooleanArray(bools, bs)
	return C.fmi2OK
}

// GetBoolean gets boolean values by value reference
func GetBoolean(id FMUID, vr ValueReference) ([]bool, Status) {
	fmu, ok := allowedGetValue(id, "GetBoolean")
	if !ok {
		return nil, StatusError
	}

	vg, err := fmu.ValueGetter()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	bs, err := vg.GetBoolean(vr)
	if err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling GetBoolean: %w", err))
		return nil, StatusError
	}

	return bs, StatusOK
}

//export fmi2GetString
func fmi2GetString(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value *C.fmi2String) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}
	var ss []C.fmi2String
	carrayToSlice(unsafe.Pointer(value), unsafe.Pointer(&ss), int(nvr))
	strs, s := GetString(FMUID(c), vs)
	if s != StatusOK {
		return C.fmi2Status(s)
	}
	copyStringArray(strs, ss)
	return C.fmi2OK
}

// GetString gets string values by value reference
func GetString(id FMUID, vr ValueReference) ([]string, Status) {
	fmu, ok := allowedGetValue(id, "GetString")
	if !ok {
		return nil, StatusError
	}

	vg, err := fmu.ValueGetter()
	if err != nil {
		fmu.logger.Error(err)
		return nil, StatusError
	}

	ss, err := vg.GetString(vr)
	if err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling GetString: %w", err))
		return nil, StatusError
	}

	return ss, StatusOK
}

//export fmi2SetReal
func fmi2SetReal(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Reals_t) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}

	fs, err := fmi2Reals(value, nvr)
	if err != nil {
		return logError(c, err)
	}
	return C.fmi2Status(SetReal(FMUID(c), vs, fs))
}

// SetReal sets floats by value references
func SetReal(id FMUID, vr ValueReference, fs []float64) Status {
	fmu, ok := allowedSetValue(id, "SetReal")
	if !ok {
		return StatusError
	}

	vs, err := fmu.ValueSetter()
	if err != nil {
		fmu.logger.Error(err)
		return StatusError
	}

	if err := vs.SetReal(vr, fs); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling SetReal: %w", err))
		return StatusError
	}

	return StatusOK
}

//export fmi2SetInteger
func fmi2SetInteger(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Integers_t) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}

	is, err := fmi2Integers(value, nvr)
	if err != nil {
		return logError(c, err)
	}
	return C.fmi2Status(SetInteger(FMUID(c), vs, is))
}

// SetInteger sets ints by value references
func SetInteger(id FMUID, vr ValueReference, is []int32) Status {
	fmu, ok := allowedSetValue(id, "SetInteger")
	if !ok {
		return StatusError
	}

	vs, err := fmu.ValueSetter()
	if err != nil {
		fmu.logger.Error(err)
		return StatusError
	}

	if err := vs.SetInteger(vr, is); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling SetInteger: %w", err))
		return StatusError
	}

	return StatusOK
}

//export fmi2SetBoolean
func fmi2SetBoolean(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Booleans_t) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}

	bs, err := fmi2Booleans(value, nvr)
	if err != nil {
		return logError(c, err)
	}
	return C.fmi2Status(SetBoolean(FMUID(c), vs, bs))
}

// SetBoolean sets bools by value references
func SetBoolean(id FMUID, vr ValueReference, bs []bool) Status {
	fmu, ok := allowedSetValue(id, "SetBoolean")
	if !ok {
		return StatusError
	}

	vs, err := fmu.ValueSetter()
	if err != nil {
		fmu.logger.Error(err)
		return StatusError
	}

	if err := vs.SetBoolean(vr, bs); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling SetBoolean: %w", err))
		return StatusError
	}

	return StatusOK
}

//export fmi2SetString
func fmi2SetString(c C.fmi2Component, vr C.valueReferences_t, nvr C.size_t, value C.fmi2Strings_t) C.fmi2Status {
	vs, err := valueReferences(vr, nvr)
	if err != nil {
		return logError(c, err)
	}

	ss, err := fmi2Strings(value, nvr)
	if err != nil {
		return logError(c, err)
	}
	return C.fmi2Status(SetString(FMUID(c), vs, ss))
}

// SetString sets strings by value references
func SetString(id FMUID, vr ValueReference, ss []string) Status {
	fmu, ok := allowedSetValue(id, "SetString")
	if !ok {
		return StatusError
	}

	vs, err := fmu.ValueSetter()
	if err != nil {
		fmu.logger.Error(err)
		return StatusError
	}

	if err := vs.SetString(vr, ss); err != nil {
		fmu.logger.Error(fmt.Errorf("Error calling SetString: %w", err))
		return StatusError
	}

	return StatusOK
}

//export fmi2GetDirectionalDerivative
func fmi2GetDirectionalDerivative(c C.fmi2Component, vUnknown_ref C.valueReferences_t, nUnknown C.size_t,
	vKnown_ref C.valueReferences_t, nKnown C.size_t,
	dvKnown C.fmi2Reals_t, dvUnknown *C.fmi2Real) C.fmi2Status {
	// TODO: implement
	return C.fmi2OK
}

func valueReferences(vr C.valueReferences_t, nvr C.size_t) (ValueReference, error) {
	if nvr == 0 {
		return nil, nil
	}

	if nvr > 0 && vr == nil {
		return nil, fmt.Errorf("Value references array is null but size is %d", nvr)
	}

	var vs []C.fmi2ValueReference
	carrayToSlice(unsafe.Pointer(vr), unsafe.Pointer(&vs), int(nvr))
	vrs := make(ValueReference, nvr)
	for i := 0; i < int(nvr); i++ {
		vrs[i] = uint(vs[i])
	}
	return vrs, nil
}
