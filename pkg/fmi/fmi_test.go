package fmi_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitlab.com/edgetic/simulation/go-fmi/pkg/fmi"
)

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
	type args struct {
		instanceName        string
		fmuType             fmi.FMUType
		fmuGUID             string
		fmuResourceLocation string
		visible             bool
		loggingOn           bool
	}
	tests := []struct {
		name string
		args args
		want fmi.FMU
	}{
		{
			"Instance should be created and stored",
			args{
				"Name",
				fmi.FMUTypeCoSimulation,
				"GUID",
				"./path",
				true,
				false,
			},
			fmi.FMU{
				Name:             "Name",
				Typee:            fmi.FMUTypeCoSimulation,
				Guid:             "GUID",
				ResourceLocation: "./path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := fmi.Instantiate(tt.args.instanceName, tt.args.fmuType, tt.args.fmuGUID, tt.args.fmuResourceLocation, tt.args.visible, tt.args.loggingOn)
			got, err := fmi.GetFMU(id)
			if err != nil {
				t.Errorf("Instantiate() GetFMU error: %v", err)
			}
			if !cmp.Equal(got, &tt.want, cmpopts.IgnoreUnexported(fmi.FMU{})) {
				t.Errorf("Instantiate() = %v, want %v", got, &tt.want)
			}
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
				fmi.Instantiate("foo", fmi.FMUTypeCoSimulation, "thing", "", false, false),
			},
			&fmi.FMU{
				Name:  "foo",
				Typee: fmi.FMUTypeCoSimulation,
				Guid:  "thing",
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
			"Handles instance that doesn't exist",
			args{
				fmi.FMUID(2),
			},
		},
		{
			"Deletes existing FMU",
			args{
				fmi.Instantiate("foo", fmi.FMUTypeCoSimulation, "", "", false, false),
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
