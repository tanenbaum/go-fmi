package fmi_test

import (
	"reflect"
	"testing"

	"gitlab.com/edgetic/simulation/go-fmi/pkg/fmi"
)

func TestGetReal(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
		wantRs    []float64
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
			nil,
		},
		{
			"GetReal error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusError,
			fmi.ModelStateContinuousTimeMode,
			nil,
		},
		{
			"Empty value reference returns no results",
			args{
				id: instantiateDefault(fmi.ModelStateStepComplete),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]float64{},
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]float64{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, got := fmi.GetReal(tt.args.id, tt.args.vr)
			if got != tt.want {
				t.Errorf("GetReal() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(rs, tt.wantRs) {
				t.Errorf("Want values %v, got %v", tt.wantRs, rs)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetInteger(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
		wantIs    []int32
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
			nil,
		},
		{
			"GetInteger error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusError,
			fmi.ModelStateContinuousTimeMode,
			nil,
		},
		{
			"Empty value reference returns no results",
			args{
				id: instantiateDefault(fmi.ModelStateStepComplete),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]int32{},
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]int32{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, got := fmi.GetInteger(tt.args.id, tt.args.vr)
			if got != tt.want {
				t.Errorf("GetInteger() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(is, tt.wantIs) {
				t.Errorf("Want values %v, got %v", tt.wantIs, is)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetBoolean(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
		wantBs    []bool
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
			nil,
		},
		{
			"GetBoolean error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusError,
			fmi.ModelStateContinuousTimeMode,
			nil,
		},
		{
			"Empty value reference returns no results",
			args{
				id: instantiateDefault(fmi.ModelStateStepComplete),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]bool{},
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]bool{false, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, got := fmi.GetBoolean(tt.args.id, tt.args.vr)
			if got != tt.want {
				t.Errorf("GetBoolean() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(bs, tt.wantBs) {
				t.Errorf("Want values %v, got %v", tt.wantBs, bs)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetString(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
		wantSs    []string
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
			nil,
		},
		{
			"GetString error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusError,
			fmi.ModelStateContinuousTimeMode,
			nil,
		},
		{
			"Empty value reference returns no results",
			args{
				id: instantiateDefault(fmi.ModelStateStepComplete),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]string{},
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]string{"0", "1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, got := fmi.GetString(tt.args.id, tt.args.vr)
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(ss, tt.wantSs) {
				t.Errorf("Want values %v, got %v", tt.wantSs, ss)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestSetReal(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
		fs []float64
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(fmi.ModelStateError),
			},
			fmi.StatusError,
			fmi.ModelStateError,
		},
		{
			"SetReal error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateEventMode),
			},
			fmi.StatusError,
			fmi.ModelStateEventMode,
		},
		{
			"SetReal called without error",
			args{
				instantiateDefault(fmi.ModelStateEventMode),
				fmi.ValueReference{0, 1},
				[]float64{1.2, 1.3},
			},
			fmi.StatusOK,
			fmi.ModelStateEventMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetReal(tt.args.id, tt.args.vr, tt.args.fs); got != tt.want {
				t.Errorf("SetReal() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestSetInteger(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
		is []int32
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(fmi.ModelStateError),
			},
			fmi.StatusError,
			fmi.ModelStateError,
		},
		{
			"SetInteger error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateEventMode),
			},
			fmi.StatusError,
			fmi.ModelStateEventMode,
		},
		{
			"SetInteger called without error",
			args{
				instantiateDefault(fmi.ModelStateEventMode),
				fmi.ValueReference{0, 1},
				[]int32{0, 1},
			},
			fmi.StatusOK,
			fmi.ModelStateEventMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetInteger(tt.args.id, tt.args.vr, tt.args.is); got != tt.want {
				t.Errorf("SetInteger() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestSetBoolean(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
		bs []bool
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(fmi.ModelStateError),
			},
			fmi.StatusError,
			fmi.ModelStateError,
		},
		{
			"SetBoolean error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateEventMode),
			},
			fmi.StatusError,
			fmi.ModelStateEventMode,
		},
		{
			"SetBoolean called without error",
			args{
				instantiateDefault(fmi.ModelStateEventMode),
				fmi.ValueReference{0, 1},
				[]bool{true, false},
			},
			fmi.StatusOK,
			fmi.ModelStateEventMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetBoolean(tt.args.id, tt.args.vr, tt.args.bs); got != tt.want {
				t.Errorf("SetBoolean() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestSetString(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
		ss []string
	}
	tests := []struct {
		name      string
		args      args
		want      fmi.Status
		wantState fmi.ModelState
	}{
		{
			"FMU state is invalid",
			args{
				id: instantiateDefault(fmi.ModelStateError),
			},
			fmi.StatusError,
			fmi.ModelStateError,
		},
		{
			"SetString error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateEventMode),
			},
			fmi.StatusError,
			fmi.ModelStateEventMode,
		},
		{
			"SetString called without error",
			args{
				instantiateDefault(fmi.ModelStateEventMode),
				fmi.ValueReference{0, 1},
				[]string{"a", "b"},
			},
			fmi.StatusOK,
			fmi.ModelStateEventMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetString(tt.args.id, tt.args.vr, tt.args.ss); got != tt.want {
				t.Errorf("SetString() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}
