package fmi

import (
	"encoding/xml"
	"fmt"
	"time"
)

const (
	// Variable causality local is default
	VariableCausalityLocal VariableCausality = iota
	VariableCausalityParameter
	VariableCausalityCalculatedParameter
	VariableCausalityInput
	VariableCausalityOutput
	VariableCausalityIndependent
)

const (
	// Variable variability continuous is default
	VariableVariabilityContinuous VariableVariability = iota
	VariableVariabilityConstant
	VariableVariabilityFixed
	VariableVariabilityTunable
	VariableVariabilityDiscrete
)

const (
	// Variable initial will be unset if omitted
	VariableInitialExact VariableInitial = iota
	VariableInitialApprox
	VariableInitialCalculated
)

var (
	variableCausalityEnum   = [...]string{"local", "parameter", "calculatedParameter", "input", "output", "independent"}
	variableVariabilityEnum = [...]string{"continuous", "constant", "fixed", "tunable", "discrete"}
	variableInitialEnum     = [...]string{"exact", "approx", "calculated"}
)

// ModelDescription represents root node of a modelDescription.xml file
type ModelDescription struct {
	modelDescriptionStatic
	Name                    string      `xml:"modelName,attr"`
	GUID                    string      `xml:"guid,attr"`
	Description             string      `xml:"description,attr,omitempty"`
	Author                  string      `xml:"author,attr,omitempty"`
	ModelVersion            string      `xml:"version,attr,omitempty"`
	Copyright               string      `xml:"copyright,attr,omitempty"`
	License                 string      `xml:"license,attr,omitempty"`
	GenerationTool          string      `xml:"generationTool,attr,omitempty"`
	GenerationDateAndTime   *time.Time  `xml:"generationDateAndTime,attr,omitempty"`
	NumberOfEventIndicators uint        `xml:"numberOfEventIndicators,attr"`
	DefaultExperiment       *Experiment `xml:"DefaultExperiment,omitempty"`
	VendorAnnotations       *struct {
		Tool []ToolAnnotation `xml:"Tool,omitempty"`
	} `xml:"VendorAnnotations,omitempty"`
}

type modelDescriptionStatic struct {
	XMLName    xml.Name `xml:"fmiModelDescription"`
	FMIVersion string   `xml:"fmiVersion,attr"`
}

type typeDefinition struct {
	// Quantity is the physical quantity of the variable, for example Angle or Energy.
	Quantity string `xml:"quantity,attr,omitempty"`
}

// RealType is used for type definitions and in variable types
type RealType struct {
	typeDefinition

	// Unit of the variable defined with UnitDefinitions.Unit.name that is used for the model equations.
	Unit string `xml:"unit,attr,omitempty"`

	/*
		DisplayUnit default display unit. The conversion to the “unit” is defined with the element
		“<fmiModelDescription><UnitDefinitions> ”. If the corresponding
		“displayUnit” is not defined under <UnitDefinitions><Unit><DisplayUnit>, then displayUnit is ignored. It is an error if
		displayUnit is defined in element Real, but unit is not, or unit is not
		defined under <UnitDefinitions><Unit>.
	*/
	DisplayUnit string `xml:"displayUnit,attr,omitempty"`

	/*
		RelativeQuantity if set attribute to true, then the “offset” of “baseUnit” and
		“displayUnit” must be ignored (for example, 10 degree Celsius = 10 Kelvin
		if “relativeQuantity = true” and not 283.15 Kelvin).
	*/
	RelativeQuantity bool `xml:"relativeQuantity,attr,omitempty"`

	/*
		Min value of variable (variable Value ≥ min ). If not defined, the
		minimum is the largest negative number that can be represented on the
		machine. The min definition is information from the FMU to the environment
		defining the region in which the FMU is designed to operate.
	*/
	Min *float64 `xml:"min,attr,omitempty"`

	/*
		Max value of variable (variableValue ≤ max ). If not defined, the
		maximum is the largest positive number that can be represented on the
		machine. The max definition is information from the FMU to the environment
		defining the region in which the FMU is designed to operate.
	*/
	Max *float64 `xml:"max,attr,omitempty"`

	/*
		Nominal value of variable. If not defined and no other information about the
		nominal value is available, then nominal = 1 is assumed.
		[The nominal value of a variable can be, for example, used to determine the
		absolute tolerance for this variable as needed by numerical algorithms:
		absoluteTolerance = nominal * tolerance *0.01
		where tolerance is, for example, the relative tolerance defined in <DefaultExperiment>.
	*/
	Nominal *float64 `xml:"nominal,attr,omitempty"`

	/*
		Unbounded if true, indicates that during time integration, the variable gets a value much
		larger than its nominal value. [Typical examples are the
		monotonically increasing rotation angles of crank shafts and the longitudinal
		position of a vehicle along the track in long distance simulations. This
		information can, for example, be used to increase numerical stability and
		accuracy by setting the corresponding bound for the relative error to zero.
	*/
	Unbounded bool `xml:"unbounded,attr,omitempty"`
}

// IntegerType is used for type definitions and in variable types
type IntegerType struct {
	typeDefinition
	// Min is min value, see RealType.Min definition.
	Min int32 `xml:"min,attr,omitempty"`
	// Max is max value, see RealType.Max definition.
	Max int32 `xml:"max,attr,omitempty"`
}

// EnumerationType is used to define an enumeration that is referenced from variable enumeration types
type EnumerationType struct {
	typeDefinition
	/*
		Item of an enumeration has a sequence of “name” and “value” pairs. The
		values can be any integer number but must be unique within the same
		enumeration (in order that the mapping between “ name ” and “ value ” is
		bijective). An Enumeration element must have at least one Item.
	*/
	Item []EnumerationItem `xml:"item,omitempty"`
}

// EnumerationItem defines an enum item inside an enumeration type
type EnumerationItem struct {
	Name        string `xml:"name,attr"`
	Value       int32  `xml:"value,attr"`
	Description string `xml:"description,attr,omitempty"`
}

// ToolAnnotation is a tool specific annotation.
// Allows any xml to be embedded for another system to use.
type ToolAnnotation struct {
	Name     string `xml:"name,attr"`
	InnerXML string `xml:",innerxml"`
}

// Experiment element for model description default experiment
type Experiment struct {
	StartTime float64 `xml:"startTime,attr,omitempty"`
	StopTime  float64 `xml:"stopTime,attr,omitempty"`
	Tolerance float64 `xml:"tolerance,attr,omitempty"`
	StepSize  float64 `xml:"stepSize,attr,omitempty"`
}

func NewModelDescription() ModelDescription {
	return ModelDescription{
		modelDescriptionStatic: modelDescriptionStatic{
			FMIVersion: GetVersion(),
		},
	}
}

func (m ModelDescription) MarshallIndent() ([]byte, error) {
	bs, err := xml.MarshalIndent(m, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("Error marshalling model description: %w", err)
	}
	return []byte(xml.Header + string(bs)), nil
}

// ScalarVariable represents variable node in ModelVariables
type ScalarVariable struct {
	/*
		Name is the full, unique name of the variable. Every variable is uniquely identified within an
		FMU instance by this name or by its ScalarVariable index (the element position in
		the ModelVariables list; the first list element has index=1 ).
	*/
	Name string `xml:"name,attr"`

	/*
		ValueReference is a handle of the variable to efficiently identify the variable value in the model interface.
		This handle is a secret of the tool that generated the C functions; it is not required to be unique.
		The only guarantee is that valueReference is sufficient to identify the
		respective variable value in the call of the C functions. This implies that it is unique for a
		particular base data type ( Real, Integer/Enumeration, Boolean, String ) with
		exception of variables that have identical values (such variables are also called “alias”
		variables). This attribute is “required”.
	*/
	ValueReference uint `xml:"valueReference,attr"`

	// Description is optional describing the meaning of the variable
	Description string `xml:"description,attr,omitempty"`

	/*
		Causality is an enumeration that defines the causality of the variable. Allowed values of this
		enumeration:

		- "parameter": Independent parameter (a data value that is constant during the
		simulation and is provided by the environment and cannot be used in connections).
		variability must be "fixed" or "tunable" . initial must be exact or not
		present (meaning exact).

		- "calculatedParameter": A data value that is constant during the simulation and
		is computed during initialization or when tunable parameters change.
		variability must be "fixed" or "tunable" . initial must be "approx",
		"calculated" or not present (meaning calculated ).

		- "input": The variable value can be provided from another model or slave. It is not
		allowed to define initial.

		- "output": The variable value can be used by another model or slave.

		- "local": Local variable that is calculated from other variables or is a continuous-
		time state.

		- "independent": The independent variable (usually “time”). All variables are a
		function of this independent variable. variability must be "continuous". At
		most one ScalarVariable of an FMU can be defined as "independent". If no
		variable is defined as "independent" , it is implicitly present with name = "time"
		and unit = "s" .

		The default of causality is “ local ”.
	*/
	Causality VariableCausality `xml:"causality,attr"`

	/*
		Variability enumeration that defines the time dependency of the variable, in other words, it
		defines the time instants when a variable can change its value. [The purpose of this
		attribute is to define when a result value needs to be inquired and to be stored. For
		example, discrete variables change their values only at event instants
		(ModelExchange) or at a communication point (CoSimulation) and it is therefore
		only necessary to inquire them with fmi2GetXXX and store them at event times].
		Allowed values of this enumeration:

		- "constant": The value of the variable never changes.

		- "fixed": The value of the variable is fixed after initialization, in other words, after
		fmi2ExitInitializationMode was called the variable value does not change anymore.

		- "tunable": The value of the variable is constant between external events
		(ModelExchange) and between Communication Points (Co-Simulation) due to
		changing variables with causality = "parameter" or "input" and
		variability = "tunable". Whenever a parameter or input signal with
		variability = "tunable" changes, an event is triggered externally
		(ModelExchange), or the change is performed at the next Communication Point
		(Co-Simulation) and the variables with variability = "tunable" and causality
		= "calculatedParameter" or "output" must be newly computed.

		- "discrete": ModelExchange: The value of the variable is constant between external and
		internal events (= time, state, step events defined implicitly in the FMU).
		Co-Simulation: By convention, the variable is from a “real” sampled data system
		and its value is only changed at Communication Points (also inside the slave).

		- "continuous": Only a variable of type = "Real" can be “continuous”.
		ModelExchange: No restrictions on value changes.
		Co-Simulation: By convention, the variable is from a differential

		The default is “continuous”.
	*/
	Variability VariableVariability `xml:"variability,attr"`

	/*
		Initial is an enumeration that defines how the variable is initialized. It is not allowed to provide a
		value for initial if causality = "input" or "independent":

		- "exact": The variable is initialized with the start value (provided under Real ,
		Integer , Boolean , String or Enumeration ).

		- "approx": The variable is an iteration variable of an algebraic loop and the
		iteration at initialization starts with the start value.

		- "calculated": The variable is calculated from other variables during
		initialization. It is not allowed to provide a “start” value.
		If initial is not present, it is defined  based on causality and variability.
	*/
	Initial *VariableInitial `xml:"initial,attr,omitempty"`

	/*
		CanHandleMultipleSetPerTimeInstant is only for ModelExchange (if only CoSimulation FMU, this attribute must not be present.

		If both ModelExchange and CoSimulation FMU, this attribute is ignored for CoSimulation). only for variables with variability = "input" :

		If present with value = false then only one fmi2SetXXX call is allowed at one super
		dense time instant (model evaluation) on this variable. That is, this input is not allowed
		to appear in a (real) algebraic loop requiring multiple calls of fmi2SetXXX on this
		variable [for example, due to a Newton iteration].
		[This flag must be set by FMUs where (internal) discrete-time states are directly
		updated when assigned (xd := f(xd) instead of xd = f(previous(xd)), and at least one
		output depends on this input and on discrete states.
	*/
	CanHandleMultipleSetPerTimeInstant bool `xml:"canHandleMultipleSetPerTimeInstant,attr,omitempty"`

	// Annotations contains custom tool annotations for this variable
	Annotations []ToolAnnotation `xml:"annotations,omitempty"`
}

// VariableCausality enum for scalar variable
type VariableCausality uint

// VariableVariability enum for scalar variable
type VariableVariability uint

// VariableInitial enum for scalar variable
type VariableInitial uint

func (e VariableCausality) MarshalText() (text []byte, err error) {
	return enumMarshalText(int(e), variableCausalityEnum[:])
}

func (e *VariableCausality) UnmarshalText(text []byte) error {
	i, err := enumUnmarshalText(text, variableCausalityEnum[:])
	if err != nil {
		return err
	}
	*e = VariableCausality(i)
	return nil
}

func (e VariableVariability) MarshalText() (text []byte, err error) {
	return enumMarshalText(int(e), variableVariabilityEnum[:])
}

func (e *VariableVariability) UnmarshalText(text []byte) error {
	i, err := enumUnmarshalText(text, variableVariabilityEnum[:])
	if err != nil {
		return err
	}
	*e = VariableVariability(i)
	return nil
}

func (e VariableInitial) MarshalText() (text []byte, err error) {
	return enumMarshalText(int(e), variableInitialEnum[:])
}

func (e *VariableInitial) UnmarshalText(text []byte) error {
	i, err := enumUnmarshalText(text, variableInitialEnum[:])
	if err != nil {
		return err
	}
	*e = VariableInitial(i)
	return nil
}

func enumMarshalText(enum int, vs []string) (text []byte, err error) {
	if enum > len(vs) {
		err = fmt.Errorf("Index %d out of range %d", enum, len(vs))
		return
	}
	text = []byte(vs[enum])
	return
}

func enumUnmarshalText(text []byte, vs []string) (int, error) {
	// assume nil or empty text means default first value
	if len(text) == 0 {
		return 0, nil
	}
	enum := string(text)
	for i, v := range vs {
		if v == enum {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Enum %s does not exist", enum)
}
