package fmi_test

import (
	"reflect"
	"testing"

	"gitlab.com/edgetic/simulation/go-fmi/pkg/fmi"
)

func TestGetFMUState(t *testing.T) {
	type args struct {
		id fmi.FMUID
	}
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 fmi.Status
	}{
		{
			"Model state is invalid",
			args{
				instantiateDefault(fmi.ModelStateStartAndEnd),
			},
			nil,
			fmi.StatusError,
		},
		{
			"State encoder error is handled",
			args{
				instantiateInstanceErrors(),
			},
			nil,
			fmi.StatusError,
		},
		{
			"State is encoded to bytes",
			args{
				instantiateDefault(),
			},
			[]byte("foo"),
			fmi.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := fmi.GetFMUState(tt.args.id)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFMUState() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetFMUState() got1 = %v, want %v", got1, tt.want1)
			}
			fmi.FreeInstance(fmi.FMUID(tt.args.id))
		})
	}
}

func TestSetFMUState(t *testing.T) {
	type args struct {
		id fmi.FMUID
		bs []byte
	}
	tests := []struct {
		name string
		args args
		want fmi.Status
	}{
		{
			"Model state is invalid",
			args{
				id: instantiateDefault(fmi.ModelStateStartAndEnd),
			},
			fmi.StatusError,
		},
		{
			"State decoder error is handled",
			args{
				id: instantiateInstanceErrors(),
			},
			fmi.StatusError,
		},
		{
			"Decode is called successfully",
			args{
				id: instantiateDefault(),
				bs: []byte("foo"),
			},
			fmi.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmi.SetFMUState(tt.args.id, tt.args.bs); got != tt.want {
				t.Errorf("SetFMUState() = %v, want %v", got, tt.want)
			}
		})
	}
}
