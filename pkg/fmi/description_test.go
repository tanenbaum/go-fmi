package fmi

import (
	"reflect"
	"testing"
	"time"
)

func TestModelDescription_MarshallIndent(t *testing.T) {
	tests := []struct {
		name             string
		modelDescription ModelDescription
		want             []byte
		wantErr          bool
	}{
		{
			"Model description is generated with all fields set",
			ModelDescription{
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
				VendorAnnotations: &struct {
					Tool []ToolAnnotation `xml:"Tool,omitempty"`
				}{
					[]ToolAnnotation{
						{
							Name:     "Foo",
							InnerXML: "<Bar>Baz</Bar>",
						},
					},
				},
			},
			[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<fmiModelDescription fmiVersion="2.0" modelName="name" guid="guid-guid" description="Thing here" author="Bob Smith" version="v0.0.1" copyright="Blah" license="MIT" generationTool="Golang" generationDateAndTime="0001-01-01T00:00:00Z" numberOfEventIndicators="2">
    <DefaultExperiment startTime="1" stopTime="2" tolerance="0.1" stepSize="0.001"></DefaultExperiment>
    <VendorAnnotations>
        <Tool name="Foo"><Bar>Baz</Bar></Tool>
    </VendorAnnotations>
</fmiModelDescription>`),
			false,
		},
		{
			"Model description is generated with optional fields omitted",
			ModelDescription{
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
			m := tt.modelDescription
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
