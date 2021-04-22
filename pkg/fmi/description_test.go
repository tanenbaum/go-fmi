package fmi

import (
	"reflect"
	"testing"
	"time"
)

func TestModelDescription_MarshallIndent(t *testing.T) {
	type fields struct {
		modelDescriptionStatic  modelDescriptionStatic
		Name                    string
		GUID                    string
		Description             string
		Author                  string
		ModelVersion            string
		Copyright               string
		License                 string
		GenerationTool          string
		GenerationDateAndTime   *time.Time
		NumberOfEventIndicators uint
		DefaultExperiment       *Experiment
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			"Model description is generated with all fields set",
			fields{
				modelDescriptionStatic: modelDescriptionStatic{
					FMIVersion: "2.0",
				},
				Name:                    "name",
				GUID:                    "guid-guid",
				Description:             "Thing here",
				Author:                  "Bob Smith",
				ModelVersion:            "v0.0.1",
				Copyright:               "Blah",
				License:                 "MIT",
				GenerationTool:          "Golang",
				GenerationDateAndTime:   &time.Time{},
				NumberOfEventIndicators: 2,
				DefaultExperiment: &Experiment{
					StartTime: 1,
					StopTime:  2,
					Tolerance: 0.1,
					StepSize:  1e-3,
				},
			},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<fmiModelDescription fmiVersion="2.0" modelName="name" guid="guid-guid" description="Thing here" author="Bob Smith" version="v0.0.1" copyright="Blah" license="MIT" generationTool="Golang" generationDateAndTime="0001-01-01T00:00:00Z" numberOfEventIndicators="2">
    <DefaultExperiment startTime="1" stopTime="2" tolerance="0.1" stepSize="0.001"></DefaultExperiment>
</fmiModelDescription>`),
			false,
		},
		{
			"Model description is generated with optional fields omitted",
			fields{
				modelDescriptionStatic: modelDescriptionStatic{
					FMIVersion: "2.0",
				},
				Name: "name",
				GUID: "guid-guid",
			},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<fmiModelDescription fmiVersion="2.0" modelName="name" guid="guid-guid" numberOfEventIndicators="0"></fmiModelDescription>`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ModelDescription{
				modelDescriptionStatic:  tt.fields.modelDescriptionStatic,
				Name:                    tt.fields.Name,
				GUID:                    tt.fields.GUID,
				Description:             tt.fields.Description,
				Author:                  tt.fields.Author,
				ModelVersion:            tt.fields.ModelVersion,
				Copyright:               tt.fields.Copyright,
				License:                 tt.fields.License,
				GenerationTool:          tt.fields.GenerationTool,
				GenerationDateAndTime:   tt.fields.GenerationDateAndTime,
				NumberOfEventIndicators: tt.fields.NumberOfEventIndicators,
				DefaultExperiment:       tt.fields.DefaultExperiment,
			}
			got, err := m.MarshallIndent()
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelDescription.MarshallIndent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelDescription.MarshallIndent() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
