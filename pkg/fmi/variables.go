package fmi

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ModelVariables encapsulates model state as fmi compatible variables.
// Implements value getting and settings and state encoding.
type ModelVariables interface {
	ValueGetterSetter
	StateEncoder
	StateDecoder

	// Variables returns scalar variables to be used in model description
	Variables() []ScalarVariable
}

type modelVariables struct {
	model   interface{}
	scalars []ScalarVariable
}

// NewModelVariables reflects the provided value to create a model variables list.
// The type should be a struct with all exported fields.
// Struct field tags are used to annotate the fields so we can infer model structure.
// This implementation uses gob encoding to handle state transmission from library.
func NewModelVariables(model interface{}) (ModelVariables, error) {
	st := reflect.TypeOf(model)
	if st.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Requires struct kind, got %s", st.Kind())
	}
	nf := st.NumField()
	if nf == 0 {
		return nil, errors.New("Model struct has no fields")
	}
	svs := make([]ScalarVariable, nf)
	for i := 0; i < nf; i++ {
		f := st.Field(i)
		v, err := parseFieldVariable(f)
		if err != nil {
			return nil, fmt.Errorf("Error parsing model variable field: %w", err)
		}
		svs[i] = v
	}
	return &modelVariables{
		model:   model,
		scalars: svs,
	}, nil
}

func (m modelVariables) Variables() []ScalarVariable {
	return m.scalars
}

func (m modelVariables) Encode() ([]byte, error) {
	bs := &bytes.Buffer{}
	enc := gob.NewEncoder(bs)
	if err := enc.Encode(m.model); err != nil {
		return nil, fmt.Errorf("Error gob encoding model variable state: %w", err)
	}
	return bs.Bytes(), nil
}

func (m *modelVariables) Decode(rs []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(rs))
	if err := dec.Decode(m.model); err != nil {
		return err
	}
	return nil
}

func (m modelVariables) GetReal(vr ValueReference) (fs []float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error getting real field for value references %v", vr)
		}
	}()
	vs, err := m.fieldValues(vr)
	if err != nil {
		return nil, fmt.Errorf("Error getting real fields for value references %v: %w", vr, err)
	}
	fvs := make([]float64, len(vs))
	for i, v := range vs {
		fvs[i] = v.Float()
	}
	fs = fvs
	return
}

func (m modelVariables) GetInteger(vr ValueReference) (is []int32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error getting integer field for value references %v", vr)
		}
	}()
	vs, err := m.fieldValues(vr)
	if err != nil {
		return nil, fmt.Errorf("Error getting integer fields for value references %v: %w", vr, err)
	}
	ivs := make([]int32, len(vs))
	for i, v := range vs {
		ivs[i] = int32(v.Int())
	}
	is = ivs
	return
}

func (m modelVariables) GetBoolean(vr ValueReference) (bs []bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error getting boolean field for value references %v", vr)
		}
	}()
	vs, err := m.fieldValues(vr)
	if err != nil {
		return nil, fmt.Errorf("Error getting boolean fields for value references %v: %w", vr, err)
	}
	bvs := make([]bool, len(vs))
	for i, v := range vs {
		bvs[i] = v.Bool()
	}
	bs = bvs
	return
}

func (m modelVariables) GetString(vr ValueReference) ([]string, error) {
	vs, err := m.fieldValues(vr)
	if err != nil {
		return nil, fmt.Errorf("Error getting string fields for value references %v: %w", vr, err)
	}
	svs := make([]string, len(vs))
	for i, v := range vs {
		// reflect Value.String() is a special case
		// we don't just want a string representation of any type
		if v.Kind() != reflect.String {
			return nil, fmt.Errorf("Field type at index %d is not a string", i)
		}
		svs[i] = v.String()
	}
	return svs, nil
}

func (m modelVariables) SetReal(ValueReference, []float64) error {
	return nil
}

func (m modelVariables) SetInteger(ValueReference, []int32) error {
	return nil
}

func (m modelVariables) SetBoolean(ValueReference, []bool) error {
	return nil
}

func (m modelVariables) SetString(ValueReference, []string) error {
	return nil
}

func (m modelVariables) fieldValues(vr ValueReference) (vs []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Field index is out of bounds or model is not a struct")
		}
	}()
	v := reflect.ValueOf(m.model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	vs = make([]reflect.Value, len(vr))
	for i, vi := range vr {
		// value references are 1-based indexes
		vi = vi - 1
		vs[i] = v.Field(int(vi))
	}
	return
}

func parseFieldVariable(field reflect.StructField) (sv ScalarVariable, err error) {
	if field.PkgPath != "" {
		err = fmt.Errorf("Model field %s is unexported and cannot be set or serialized", field.Name)
		return
	}
	if len(field.Index) != 1 {
		err = fmt.Errorf("Model field %s is embedded. This is not allowed", field.Name)
		return
	}

	v, err := parseFieldType(field)
	if err != nil {
		return
	}

	tags := field.Tag
	canHandleMultipleSetPerTimeInstant, err := parseBoolTag(tags, "canhandlemultiplesetpertimeinstant")
	if err != nil {
		return
	}

	var causality VariableCausality
	if err = parseVariableEnumTag(tags, "causality", &causality); err != nil {
		return
	}

	var variability VariableVariability
	if err = parseVariableEnumTag(tags, "variability", &variability); err != nil {
		return
	}

	var initial *VariableInitial
	if _, got := tags.Lookup("initial"); got {
		var init VariableInitial
		if err = parseVariableEnumTag(tags, "initial", &init); err != nil {
			return
		}
		initial = &init
	}

	sv.scalarVariable = v
	sv.Name = field.Name
	sv.Description = tags.Get("description")
	// Value references are a 1-based index
	sv.ValueReference = uint(field.Index[0] + 1)
	sv.CanHandleMultipleSetPerTimeInstant = canHandleMultipleSetPerTimeInstant
	sv.Causality = causality
	sv.Variability = variability
	sv.Initial = initial
	return
}

func parseFieldType(field reflect.StructField) (*scalarVariable, error) {
	switch field.Type.Kind() {
	case reflect.Float64:
		r, err := parseRealTags(field)
		if err != nil {
			return nil, fmt.Errorf("Error parsing Real tags for model variable %s: %w", field.Name, err)
		}
		return &scalarVariable{
			variableType: VariableTypeReal,
			Real:         r,
		}, nil
	case reflect.Int32:
		i, err := parseIntegerTags(field)
		if err != nil {
			return nil, fmt.Errorf("Error parsing Integer tags for model variable %s: %w", field.Name, err)
		}
		return &scalarVariable{
			variableType: VariableTypeInteger,
			Integer:      i,
		}, nil
	case reflect.String:
		return &scalarVariable{
			variableType: VariableTypeString,
			String:       parseStringTags(field),
		}, nil
	case reflect.Bool:
		b, err := parseBooleanTags(field)
		if err != nil {
			return nil, fmt.Errorf("Error parsing Boolean tags for model variable %s: %w", field.Name, err)
		}
		return &scalarVariable{
			variableType: VariableTypeBoolean,
			Boolean:      b,
		}, nil
	}

	return nil, fmt.Errorf("Model struct field type %s not supported", field.Type)
}

func parseRealTags(field reflect.StructField) (*RealVariable, error) {
	tags := field.Tag
	min, err := parseFloatTag(tags, "min")
	if err != nil {
		return nil, err
	}
	max, err := parseFloatTag(tags, "max")
	if err != nil {
		return nil, err
	}
	nominal, err := parseFloatTag(tags, "nominal")
	if err != nil {
		return nil, err
	}
	relativeQuantity, err := parseBoolTag(tags, "relativequantity")
	if err != nil {
		return nil, err
	}
	unbounded, err := parseBoolTag(tags, "unbounded")
	if err != nil {
		return nil, err
	}
	start, err := parseFloatTag(tags, "start")
	if err != nil {
		return nil, err
	}
	derivative, err := parseFloatTag(tags, "derivative")
	if err != nil {
		return nil, err
	}
	reinit, err := parseBoolTag(tags, "reinit")
	if err != nil {
		return nil, err
	}

	return &RealVariable{
		RealType: RealType{
			Min:              min,
			Max:              max,
			Nominal:          nominal,
			Unit:             tags.Get("unit"),
			DisplayUnit:      tags.Get("displayunit"),
			RelativeQuantity: relativeQuantity,
			Unbounded:        unbounded,
			typeDefinition:   parseTypeDefinitionTag(tags),
		},
		declaredType: parseDeclaredTypeTag(tags),
		Start:        start,
		Derivative:   derivative,
		Reinit:       reinit,
	}, nil
}

func parseIntegerTags(field reflect.StructField) (*IntegerVariable, error) {
	tags := field.Tag
	min, err := parseIntTag(tags, "min")
	if err != nil {
		return nil, err
	}
	max, err := parseIntTag(tags, "max")
	if err != nil {
		return nil, err
	}
	start, err := parseIntTag(tags, "start")
	if err != nil {
		return nil, err
	}

	return &IntegerVariable{
		declaredType: parseDeclaredTypeTag(tags),
		IntegerType: IntegerType{
			typeDefinition: parseTypeDefinitionTag(tags),
			Min:            min,
			Max:            max,
		},
		Start: start,
	}, nil
}

func parseStringTags(field reflect.StructField) *StringVariable {
	tags := field.Tag

	return &StringVariable{
		declaredType: parseDeclaredTypeTag(tags),
		StringType: StringType{
			typeDefinition: parseTypeDefinitionTag(tags),
		},
		Start: tags.Get("start"),
	}
}

func parseBooleanTags(field reflect.StructField) (*BooleanVariable, error) {
	tags := field.Tag

	var start *bool
	if _, ok := tags.Lookup("start"); ok {
		s, err := parseBoolTag(tags, "start")
		if err != nil {
			return nil, err
		}
		start = &s
	}

	return &BooleanVariable{
		declaredType: parseDeclaredTypeTag(tags),
		BooleanType: BooleanType{
			typeDefinition: parseTypeDefinitionTag(tags),
		},
		Start: start,
	}, nil
}

func parseFloatTag(t reflect.StructTag, n string) (v *float64, err error) {
	s, ok := t.Lookup(n)
	if !ok {
		return
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = fmt.Errorf("Error parsing Real model struct tag %s: %w", n, err)
	}
	v = &f
	return
}

func parseIntTag(t reflect.StructTag, n string) (v *int32, err error) {
	s, ok := t.Lookup(n)
	if !ok {
		return
	}
	i, err := strconv.ParseInt(s, 10, 32)
	i32 := int32(i)
	if err != nil {
		err = fmt.Errorf("Error parsing Integer model struct tag %s: %w", n, err)
	}
	v = &i32
	return
}

func parseBoolTag(t reflect.StructTag, n string) (b bool, err error) {
	s, ok := t.Lookup(n)
	if !ok {
		return
	}
	b, err = strconv.ParseBool(s)
	if err != nil {
		err = fmt.Errorf("Error parsing Boolean struct tag %s: %w", n, err)
	}
	return
}

func parseDeclaredTypeTag(t reflect.StructTag) declaredType {
	return declaredType{
		DeclaredType: t.Get("declaredtype"),
	}
}

func parseTypeDefinitionTag(t reflect.StructTag) typeDefinition {
	return typeDefinition{
		Quantity: t.Get("quantity"),
	}
}

func parseVariableEnumTag(t reflect.StructTag, n string, m encoding.TextUnmarshaler) error {
	s := t.Get(n)
	err := m.UnmarshalText([]byte(s))
	if err != nil {
		return fmt.Errorf("Error unmarshaling model variable enum %s: %w", n, err)
	}
	return nil
}
