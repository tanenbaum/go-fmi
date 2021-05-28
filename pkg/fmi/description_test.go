package fmi

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewModelDescription(t *testing.T) {
	md := NewModelDescription()
	assert.Equal(t, ModelDescription{
		modelDescriptionStatic: modelDescriptionStatic{
			FMIVersion:               "2.0",
			VariableNamingConvention: "flat",
			LogCategories: []logCategory{
				{
					Name: "logEvents",
				},
				{
					Name: "logStatusWarning",
				},
				{
					Name: "logStatusDiscard",
				},
				{
					Name: "logStatusError",
				},
				{
					Name: "logStatusFatal",
				},
				{
					Name: "logStatusPending",
				},
				{
					Name: "logAll",
				},
			},
		},
	}, md)
}

func TestModelDescription_MarshallIndent(t *testing.T) {
	one := 1
	minusOne := -1
	factor := 57.29577951308232
	offset := 0.0
	i32 := int32(2)
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
					FMIVersion:               "2.0",
					VariableNamingConvention: "flat",
					LogCategories:            buildLogCategories(),
				},
				Name:                    "name",
				GUID:                    "guid-guid",
				Description:             "Thing here",
				Author:                  "Bob Smith",
				Version:                 "v0.0.1",
				Copyright:               "Blah",
				License:                 "MIT",
				GenerationTool:          "Golang",
				GenerationDateAndTime:   &time.Time{},
				NumberOfEventIndicators: 2,
				UnitDefinitions: []Unit{
					{
						Name: "rads/s",
						BaseUnit: &BaseUnit{
							S:   &minusOne,
							Rad: &one,
						},
						DisplayUnits: []DisplayUnit{
							{
								Name:   "deg/s",
								Factor: &factor,
							},
						},
					},
					{
						Name: "bar",
						BaseUnit: &BaseUnit{
							KG:     &one,
							M:      &minusOne,
							S:      &one,
							Factor: &factor,
							Offset: &offset,
						},
					},
				},
				TypeDefinitions: []SimpleType{
					{
						Name:        "realtype",
						Description: "real type desc",
						Real: &RealType{
							TypeDefinition: TypeDefinition{
								Quantity: "thing",
							},
							Unit:             "kg",
							DisplayUnit:      "bar",
							RelativeQuantity: true,
							Min:              &offset,
							Max:              &factor,
							Nominal:          &factor,
							Unbounded:        true,
						},
					},
					{
						Name:        "integertype",
						Description: "integer type desc",
						Integer: &IntegerType{
							TypeDefinition: TypeDefinition{
								Quantity: "thing2",
							},
							Min: &i32,
							Max: &i32,
						},
					},
					{
						Name:        "booleantype",
						Description: "boolean type desc",
						Boolean: &BooleanType{
							TypeDefinition: TypeDefinition{
								Quantity: "thing3",
							},
						},
					},
					{
						Name:        "stringtype",
						Description: "string type desc",
						String: &StringType{
							TypeDefinition: TypeDefinition{
								Quantity: "thing4",
							},
						},
					},
					{
						Name:        "enumtype",
						Description: "enum type desc",
						Enumeration: &EnumerationType{
							TypeDefinition: TypeDefinition{
								Quantity: "thing5",
							},
							Item: []EnumerationItem{
								{
									Name:        "enum1",
									Value:       45,
									Description: "enum one",
								},
								{
									Name:        "enum2",
									Value:       -1,
									Description: "enum two",
								},
							},
						},
					},
				},
				DefaultExperiment: func() *Experiment {
					start := 1.0
					stop := 2.0
					tolerance := 0.1
					StepSize := 1e-3
					return &Experiment{
						StartTime: &start,
						StopTime:  &stop,
						Tolerance: &tolerance,
						StepSize:  &StepSize,
					}
				}(),
				VendorAnnotations: []ToolAnnotation{
					{
						Name:     "Foo",
						InnerXML: "<Bar>Baz</Bar>",
					},
				},
				ModelVariables: []ScalarVariable{
					{
						Name:           "varreal",
						ValueReference: 1,
						Description:    "real desc",
						Causality:      VariableCausalityInput,
						Variability:    VariableVariabilityContinuous,
						Initial: func() *VariableInitial {
							v := VariableInitialApprox
							return &v
						}(),
						CanHandleMultipleSetPerTimeInstant: true,
						Annotations: []ToolAnnotation{
							{
								Name:     "Bar",
								InnerXML: "<A>1</A>",
							},
						},
						ScalarVariableType: &ScalarVariableType{
							Real: &RealVariable{
								DeclaredType: DeclaredType{
									DeclaredType: "realtype",
								},
								RealType: RealType{
									TypeDefinition: TypeDefinition{
										Quantity: "thing",
									},
									Unit:             "kg",
									DisplayUnit:      "disp",
									RelativeQuantity: true,
									Min:              &offset,
									Max:              &factor,
									Nominal:          &offset,
									Unbounded:        true,
								},
								Start:      &offset,
								Derivative: &factor,
								Reinit:     true,
							},
						},
					},
					{
						Name:           "varinteger",
						ValueReference: 2,
						Description:    "integer desc",
						Causality:      VariableCausalityIndependent,
						Variability:    VariableVariabilityDiscrete,
						Initial: func() *VariableInitial {
							v := VariableInitialExact
							return &v
						}(),
						ScalarVariableType: &ScalarVariableType{
							Integer: &IntegerVariable{
								IntegerType: IntegerType{
									Min: &i32,
									Max: &i32,
								},
								Start: &i32,
							},
						},
					},
					{
						Name:           "varboolean",
						ValueReference: 3,
						Description:    "boolean desc",
						ScalarVariableType: &ScalarVariableType{
							Boolean: &BooleanVariable{},
						},
					},
					{
						Name:           "varstring",
						ValueReference: 4,
						Description:    "string desc",
						ScalarVariableType: &ScalarVariableType{
							String: &StringVariable{
								Start: "foo",
							},
						},
					},
				},
				ModelStructure: ModelStructure{
					Outputs: []Unknown{
						{
							Index:        3,
							Dependencies: UintAttributeList{1, 2},
						},
					},
					Derivatives: []Unknown{
						{
							Index:        1,
							Dependencies: UintAttributeList{3},
						},
						{
							Index:            2,
							Dependencies:     UintAttributeList{3},
							DependenciesKind: StringAttributeList{"constant"},
						},
					},
					InitialUnknowns: []Unknown{
						{
							Index:            1,
							Dependencies:     UintAttributeList{2, 3, 4},
							DependenciesKind: StringAttributeList{"constant", "dependent", "fixed"},
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

func TestUintAttributeList_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		l        UintAttributeList
		wantText []byte
		wantErr  bool
	}{
		{
			"empty slice returns empty string",
			UintAttributeList{},
			[]byte(""),
			false,
		},
		{
			"uints slice returns numbers without brackets",
			UintAttributeList{1, 2, 3},
			[]byte("1 2 3"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, err := tt.l.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("UintAttributeList.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotText, tt.wantText) {
				t.Errorf("UintAttributeList.MarshalText() = %v, want %v", gotText, tt.wantText)
			}
		})
	}
}
