package fmi_test

import (
	"errors"
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

func (m mockInstance) SetupExperiment(toleranceDefined bool, tolerance float64,
	startTime float64, stopTimeDefined bool, stopTime float64) error {
	if m.err {
		return errors.New("SetupExperiment")
	}
	return nil
}

func (m mockInstance) EnterInitializationMode() error {
	if m.err {
		return errors.New("EnterInitializationMode")
	}
	return nil
}

func noopLogger(status fmi.Status, category, message string) {}

// model setup for testing
func init() {
	fmi.RegisterModel(&mockModel{
		guid:     "GUID",
		instance: &mockInstance{},
	})
	fmi.RegisterModel(&mockModel{
		guid: "ModelErrors",
		err:  true,
	})
	fmi.RegisterModel(&mockModel{
		guid: "InstanceErrors",
		instance: &mockInstance{
			err: true,
		},
	})
}

func instantiateDefault() fmi.FMUID {
	return fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger))
}

func instantiateModelErrors() fmi.FMUID {
	return fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "ModelErrors", "", false, noopLogger))
}

func instantiateInstanceErrors() fmi.FMUID {
	return fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "InstanceErrors", "", false, noopLogger))
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
				fmi.FMUID(2),
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
				fmi.FMUID(2),
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
				id: func() fmi.FMUID {
					id := instantiateDefault()
					fmu, _ := fmi.GetFMU(id)
					fmu.State = fmi.ModelStateStartAndEnd
					return id
				}(),
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
		name string
		args args
		want fmi.Status
	}{
		{
			"FMU state is invalid",
			args{
				id: func() fmi.FMUID {
					id := instantiateDefault()
					fmu, _ := fmi.GetFMU(id)
					fmu.State = fmi.ModelStateError
					return id
				}(),
			},
			fmi.StatusError,
		},
		{
			"SetupExperiment error is returned",
			args{
				id: instantiateInstanceErrors(),
			},
			fmi.StatusError,
		},
		{
			"SetupExperiment is called",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetupExperiment(tt.args.id, tt.args.toleranceDefined, tt.args.tolerance, tt.args.startTime, tt.args.stopTimeDefined, tt.args.stopTime); got != tt.want {
				t.Errorf("SetupExperiment() = %v, want %v", got, tt.want)
			}
			fmi.FreeInstance(tt.args.id)
		})
	}
}

func TestEnterInitializationMode(t *testing.T) {
	type args struct {
		id fmi.FMUID
	}
	tests := []struct {
		name string
		args args
		want fmi.Status
	}{
		{
			"FMU state is invalid",
			args{
				id: func() fmi.FMUID {
					id := instantiateDefault()
					fmu, _ := fmi.GetFMU(id)
					fmu.State = fmi.ModelStateError
					return id
				}(),
			},
			fmi.StatusError,
		},
		{
			"EnterInitializationMode error is returned",
			args{
				id: instantiateInstanceErrors(),
			},
			fmi.StatusError,
		},
		{
			"EnterInitializationMode is called",
			args{
				id: instantiateDefault(),
			},
			fmi.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.EnterInitializationMode(tt.args.id); got != tt.want {
				t.Errorf("EnterInitializationMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
