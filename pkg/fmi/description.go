package fmi

import (
	"encoding/xml"
	"fmt"
	"time"
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
}

type modelDescriptionStatic struct {
	XMLName    xml.Name `xml:"fmiModelDescription"`
	FMIVersion string   `xml:"fmiVersion,attr"`
}

// Experiment element for model description default experiment
type Experiment struct {
	StartTime float64 `xml:"startTime,attr,omitempty"`
	StopTime  float64 `xml:"stopTime,attr,omitempty"`
	Tolerance float64 `xml:"tolerance,attr,omitempty"`
	StepSize  float64 `xml:"stepSize,attr,omitempty"`
}

type DateTime time.Time

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
	Name           string `xml:"name,attr"`
	ValueReference uint   `xml:"valueReference,attr"`
	Description    string `xml:"description,attr,omitempty"`
	Causality      string `xml:"causality,attr"`
	Initial        string `xml:"initial,attr"`
}
