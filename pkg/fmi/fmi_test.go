package fmi_test

import (
	"reflect"
	"testing"

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
				Visible:          true,
				LoggingOn:        false,
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
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("Instantiate() = %v, want %v", got, &tt.want)
			}
		})
	}
}
