package fmi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModelVariables(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    ModelVariables
		wantErr bool
	}{
		{
			"Returns error if not a struct",
			args{
				42,
			},
			nil,
			true,
		},
		{
			"Struct must have at least one field",
			args{
				struct{}{},
			},
			nil,
			true,
		},
		{
			"Struct must have exported fields",
			args{
				struct{ a float64 }{
					41.2,
				},
			},
			nil,
			true,
		},
		{
			"Struct exported fields need no specific tags, names are inferred",
			args{
				struct {
					A float64
					B int32
					C bool
					D string
				}{
					41.2,
					12,
					true,
					"foo",
				},
			},
			&modelVariables{
				model: struct {
					A float64
					B int32
					C bool
					D string
				}{
					41.2,
					12,
					true,
					"foo",
				},
				scalars: []ScalarVariable{
					{
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeReal,
							Real:         &RealVariable{},
						},
						Name:           "A",
						ValueReference: 1,
					},
					{
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeInteger,
							Integer:      &IntegerVariable{},
						},
						Name:           "B",
						ValueReference: 2,
					},
					{
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeBoolean,
							Boolean:      &BooleanVariable{},
						},
						Name:           "C",
						ValueReference: 3,
					},
					{
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeString,
							String:       &StringVariable{},
						},
						Name:           "D",
						ValueReference: 4,
					},
				},
			},
			false,
		},
		{
			"struct with unsupported type returns error",
			args{
				struct{ A int64 }{
					42,
				},
			},
			nil,
			true,
		},
		{
			"struct with embedded types returns error",
			args{
				struct{ A struct{ B float64 } }{
					A: struct{ B float64 }{
						B: 42,
					},
				},
			},
			nil,
			true,
		},
		{
			"generic scalar variable tags are parsed and set",
			args{
				struct {
					A float64 `description:"foo" causality:"parameter" variability:"tunable" initial:"approx" canhandlemultiplesetpertimeinstant:"true"`
				}{
					42,
				},
			},
			&modelVariables{
				model: struct {
					A float64 `description:"foo" causality:"parameter" variability:"tunable" initial:"approx" canhandlemultiplesetpertimeinstant:"true"`
				}{
					42,
				},
				scalars: []ScalarVariable{
					{
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeReal,
							Real:         &RealVariable{},
						},
						Name:           "A",
						ValueReference: 1,
						Description:    "foo",
						Causality:      VariableCausalityParameter,
						Variability:    VariableVariabilityTunable,
						Initial: func() *VariableInitial {
							i := VariableInitialApprox
							return &i
						}(),
						CanHandleMultipleSetPerTimeInstant: true,
					},
				},
			},
			false,
		},
		{
			"generic scalar variable tag causality is validated",
			args{
				struct {
					A float64 `causality:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"generic scalar variable tag variability is validated",
			args{
				struct {
					A float64 `variability:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"generic scalar variable tag initial is validated",
			args{
				struct {
					A float64 `initial:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"all tags for real variable type can be set",
			args{
				struct {
					A float64 `declaredtype:"foo" start:"1" derivative:"0.2" reinit:"true" quantity:"angle" unit:"kg" displayunit:"kilograms" relativequantity:"true" min:"0.01" max:"3" nominal:"1" unbounded:"true"`
				}{},
			},
			&modelVariables{
				model: struct {
					A float64 `declaredtype:"foo" start:"1" derivative:"0.2" reinit:"true" quantity:"angle" unit:"kg" displayunit:"kilograms" relativequantity:"true" min:"0.01" max:"3" nominal:"1" unbounded:"true"`
				}{},
				scalars: []ScalarVariable{
					{
						Name:           "A",
						ValueReference: 1,
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeReal,
							Real: &RealVariable{
								DeclaredType: DeclaredType{
									DeclaredType: "foo",
								},
								RealType: RealType{
									TypeDefinition: TypeDefinition{
										Quantity: "angle",
									},
									Unit:             "kg",
									DisplayUnit:      "kilograms",
									RelativeQuantity: true,
									Min: func() *float64 {
										f := 0.01
										return &f
									}(),
									Max: func() *float64 {
										f := 3.0
										return &f
									}(),
									Nominal: func() *float64 {
										f := 1.0
										return &f
									}(),
									Unbounded: true,
								},
								Start: func() *float64 {
									f := 1.0
									return &f
								}(),
								Derivative: func() *float64 {
									f := 0.2
									return &f
								}(),
								Reinit: true,
							},
						},
					},
				},
			},
			false,
		},
		{
			"min tag for real variable is validated",
			args{
				struct {
					A float64 `min:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"reinit tag for real variable is validated",
			args{
				struct {
					A float64 `reinit:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"all tags for integer variable type can be set",
			args{
				struct {
					A int32 `declaredtype:"foo" start:"1"  quantity:"angle" min:"1" max:"3"`
				}{},
			},
			&modelVariables{
				model: struct {
					A int32 `declaredtype:"foo" start:"1"  quantity:"angle" min:"1" max:"3"`
				}{},
				scalars: []ScalarVariable{
					{
						Name:           "A",
						ValueReference: 1,
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeInteger,
							Integer: &IntegerVariable{
								DeclaredType: DeclaredType{
									DeclaredType: "foo",
								},
								Start: func() *int32 {
									i := int32(1)
									return &i
								}(),
								IntegerType: IntegerType{
									TypeDefinition: TypeDefinition{
										Quantity: "angle",
									},
									Min: func() *int32 {
										i := int32(1)
										return &i
									}(),
									Max: func() *int32 {
										i := int32(3)
										return &i
									}(),
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"min tag for integer variable is validated",
			args{
				struct {
					A int32 `min:"foo"`
				}{},
			},
			nil,
			true,
		},
		{
			"start tag for integer variable is validated",
			args{
				struct {
					A int32 `max:"4.5"`
				}{},
			},
			nil,
			true,
		},
		{
			"all tags for string variable type can be set",
			args{
				struct {
					A string `declaredtype:"foo" start:"potato" quantity:"bar"`
				}{},
			},
			&modelVariables{
				model: struct {
					A string `declaredtype:"foo" start:"potato" quantity:"bar"`
				}{},
				scalars: []ScalarVariable{
					{
						Name:           "A",
						ValueReference: 1,
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeString,
							String: &StringVariable{
								DeclaredType: DeclaredType{
									DeclaredType: "foo",
								},
								StringType: StringType{
									TypeDefinition: TypeDefinition{
										Quantity: "bar",
									},
								},
								Start: "potato",
							},
						},
					},
				},
			},
			false,
		},
		{
			"all tags for boolean variable type can be set",
			args{
				struct {
					A bool `declaredtype:"foo" start:"true" quantity:"bar"`
				}{},
			},
			&modelVariables{
				model: struct {
					A bool `declaredtype:"foo" start:"true" quantity:"bar"`
				}{},
				scalars: []ScalarVariable{
					{
						Name:           "A",
						ValueReference: 1,
						ScalarVariableType: &ScalarVariableType{
							variableType: VariableTypeBoolean,
							Boolean: &BooleanVariable{
								DeclaredType: DeclaredType{
									DeclaredType: "foo",
								},
								Start: func() *bool {
									b := true
									return &b
								}(),
								BooleanType: BooleanType{
									TypeDefinition: TypeDefinition{
										Quantity: "bar",
									},
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"start tag for boolean variable is validated",
			args{
				struct {
					A bool `start:"4.5"`
				}{},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewModelVariables(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewModelVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_modelVariables_Variables(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	tests := []struct {
		name   string
		fields fields
		want   []ScalarVariable
	}{
		{
			"scalar variables are returned",
			fields{
				scalars: []ScalarVariable{
					{},
				},
			},
			[]ScalarVariable{
				{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if got := m.Variables(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelVariables.Variables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelVariables_Encode(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			"standard model can be encoded successfully",
			fields{
				model: &struct {
					A float64
				}{
					A: 42,
				},
			},
			[]byte{18, 255, 129, 3, 1, 2, 255, 130, 0, 1, 1, 1, 1, 65, 1, 8, 0, 0, 0, 7, 255, 130, 1, 254, 69, 64, 0},
			false,
		},
		{
			"encoding error is returned when field is private",
			fields{
				model: struct {
					a float64
				}{
					a: 42,
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			got, err := m.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelVariables.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelVariables_Decode(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		rs []byte
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantModel interface{}
	}{
		{
			"Decode should update model values",
			fields{
				model: &struct {
					A float64
				}{
					A: 1,
				},
			},
			args{
				[]byte{18, 255, 129, 3, 1, 2, 255, 130, 0, 1, 1, 1, 1, 65, 1, 8, 0, 0, 0, 7, 255, 130, 1, 254, 69, 64, 0},
			},
			false,
			&struct {
				A float64
			}{
				A: 42,
			},
		},
		{
			"Decode error is returned",
			fields{
				model: &struct {
					A float64
				}{
					A: 1,
				},
			},
			args{
				[]byte{0},
			},
			true,
			&struct {
				A float64
			}{
				A: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if err := m.Decode(tt.args.rs); (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.wantModel, m.model) {
				t.Errorf("Want model %v got %v", tt.wantModel, m.model)
			}
		})
	}
}

func Test_modelVariables_GetReal(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantFs  []float64
		wantErr bool
	}{
		{
			"Multiple real values can be selected",
			fields{
				model: &struct {
					A float64
					B float64
				}{
					A: 1.1,
					B: 2.2,
				},
			},
			args{
				ValueReference{1, 2},
			},
			[]float64{1.1, 2.2},
			false,
		},
		{
			"Error returned if model is not a struct",
			fields{
				model: 42,
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
		{
			"Error returned if field value is not float64",
			fields{
				model: &struct {
					A string
				}{
					A: "foo",
				},
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
		{
			"Value reference out of bounds returns error",
			fields{
				model: &struct {
					A float64
				}{
					A: 1.1,
				},
			},
			args{
				ValueReference{2},
			},
			nil,
			true,
		},
		{
			"Value reference is 1-based index",
			fields{
				model: &struct {
					A float64
				}{
					A: 1.1,
				},
			},
			args{
				ValueReference{0},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			gotFs, err := m.GetReal(tt.args.vr)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.GetReal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFs, tt.wantFs) {
				t.Errorf("modelVariables.GetReal() = %v, want %v", gotFs, tt.wantFs)
			}
		})
	}
}

func Test_modelVariables_GetInteger(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantIs  []int32
		wantErr bool
	}{
		{
			"Multiple integer values can be selected",
			fields{
				model: &struct {
					A int32
					B int32
					C float64
				}{
					A: 1,
					B: 2,
					C: 1.2,
				},
			},
			args{
				ValueReference{1, 2},
			},
			[]int32{1, 2},
			false,
		},
		{
			"Error returned if model is not a struct",
			fields{
				model: 42,
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
		{
			"Error returned if field value is not int32",
			fields{
				model: &struct {
					A string
				}{
					A: "foo",
				},
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			gotIs, err := m.GetInteger(tt.args.vr)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.GetInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIs, tt.wantIs) {
				t.Errorf("modelVariables.GetInteger() = %v, want %v", gotIs, tt.wantIs)
			}
		})
	}
}

func Test_modelVariables_GetBoolean(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantBs  []bool
		wantErr bool
	}{
		{
			"Multiple boolean values can be selected",
			fields{
				model: &struct {
					A string
					B bool
					C bool
				}{
					A: "1",
					B: true,
					C: false,
				},
			},
			args{
				ValueReference{2, 3},
			},
			[]bool{true, false},
			false,
		},
		{
			"Error returned if model is not a struct",
			fields{
				model: false,
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
		{
			"Error returned if field value is not bool",
			fields{
				model: &struct {
					A string
				}{
					A: "foo",
				},
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			gotBs, err := m.GetBoolean(tt.args.vr)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.GetBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBs, tt.wantBs) {
				t.Errorf("modelVariables.GetBoolean() = %v, want %v", gotBs, tt.wantBs)
			}
		})
	}
}

func Test_modelVariables_GetString(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSs  []string
		wantErr bool
	}{
		{
			"Multiple string values can be selected",
			fields{
				model: &struct {
					A string
					B bool
					C string
				}{
					A: "foo",
					B: true,
					C: "bar",
				},
			},
			args{
				ValueReference{1, 3},
			},
			[]string{"foo", "bar"},
			false,
		},
		{
			"Error returned if model is not a struct",
			fields{
				model: 32,
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
		{
			"Error returned if field value is not string",
			fields{
				model: &struct {
					A bool
				}{
					A: false,
				},
			},
			args{
				ValueReference{1},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			gotSs, err := m.GetString(tt.args.vr)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSs, tt.wantSs) {
				t.Errorf("modelVariables.GetString() = %v, want %v", gotSs, tt.wantSs)
			}
		})
	}
}

func Test_modelVariables_SetReal(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
		fs []float64
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantModel interface{}
	}{
		{
			"Setting multiple field reals should work",
			fields{
				model: &struct {
					A float64
					B float64
				}{
					A: 1.1,
					B: 2.2,
				},
			},
			args{
				ValueReference{1, 2},
				[]float64{-1.2, 5},
			},
			false,
			&struct {
				A float64
				B float64
			}{
				A: -1.2,
				B: 5,
			},
		},
		{
			"Variable reference must be within range",
			fields{
				model: &struct {
					A float64
				}{
					A: 1.1,
				},
			},
			args{
				ValueReference{2},
				[]float64{5},
			},
			true,
			&struct {
				A float64
			}{
				A: 1.1,
			},
		},
		{
			"setting of wrong type returns error",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{1},
				[]float64{1},
			},
			true,
			&struct {
				A bool
			}{},
		},
		{
			"Value references and reals should be same length",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{1},
				[]float64{1, 2},
			},
			true,
			&struct {
				A bool
			}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if err := m.SetReal(tt.args.vr, tt.args.fs); (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.SetReal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m.model, tt.wantModel) {
				t.Errorf("Want model %v got %v", m.model, tt.wantModel)
			}
		})
	}
}

func Test_modelVariables_SetInteger(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
		is []int32
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantModel interface{}
	}{
		{
			"Setting multiple field reals should work",
			fields{
				model: &struct {
					A int32
					B bool
					C int32
				}{
					B: true,
				},
			},
			args{
				ValueReference{1, 3},
				[]int32{1, 2},
			},
			false,
			&struct {
				A int32
				B bool
				C int32
			}{
				A: 1,
				B: true,
				C: 2,
			},
		},
		{
			"Variable reference must be within range",
			fields{
				model: &struct {
					A int32
				}{},
			},
			args{
				ValueReference{2},
				[]int32{5},
			},
			true,
			&struct {
				A int32
			}{},
		},
		{
			"setting of wrong type returns error",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{1},
				[]int32{1},
			},
			true,
			&struct {
				A bool
			}{},
		},
		{
			"Value references and integers should be same length",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{1},
				[]int32{1, 2},
			},
			true,
			&struct {
				A bool
			}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if err := m.SetInteger(tt.args.vr, tt.args.is); (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.SetInteger() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m.model, tt.wantModel) {
				t.Errorf("Want model %v got %v", m.model, tt.wantModel)
			}
		})
	}
}

func Test_modelVariables_SetBoolean(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
		bs []bool
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantModel interface{}
	}{
		{
			"Setting multiple field booleans should work",
			fields{
				model: &struct {
					A bool
					B bool
					C string
				}{
					C: "foo",
				},
			},
			args{
				ValueReference{1, 2},
				[]bool{true, false},
			},
			false,
			&struct {
				A bool
				B bool
				C string
			}{
				A: true,
				B: false,
				C: "foo",
			},
		},
		{
			"Variable reference must be within range",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{2},
				[]bool{true},
			},
			true,
			&struct {
				A bool
			}{},
		},
		{
			"setting of wrong type returns error",
			fields{
				model: &struct {
					A string
				}{},
			},
			args{
				ValueReference{1},
				[]bool{true},
			},
			true,
			&struct {
				A string
			}{},
		},
		{
			"Value references and integers should be same length",
			fields{
				model: &struct {
					A bool
				}{},
			},
			args{
				ValueReference{1},
				[]bool{true, false},
			},
			true,
			&struct {
				A bool
			}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if err := m.SetBoolean(tt.args.vr, tt.args.bs); (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.SetBoolean() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m.model, tt.wantModel) {
				t.Errorf("Want model %v got %v", m.model, tt.wantModel)
			}
		})
	}
}

func Test_modelVariables_SetString(t *testing.T) {
	type fields struct {
		model   interface{}
		scalars []ScalarVariable
	}
	type args struct {
		vr ValueReference
		ss []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantModel interface{}
	}{
		{
			"Setting multiple field strings should work",
			fields{
				model: &struct {
					A string
					B string
				}{
					A: "bar",
				},
			},
			args{
				ValueReference{1, 2},
				[]string{"foo", "bar"},
			},
			false,
			&struct {
				A string
				B string
			}{
				A: "foo",
				B: "bar",
			},
		},
		{
			"Variable reference must be within range",
			fields{
				model: &struct {
					A string
				}{},
			},
			args{
				ValueReference{2},
				[]string{"foo"},
			},
			true,
			&struct {
				A string
			}{},
		},
		{
			"setting of wrong type returns error",
			fields{
				model: &struct {
					A float64
				}{},
			},
			args{
				ValueReference{1},
				[]string{"foo"},
			},
			true,
			&struct {
				A float64
			}{},
		},
		{
			"Value references and integers should be same length",
			fields{
				model: &struct {
					A string
				}{},
			},
			args{
				ValueReference{1},
				[]string{"foo", "bar"},
			},
			true,
			&struct {
				A string
			}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := modelVariables{
				model:   tt.fields.model,
				scalars: tt.fields.scalars,
			}
			if err := m.SetString(tt.args.vr, tt.args.ss); (err != nil) != tt.wantErr {
				t.Errorf("modelVariables.SetString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m.model, tt.wantModel) {
				t.Errorf("Want model %v got %v", m.model, tt.wantModel)
			}
		})
	}
}
