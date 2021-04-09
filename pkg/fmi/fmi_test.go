package fmi_test

import (
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitlab.com/edgetic/simulation/go-fmi/pkg/fmi"
)

type mockModel struct {
	fmi.Model
	guid string
}

func (m mockModel) Description() fmi.ModelDescription {
	return fmi.ModelDescription{
		GUID: m.guid,
	}
}

func noopLogger(status fmi.Status, category, message string) {}

// model setup for testing
func init() {
	fmi.RegisterModel(&mockModel{
		guid: "GUID",
	})
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
				fmi.FMUID(fmi.Instantiate("foo", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
			},
			&fmi.FMU{
				Name:  "foo",
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
				fmi.FMUID(fmi.Instantiate("foo", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
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
					id := fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger))
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
				id:        fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
				loggingOn: false,
			},
			fmi.StatusOK,
		},
		{
			"Logging on with no categories",
			args{
				id:        fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
				loggingOn: true,
			},
			fmi.StatusOK,
		},
		{
			"Invalid logger category returns error",
			args{
				id:         fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
				loggingOn:  true,
				categories: []string{"foo"},
			},
			fmi.StatusError,
		},
		{
			"Categories are merged and set",
			args{
				id:         fmi.FMUID(fmi.Instantiate("name", fmi.FMUTypeCoSimulation, "GUID", "", false, noopLogger)),
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
