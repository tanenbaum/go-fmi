package fmi

import "encoding/xml"

type ModelDescription struct {
	XMLName xml.Name `xml:"fmiModelDescription"`
	Version string   `xml:"fmiVersion,attr"`
	Name    string   `xml:"modelName,attr"`
	GUID    string   `xml:"guid,attr"`
}

func NewModelDescription() ModelDescription {
	return ModelDescription{
		Version: GetVersion(),
	}
}
