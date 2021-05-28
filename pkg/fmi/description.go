package fmi

import (
	"encoding/xml"
	"fmt"
	"strings"
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

const (
	VariableTypeReal VariableType = iota + 1
	VariableTypeInteger
	VariableTypeBoolean
	VariableTypeString
	VariableTypeEnumeration
)

var (
	variableCausalityEnum   = [...]string{"local", "parameter", "calculatedParameter", "input", "output", "independent"}
	variableVariabilityEnum = [...]string{"continuous", "constant", "fixed", "tunable", "discrete"}
	variableInitialEnum     = [...]string{"exact", "approx", "calculated"}
)

// ModelDescription represents root node of a modelDescription.xml file
type ModelDescription struct {
	modelDescriptionStatic
	// Name is the name of the model as used in the modeling environment that generated the XML file.
	Name string `xml:"modelName,attr"`

	/*
		GUID is a “Globally Unique IDentifier” that used to check that
		the XML file is compatible with the C functions of the FMU.
		Set this constant in the FMU library so that the XML and library can be verified.
	*/
	GUID string `xml:"guid,attr"`

	// Description optional string with a brief description of the model.
	Description string `xml:"description,attr,omitempty"`

	// Author is optional string with the name and organization of the model author.
	Author string `xml:"author,attr,omitempty"`

	// Version is optional version of the model, for example “1.0”.
	Version string `xml:"version,attr,omitempty"`

	// Copyright is optional information on the intellectual property copyright for this FMU.
	Copyright string `xml:"copyright,attr,omitempty"`

	// License is optional information on the intellectual property licensing for this FMU.
	License string `xml:"license,attr,omitempty"`

	// GenerationTool is optional name of the tool that generated the XML file.
	GenerationTool string `xml:"generationTool,attr,omitempty"`
	/*
		GenerationDateAndTime is optional date and time when the XML file was generated.
		The format is a subset of “xs:dateTime” and should be: “YYYY-MM-DDThh:mm:ssZ".
	*/
	GenerationDateAndTime *time.Time `xml:"generationDateAndTime,attr,omitempty"`
	/*
		NumberOfEventIndicators is the (fixed) number of event indicators for an FMU based on FMI for Model Exchange.
		For Co-Simulation, this value is ignored.
	*/
	NumberOfEventIndicators uint `xml:"numberOfEventIndicators,attr,omitempty"`

	/*
		UnitDefinitions is a global list of unit and display unit definitions [for example, to convert
		display units into the units used in the model equations]. These
		definitions are used in the XML element “ModelVariables”.
	*/
	UnitDefinitions []Unit `xml:"UnitDefinitions>Unit,omitempty"`

	// TypeDefinitions to be shared by ModelVariables
	TypeDefinitions []SimpleType `xml:"TypeDefinitions>SimpleType,omitempty"`

	// DefaultExperiment is optional default experiment parameters.
	DefaultExperiment *Experiment `xml:"DefaultExperiment,omitempty"`

	// VendorAnnotations is optional data for vendor tools containing list of Tool elements
	VendorAnnotations []ToolAnnotation `xml:"VendorAnnotations>Tool,omitempty"`

	/*
		ModelVariables consists of ordered set of "ScalarVariable" elements.
		The first element has index = 1, the second index=2, etc. This ScalarVariable index is
		used in element ModelStructure to uniquely and efficiently refer to ScalarVariable definitions.
		A “ScalarVariable” represents a variable of primitive type, like a real or integer variable. For simplicity,
		only scalar variables are supported in the schema file in this version and structured entities (like arrays
		or records) have to be mapped to scalars.
	*/
	ModelVariables []ScalarVariable `xml:"ModelVariables>ScalarVariable"`

	/*
		ModelStructure defines the structure of the model. Especially, the ordered lists of
		outputs, continuous-time states and initial unknowns (the unknowns
		during Initialization Mode) are defined here.
		Furthermore, the dependency of
		the unknowns from the knowns can be optionally defined. [This
		information can be, for example, used to compute efficiently a sparse
		Jacobian for simulation, or to utilize the input/output dependency in
		order to detect that in some cases there are actually no algebraic
		loops when connecting FMUs together.]
	*/
	ModelStructure ModelStructure `xml:"ModelStructure"`
}

/*
	Unit is defined by its name attribute such as “N.m” or “N*m” or “Nm”,
	which must be unique with respect to all other defined elements of the UnitDefinitions list.
	If a variable is associated with a Unit , then the value of the variable has to be
	provided with the fmi2SetXXX functions and is returned by the fmi2GetXXX functions with respect to this Unit.
*/
type Unit struct {
	// Name of unit element, e.g. N.m, Nm, %/s. Name must be unique with respect to all other elements of the UnitDefinitions list
	Name string `xml:"name,attr"`
	// BaseUnit is optional unit conversion
	BaseUnit *BaseUnit `xml:"BaseUnit,omitempty"`
	// DisplayUnits for Unit to display types
	DisplayUnits []DisplayUnit `xml:"DisplayUnit,omitempty"`
}

// BaseUnit is used to convert Unit with factor and offset attributes
type BaseUnit struct {
	// KG exponent of SI base unit "kg"
	KG *int `xml:"kg,attr,omitempty"`
	// M exponent of SI base unit "m"
	M *int `xml:"m,attr,omitempty"`
	// S exponent of SI based unit "s"
	S *int `xml:"s,attr,omitempty"`
	// A exponent of SI based unit "A"
	A *int `xml:"A,attr,omitempty"`
	// K exponent of SI based unit "K"
	K *int `xml:"K,attr,omitempty"`
	// Mol exponent of SI based unit "mol"
	Mol *int `xml:"mol,attr,omitempty"`
	// CD exponent of SI based unit "cd"
	CD *int `xml:"cd,attr,omitempty"`
	// Rad exponent of SI based unit "rad"
	Rad *int `xml:"rad,attr,omitempty"`
	// Factor for base unit conversion
	Factor *float64 `xml:"factor,attr,omitempty"`
	// Offset for base unit conversion
	Offset *float64 `xml:"offset,attr,omitempty"`
}

// DisplayUnit defines unit conversion for display purposes with factor and offset.
// A value with respect to Unit is converted with respect to DisplayUnit by the equation: display_unit = factor*unit + offset.
type DisplayUnit struct {
	// Name of DisplayUnit element, e.g. if Unit name is "rad", DisplayUnit name might be "deg".
	// Name must be unique with respect to other names in same DisplayUnit list.
	Name string `xml:"name,attr"`
	// Factor for display unit conversion
	Factor *float64 `xml:"factor,attr,omitempty"`
	// Offset for display unit conversion
	Offset *float64 `xml:"offset,attr,omitempty"`
}

// SimpleType represents shared properties to be used by one or more ScalarVariables.
// One of the elements Real, Integer, Boolean, String or Enumeration must be present.
type SimpleType struct {
	// Name of SimpleType, unique with respect to all other elements in this list.
	// Name of SimpleType must be different to all "name"s of ScalarVariables.
	Name string `xml:"name,attr"`
	// Description of SimpleType.
	Description string `xml:"description,attr,omitempty"`
	// Real type
	Real *RealType `xml:"Real,omitempty"`
	// Integer type
	Integer *IntegerType `xml:"Integer,omitempty"`
	// Boolean type
	Boolean *BooleanType `xml:"Boolean,omitempty"`
	// String type
	String *StringType `xml:"String,omitempty"`
	// Enumeration type
	Enumeration *EnumerationType `xml:"Enumeration,omitempty"`
}

type modelDescriptionStatic struct {
	XMLName xml.Name `xml:"fmiModelDescription"`
	// FMIVersion is version for model exchange or co-simulation. Derived from C headers.
	FMIVersion string `xml:"fmiVersion,attr"`
	// VariableNamingConvention defines convention of variables. Set to "flat" in this library.
	VariableNamingConvention string `xml:"variableNamingConvention,attr"`
	// LogCategories are fixed log categories based on logger
	LogCategories []logCategory `xml:"LogCategories>Category,omitempty"`
}

type logCategory struct {
	// Name of log category, must be unique with respect to all other elements of LogCategories.
	Name string `xml:"name,attr"`
	// Description of log the category
	Description string `xml:"description,attr,omitempty"`
}

func buildLogCategories() []logCategory {
	cs := make([]logCategory, len(loggerCategories))
	for i, l := range loggerCategories {
		cs[i] = logCategory{
			Name: l.String(),
		}
	}
	return cs
}

type TypeDefinition struct {
	// Quantity is the physical quantity of the variable, for example Angle or Energy.
	Quantity string `xml:"quantity,attr,omitempty"`
}

// RealType is used for type definitions and in variable types
type RealType struct {
	TypeDefinition

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
	TypeDefinition
	// Min is min value, see RealType.Min definition.
	Min *int32 `xml:"min,attr,omitempty"`
	// Max is max value, see RealType.Max definition.
	Max *int32 `xml:"max,attr,omitempty"`
}

// StringType is used for type definition and in variable types
type StringType struct {
	TypeDefinition
}

// BooleanType is used for type definitions and in variable types
type BooleanType struct {
	TypeDefinition
}

// EnumerationType is used to define an enumeration that is referenced from variable enumeration types
type EnumerationType struct {
	TypeDefinition
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
	// StartTime is optional start time
	StartTime *float64 `xml:"startTime,attr,omitempty"`
	// StopTime is optional stop time
	StopTime *float64 `xml:"stopTime,attr,omitempty"`
	// Tolerance is default tolerance
	Tolerance *float64 `xml:"tolerance,attr,omitempty"`
	// StepSize is default step size
	StepSize *float64 `xml:"stepSize,attr,omitempty"`
}

func NewModelDescription() ModelDescription {
	return ModelDescription{
		modelDescriptionStatic: modelDescriptionStatic{
			FMIVersion:               GetVersion(),
			VariableNamingConvention: "flat",
			LogCategories:            buildLogCategories(),
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
	Annotations []ToolAnnotation `xml:"Annotations>Tool,omitempty"`

	*ScalarVariableType
}

type ScalarVariableType struct {
	variableType VariableType

	// Real holds attributes for real (float64) variable
	Real *RealVariable `xml:",omitempty"`
	// Integer holds attributes for integer (int32) variable
	Integer *IntegerVariable `xml:",omitempty"`
	// Boolean holds attributes for boolean variable
	Boolean *BooleanVariable `xml:",omitempty"`
	// String holds attributes for string variable
	String *StringVariable `xml:",omitempty"`
}

func (v *ScalarVariableType) updateVariableType() {
	if v.variableType != 0 {
		return
	}

	if v.Real != nil {
		v.variableType = VariableTypeReal
	} else if v.Integer != nil {
		v.variableType = VariableTypeInteger
	} else if v.Boolean != nil {
		v.variableType = VariableTypeBoolean
	} else if v.String != nil {
		v.variableType = VariableTypeString
	} else {
		panic("Scalar variable type is empty")
	}
}

func (v *ScalarVariableType) Type() VariableType {
	v.updateVariableType()

	return v.variableType
}

// VariableCausality enum for scalar variable
type VariableCausality uint

// VariableVariability enum for scalar variable
type VariableVariability uint

// VariableInitial enum for scalar variable
type VariableInitial uint

// VariableType enum for type definitions and scalars variables
type VariableType uint

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

type DeclaredType struct {
	/*
		DeclaredType is the name of type defined with TypeDefinitions / SimpleType.
		The value defined in the corresponding TypeDefinition is used as
		default. [For example, if “ min ” is present both in Real (of TypeDefinition ) and
		in “Real” (of ScalarVariable ), then the “min” of ScalarVariable is actually
		used.] For Real, Integer, Boolean, String, this attribute is optional. For
		Enumeration it is required, because the Enumeration items are defined in
		TypeDefinitions / SimpleType.
	*/
	DeclaredType string `xml:"declaredType,attr,omitempty"`
}

// RealVariable is used in scalar variables to define Reals
type RealVariable struct {
	RealType
	DeclaredType

	/*
		Start is initial or guess value of variable. This value is also stored in the Go data structures.
		[Therefore, calling fmi2SetXXX to set start values is only necessary, if a different
		value as stored in the xml file is desired. WARNING: It is not recommended to
		change the start values in the modelDescription.xml file of an FMU, as this
		would break the consistency with the hard-coded start values in the C Code.
		This could lead to unpredictable behaviour of the FMU in different importing tools,
		as it is not mandatory to call fmi2SetXXX to set start values during initialization.
		If initial = ′′exact′′ or ′′approx′′ or causality = ′′input′′ , a start value must be provided.
		If initial = ′′calculated′′ or causality = ′′independent′′ , it is not allowed to provide a start value.
		Variables with causality = "parameter" or "input" , as well as variables with variability = "constant" , must have a "start" value.

		If causality = "parameter" , the start -value is the value of it.

		If causality = "input" , the start value is used by the model as value of
		the input, if the input is not set by the environment.

		If variability = "constant" , the start value is the value of the constant.

		If causality = "output" or "local" then the start value is either an
		“initial” or a “guess” value, depending on the setting of attribute "initial" .
	*/
	Start *float64 `xml:"start,attr,omitempty"`

	/*
		Derivative, if present, this variable is the derivative of variable with ScalarVariable index
		"derivative". [For example, if there are 10 ScalarVariables and derivative = 3 for
		ScalarVariable 8, then ScalarVariable 8 is the derivative of ScalarVariable 3 with
		respect to the independent variable (usually time). This information might be
		especially used if an input or an output is the derivative of another input or output,
		or to define the states.]
		The state derivatives of an FMU are listed under element
		<ModelStructure><Derivatives> . All ScalarVariables listed in this element
		must have attribute derivative (in order that the continuous-time states are
		uniquely defined).
	*/
	Derivative *float64 `xml:"derivative,attr,omitempty"`

	/*
		Reinit only for ModelExchange (if only CoSimulation FMU, this attribute must not be
		present. If both ModelExchange and CoSimulation FMU, this attribute is ignored
		for CoSimulation):
		Can only be present for a continuous-time state.
		If true, state can be reinitialized at an event by the FMU.
		If false, state will not be reinitialized at an event by the FMU.
	*/
	Reinit bool `xml:"reinit,attr,omitempty"`
}

// IntegerVariable is used in scalar variables to define Integer
type IntegerVariable struct {
	IntegerType
	DeclaredType
	// Start is defined as per RealVariable.Start
	Start *int32 `xml:"start,attr,omitempty"`
}

// BooleanVariable is used in scalar variables to define Boolean
type BooleanVariable struct {
	BooleanType
	DeclaredType
	// Start is defined as per RealVariable.Start
	Start *bool `xml:"start,attr,omitempty"`
}

// StringVariable is used in scalar variables to define String
type StringVariable struct {
	StringType
	DeclaredType
	// Start is defined as per RealVariable.Start
	Start string `xml:"start,attr,omitempty"`
}

/*
	ModelStructure is with respect to the underlying model equations, independently how these model equations are solved.
	[For example, when exporting a model both in Model Exchange and Co-Simulation format; then the model structure is identical in both cases.
	The Co-Simulation FMU has either an integrator included that solves the model equations, or the discretization formula of the integrator and the
	model equations are solved together (“inline integration”). In both cases the model has the same
	continuous-time states. In the second case the internal implementation is a discrete -time system, but
	from the outside this is still a continuous-time model that is solved with an integration method.]

	The required part defines an ordering of the outputs and of the (exposed) derivatives, and defines the
	unknowns that are available during Initialization [Therefore, when linearizing an FMU, every tool will use
	the same ordering for the outputs, states, and derivatives for the linearized model. The ordering of the
	inputs should be performed in this case according to the ordering in ModelVariables.] A ModelExchange
	FMU must expose all derivatives of its continuous-time states in element <Derivatives>.
	A Co-Simulation FMU does not need to expose these state derivatives. [If a Co-Simulation FMU exposes its
	state derivatives, they are usually not utilized for the co-simulation, but, for example, to linearize the FMU at a communication point.]

	The optional part defines in which way derivatives and outputs depend on inputs, and continuous-time
	states at the current super dense time instant (ModelExchange) or at the current Communication Point (CoSimulation).
	[A simulation environment can utilize this information to improve the efficiency, for
	example, when connecting FMUs together, or when computing the partial derivative of the derivatives
	with respect to the states in the simulation engine.]
*/
type ModelStructure struct {
	/*
		Outputs is an ordered list of all outputs, in other words a list of ScalarVariable indices
		where every corresponding ScalarVariable must have causality = "output"
		(and every variable with causality=”output” must be listed here).
		[Note that all output variables are listed here, especially discrete and
		continuous outputs. The ordering of the variables in this list is defined by the
		exporting tool. Usually, it is best to order according to the declaration order in the
		source model, since then the <Outputs> list does not change if the declaration
		order of outputs in the source model is not changed. This is, for example,
		important for linearization, in order that the interpretation of the output vector
		does not change for a re-exported FMU.]. Attribute dependencies defines the
		dependencies of the outputs from the knowns at the current super dense time
		instant in Event and in Continuous-Time Mode (ModelExchange) and at the
		current Communication Point (CoSimulation).
	*/
	Outputs []Unknown `xml:"Outputs>Unknown,omitempty"`

	/*
		Derivatives is an ordered list of all state derivatives, in other words, a list of ScalarVariable
		indices where every corresponding ScalarVariable must be a state
		derivative. [Note that only continuous Real variables are listed here. If a state or
		a derivative of a state shall not be exposed from the FMU, or if states are not
		statically associated with a variable (due to dynamic state selection), then
		dummy ScalarVariables have to be introduced.
		The ordering of the variables in this list is defined by the
		exporting tool. Usually, it is best to order according to the declaration order of the
		states in the source model, since then the <Derivatives> list does not change if
		the declaration order of states in the source model is not changed. This is, for
		example, important for linearization, in order that the interpretation of the state
		vector does not change for a re-exported FMU.]. The number of Unknown
		elements in the Derivatives element uniquely define the number of continuous
		time state variables, as required by the corresponding Model Exchange functions
		(integer argument nx of fmi2GetContinuousStates, fmi2SetcontinuousStates, fmi2GetDerivatives, fmi2GetNominalsOfContinuousStates) that require it.
		The corresponding continuous-time states are defined by attribute derivative of
		the corresponding ScalarVariable state derivative element. [Note that higher
		order derivatives must be mapped to first order derivatives but the mapping
		definition can be preserved due to attribute derivative.

		For Co-Simulation, element “Derivatives” is ignored if capability flag
		providesDirectionalDerivative has a value of false, in other words, it
		cannot be computed. [This is the default. If an FMU supports both
		ModelExchange and CoSimulation, then the “Derivatives” element might be
		present, since it is needed for ModelExchange. If the above flag is set to false for
		the CoSimulation case, then the “Derivatives” element is ignored for
		CoSimulation. If “inline integration” is used for a CoSimulation slave, then the
		model still has continuous-time states and just a special solver is used (internally
		the implementation results in a discrete-time system, but from the outside, it is
		still a continuous-time system).]
	*/
	Derivatives []Unknown `xml:"Derivatives>Unknown,omitempty"`

	/*
		InitialUnknowns is ordered list of all exposed Unknowns in Initialization Mode. This list consists of
		all variables with

		(1) causality = "output" and ( initial="approx" or "calculated" ), and

		(2) causality = "calculatedParameter" and

		(3) all continuous-time states and all state derivatives (defined with element
		<Derivatives> from <ModelStructure> ) with initial="approx" or
		"calculated" [if a Co-Simulation FMU does not define the
		<Derivatives> element, (3) cannot be present.].

		The resulting list is not allowed to have duplicates (for example, if a state is also
		an output, it is included only once in the list). The Unknowns in this list must be
		ordered according to their ScalarVariable index (for example, if for two variables
		A and B the ScalarVariable index of A is less than the index of B, then A must
		appear before B in InitialUnknowns ).

		Attribute dependencies defines the dependencies of the Unknowns from the
		Knowns in Initialization Mode at the initial time.
	*/
	InitialUnknowns []Unknown `xml:"InitialUnknowns>Unknown,omitempty"`
}

/*
	Unknown is dependency of scalar Unknown from Knowns in Continuous-Time and Event Mode (ModelExchange),
	and at Communication Points (CoSimulation): Unknown=f(Known_1, Known_2, ...).
	Knowns are "inputs", "continuous states" and "independent variable" (usually time).
*/
type Unknown struct {
	// Index is ScalarVariable index of Unknown
	Index uint `xml:"index,attr"`
	/*
		Dependencies attribute defining the dependencies of the unknown v unknown (directly or
		indirectly via auxiliary variables) with respect to v known . If not present, it must be
		assumed that the Unknown depends on all Knowns. If present as empty list, the
		Unknown depends on none of the Knowns. Otherwise the Unknown depends on
		the Knowns defined by the given ScalarVariable indices. The indices are ordered
		according to magnitude, starting with the smallest index.
	*/
	Dependencies UintAttributeList `xml:"dependencies,attr,omitempty"`

	/*
		DependenciesKind is an attribute list of type of dependencies in "dependencies" attribute.
		If not present, it must be assumed that the Unknown v unknown depends on the
		Knowns v known without a particular structure. Otherwise, the corresponding
		Known v known,i enters the equation as:

		If "dependenciesKind" is present, "dependencies" must be present and must
		have the same number of list elements

		- dependent: no particular structure

		Only for Real unknowns:

		- constant: constant factor

		Only for Real unknowns in event and continuous-time mode (model exchange) and at comms points (CoSimulation),
		and not for InitialUnknowns for Initialization Mode:

		- fixed: fixed factor

		- tunable: tunable factor

		- discrete: discrete factor
	*/
	DependenciesKind StringAttributeList `xml:"dependenciesKind,attr,omitempty"`
}

// UintAttributeList represents space delimited list of unsigned integers in an xml attribute
type UintAttributeList []uint

// StringAttributeList represents space delimited list of strings in an xml attribute
type StringAttributeList []string

func (l UintAttributeList) MarshalText() (text []byte, err error) {
	return []byte(strings.Trim(fmt.Sprintf("%v", l), "[]")), nil
}

func (l StringAttributeList) MarshalText() (text []byte, err error) {
	return []byte(strings.Trim(fmt.Sprintf("%v", l), "[]")), nil
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
