package fmi_test

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitlab.com/edgetic/simulation/go-fmi/pkg/fmi"
)

type mockModel struct {
	fmi.Model
	guid     string
	err      bool
	instance fmi.ModelInstance
}

type mockInstance struct {
	fmi.ModelInstance
	err bool
}

func (m mockModel) Description() fmi.ModelDescription {
	return fmi.ModelDescription{
		GUID: m.guid,
	}
}

func (m mockModel) Instantiate(l fmi.Logger) (fmi.ModelInstance, error) {
	if m.err {
		return nil, errors.New("Instantiate")
	}
	return m.instance, nil
}

func (m mockInstance) errOrNil(name string) error {
	if m.err {
		return errors.New(name)
	}
	return nil
}

func (m mockInstance) SetupExperiment(toleranceDefined bool, tolerance float64,
	startTime float64, stopTimeDefined bool, stopTime float64) error {
	return m.errOrNil("SetupExperiment")
}

func (m mockInstance) EnterInitializationMode() error {
	return m.errOrNil("EnterInitializationMode")
}

func (m mockInstance) ExitInitializationMode() error {
	return m.errOrNil("ExitInitializationMode")
}

func (m mockInstance) Terminate() error {
	return m.errOrNil("Terminate")
}

func (m mockInstance) Reset() error {
	return m.errOrNil("Reset")
}

func (m mockInstance) GetReal(vr fmi.ValueReference) ([]float64, error) {
	if m.err {
		return nil, errors.New("GetReal")
	}
	fs := make([]float64, len(vr))
	for i := range vr {
		fs[i] = float64(i)
	}
	return fs, nil
}

func (m mockInstance) GetInteger(vr fmi.ValueReference) ([]int32, error) {
	if m.err {
		return nil, errors.New("GetInteger")
	}
	fs := make([]int32, len(vr))
	for i := range vr {
		fs[i] = int32(i)
	}
	return fs, nil
}

func (m mockInstance) GetBoolean(vr fmi.ValueReference) ([]bool, error) {
	if m.err {
		return nil, errors.New("GetBoolean")
	}
	bs := make([]bool, len(vr))
	for i := range vr {
		if i%2 == 1 {
			bs[i] = true
		} else {
			bs[i] = false
		}
	}
	return bs, nil
}

func (m mockInstance) GetString(vr fmi.ValueReference) ([]string, error) {
	if m.err {
		return nil, errors.New("GetString")
	}
	ss := make([]string, len(vr))
	for i := range vr {
		ss[i] = strconv.Itoa(i)
	}
	return ss, nil
}

func (m mockInstance) SetReal(fmi.ValueReference, []float64) error {
	return m.errOrNil("SetReal")
}

func noopLogger(status fmi.Status, category, message string) {}

// model setup for testing
func init() {
	// default model
	fmi.RegisterModel(&mockModel{
		guid:     "GUID",
		instance: &mockInstance{},
	})
	// model methods return errors
	fmi.RegisterModel(&mockModel{
		guid: "ModelErrors",
		err:  true,
	})
	// model instances return errors
	fmi.RegisterModel(&mockModel{
		guid: "InstanceErrors",
		instance: &mockInstance{
			err: true,
		},
	})
}

func instantiateDefault(state ...fmi.ModelState) fmi.FMUID {
	id := fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger))
	instantiateState(id, state...)
	return id
}

func instantiateState(id fmi.FMUID, state ...fmi.ModelState) fmi.FMUID {
	if len(state) == 0 {
		return id
	}
	fmu, _ := fmi.GetFMU(id)
	var mask fmi.ModelState
	for _, s := range state {
		mask |= s
	}
	fmu.State = mask
	return id
}

func instantiateModelErrors(state ...fmi.ModelState) fmi.FMUID {
	id := fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "ModelErrors", "", false, noopLogger))
	instantiateState(id, state...)
	return id
}

func instantiateInstanceErrors(state ...fmi.ModelState) fmi.FMUID {
	id := fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "InstanceErrors", "", false, noopLogger))
	instantiateState(id, state...)
	return id
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"Is 2.0",
			"2.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetVersion(); got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypesPlatform(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"Is default",
			"default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetTypesPlatform(); got != tt.want {
				t.Errorf("GetTypesPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstantiate(t *testing.T) {
	loggerExpectError := func(status fmi.Status, category, message string) {
		if category != "logStatusError" {
			t.Errorf("Expected category error, got %v", category)
		}
	}
	loggerExpectNothing := func(status fmi.Status, category, message string) {
		t.Errorf("Logger shouldn't have been called with %v, %s, %s", status, category, message)
	}
	type args struct {
		instanceName        string
		fmuType             fmi.FMUType
		fmuGUID             string
		fmuResourceLocation string
		loggingOn           bool
		logger              fmi.LoggerCallback
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
		want    *fmi.FMU
	}{
		{
			"Instance name is required",
			args{
				"",
				fmi.FMUTypeCoSimulation,
				"GUID",
				"",
				true,
				loggerExpectError,
			},
			true,
			nil,
		},
		{
			"GUID is required",
			args{
				"Name",
				fmi.FMUTypeCoSimulation,
				"",
				"",
				true,
				loggerExpectError,
			},
			true,
			nil,
		},
		{
			"Instance must match registered model guid",
			args{
				"Name",
				fmi.FMUTypeCoSimulation,
				"MISSING",
				"",
				false,
				loggerExpectError,
			},
			true,
			nil,
		},
		{
			"Instantiate error is handled",
			args{
				"Name",
				fmi.FMUTypeCoSimulation,
				"ModelErrors",
				"",
				false,
				loggerExpectError,
			},
			true,
			nil,
		},
		{
			"Instance should be created and stored",
			args{
				"Name",
				fmi.FMUTypeCoSimulation,
				"GUID",
				"./path",
				false,
				loggerExpectNothing,
			},
			false,
			&fmi.FMU{
				Name:             "Name",
				Typee:            fmi.FMUTypeCoSimulation,
				GUID:             "GUID",
				ResourceLocation: "./path",
				State:            fmi.ModelStateInstantiated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := fmi.Instantiate(tt.args.instanceName, tt.args.fmuType, tt.args.fmuGUID, tt.args.fmuResourceLocation, tt.args.loggingOn, tt.args.logger)
			if id == nil {
				if !tt.wantNil {
					t.Errorf("Expected id %v, nil to be %v", id, tt.wantNil)
				}
				return
			}
			got, err := fmi.GetFMU(fmi.FMUID(id))
			if err != nil {
				t.Errorf("Instantiate() GetFMU error: %v", err)
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(fmi.FMU{})) {
				t.Errorf("Instantiate() = %v, want %v", got, tt.want)
			}
			fmi.FreeInstance(fmi.FMUID(id))
		})
	}
}

func TestGetFMU(t *testing.T) {
	type args struct {
		id fmi.FMUID
	}
	tests := []struct {
		name    string
		args    args
		want    *fmi.FMU
		wantErr bool
	}{
		{
			"id does not exist, returns error",
			args{
				fmi.FMUID(0),
			},
			nil,
			true,
		},
		{
			"id exists, return fmu",
			args{
				instantiateDefault(),
			},
			&fmi.FMU{
				Name:  "name",
				Typee: fmi.FMUTypeCoSimulation,
				GUID:  "GUID",
				State: fmi.ModelStateInstantiated,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fmi.GetFMU(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFMU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(fmi.FMU{})) {
				t.Errorf("GetFMU() = %v, want %v", got, tt.want)
			}
			fmi.FreeInstance(tt.args.id)
		})
	}
}

func TestFreeInstance(t *testing.T) {
	type args struct {
		id fmi.FMUID
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Nil instance is ignored",
			args{
				fmi.FMUID(unsafe.Pointer(nil)),
			},
		},
		{
			"Handles instance that doesn't exist",
			args{
				fmi.FMUID(0),
			},
		},
		{
			"Deletes existing FMU",
			args{
				instantiateDefault(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmi.FreeInstance(tt.args.id)
			if _, err := fmi.GetFMU(tt.args.id); err == nil {
				t.Errorf("Expected FMU to have been freed: %v", err)
			}
		})
	}
}

func TestRegisterModel(t *testing.T) {
	type args struct {
		model fmi.Model
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"GUID is required",
			args{
				mockModel{},
			},
			true,
		},
		{
			"GUID duplicate not allowed",
			args{
				mockModel{
					guid: "GUID",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fmi.RegisterModel(tt.args.model); (err != nil) != tt.wantErr {
				t.Errorf("RegisterModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDebugLogging(t *testing.T) {
	type args struct {
		id         fmi.FMUID
		loggingOn  bool
		categories []string
	}
	tests := []struct {
		name string
		args args
		want fmi.Status
	}{
		{
			"Unknown id returns error",
			args{
				id: 0,
			},
			fmi.StatusError,
		},
		{
			"Invalid state returns error",
			args{
				id: instantiateDefault(fmi.ModelStateStartAndEnd),
			},
			fmi.StatusError,
		},
		{
			"Logging can be set to off",
			args{
				id:        instantiateDefault(),
				loggingOn: false,
			},
			fmi.StatusOK,
		},
		{
			"Logging on with no categories",
			args{
				id:        instantiateDefault(),
				loggingOn: true,
			},
			fmi.StatusOK,
		},
		{
			"Invalid logger category returns error",
			args{
				id:         instantiateDefault(),
				loggingOn:  true,
				categories: []string{"foo"},
			},
			fmi.StatusError,
		},
		{
			"Categories are merged and set",
			args{
				id:         instantiateDefault(),
				loggingOn:  true,
				categories: []string{"logStatusDiscard", "logStatusPending"},
			},
			fmi.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetDebugLogging(tt.args.id, tt.args.loggingOn, tt.args.categories); got != tt.want {
				t.Errorf("SetDebugLogging() = %v, want %v", got, tt.want)
			}
			fmi.FreeInstance(tt.args.id)
		})
	}
}

func verifyFMUStateAndCleanUp(t *testing.T, id fmi.FMUID, state fmi.ModelState) {
	fmu, err := fmi.GetFMU(id)
	defer fmi.FreeInstance(id)
	if err != nil {
		t.Errorf("Error getting FMU: %w", err)
		return
	}

	if fmu.State != state {
		t.Errorf("Expected FMU state %v, got %v", fmu.State, state)
	}
}

func TestSetupExperiment(t *testing.T) {
	type args struct {
		id               fmi.FMUID
		toleranceDefined bool
		tolerance        float64
		startTime        float64
		stopTimeDefined  bool
		stopTime         float64
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
			"SetupExperiment error is returned",
			args{
				id: instantiateInstanceErrors(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
		},
		{
			"SetupExperiment is called",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusOK,
			fmi.ModelStateInstantiated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetupExperiment(tt.args.id, tt.args.toleranceDefined, tt.args.tolerance, tt.args.startTime, tt.args.stopTimeDefined, tt.args.stopTime); got != tt.want {
				t.Errorf("SetupExperiment() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestEnterInitializationMode(t *testing.T) {
	type args struct {
		id fmi.FMUID
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
			"EnterInitializationMode error is returned",
			args{
				id: instantiateInstanceErrors(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
		},
		{
			"EnterInitializationMode is called",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusOK,
			fmi.ModelStateInitializationMode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.EnterInitializationMode(tt.args.id); got != tt.want {
				t.Errorf("EnterInitializationMode() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestExitInitializationMode(t *testing.T) {
	type args struct {
		id fmi.FMUID
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
			"ExitInitializationMode error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateInitializationMode),
			},
			fmi.StatusError,
			fmi.ModelStateInitializationMode,
		},
		{
			"ExitInitializationMode is called",
			args{
				id: instantiateDefault(fmi.ModelStateInitializationMode),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.ExitInitializationMode(tt.args.id); got != tt.want {
				t.Errorf("ExitInitializationMode() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestTerminate(t *testing.T) {
	type args struct {
		id fmi.FMUID
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
				id: instantiateDefault(),
			},
			fmi.StatusError,
			fmi.ModelStateInstantiated,
		},
		{
			"Terminate error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusError,
			fmi.ModelStateContinuousTimeMode,
		},
		{
			"Terminate is called",
			args{
				id: instantiateDefault(fmi.ModelStateStepComplete),
			},
			fmi.StatusOK,
			fmi.ModelStateTerminated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.Terminate(tt.args.id); got != tt.want {
				t.Errorf("Terminate() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestReset(t *testing.T) {
	type args struct {
		id fmi.FMUID
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
				id: instantiateDefault(fmi.ModelStateFatal),
			},
			fmi.StatusError,
			fmi.ModelStateFatal,
		},
		{
			"Reset error is returned",
			args{
				id: instantiateInstanceErrors(fmi.ModelStateTerminated),
			},
			fmi.StatusError,
			fmi.ModelStateTerminated,
		},
		{
			"Reset is called",
			args{
				id: instantiateDefault(fmi.ModelStateContinuousTimeMode),
			},
			fmi.StatusOK,
			fmi.ModelStateInstantiated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.Reset(tt.args.id); got != tt.want {
				t.Errorf("Reset() = %v, want %v", got, tt.want)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetReal(t *testing.T) {
	type args struct {
		id fmi.FMUID
		vr fmi.ValueReference
		rs []float64
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
			nil,
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
				make([]float64, 2),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]float64{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetReal(tt.args.id, tt.args.vr, tt.args.rs); got != tt.want {
				t.Errorf("GetReal() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.args.rs, tt.wantRs) {
				t.Errorf("Want values %v, got %v", tt.wantRs, tt.args.rs)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetInteger(t *testing.T) {
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
			nil,
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
				make([]int32, 2),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]int32{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetInteger(tt.args.id, tt.args.vr, tt.args.is); got != tt.want {
				t.Errorf("GetInteger() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.args.is, tt.wantIs) {
				t.Errorf("Want values %v, got %v", tt.wantIs, tt.args.is)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetBoolean(t *testing.T) {
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
			nil,
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
				make([]bool, 2),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]bool{false, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetBoolean(tt.args.id, tt.args.vr, tt.args.bs); got != tt.want {
				t.Errorf("GetBoolean() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.args.bs, tt.wantBs) {
				t.Errorf("Want values %v, got %v", tt.wantBs, tt.args.bs)
			}
			verifyFMUStateAndCleanUp(t, tt.args.id, tt.wantState)
		})
	}
}

func TestGetString(t *testing.T) {
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
			nil,
		},
		{
			"Values slice is populated",
			args{
				instantiateDefault(fmi.ModelStateStepComplete),
				fmi.ValueReference{0, 1},
				make([]string, 2),
			},
			fmi.StatusOK,
			fmi.ModelStateStepComplete,
			[]string{"0", "1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.GetString(tt.args.id, tt.args.vr, tt.args.ss); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.args.ss, tt.wantSs) {
				t.Errorf("Want values %v, got %v", tt.wantSs, tt.args.ss)
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
				id: instantiateDefault(fmi.ModelStateEventMode),
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
