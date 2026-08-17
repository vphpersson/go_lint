package flagged

type Flagged struct {
	A string `json:"omitempty"`             // want `json tag name "omitempty" is an option word; the field is serialized under the name "omitempty", not omitted; write ",omitempty" to apply the option`
	B string `json:"omitzero"`              // want `json tag name "omitzero" is an option word`
	C string `yaml:"omitempty"`             // want `yaml tag name "omitempty" is an option word`
	D string `xml:"omitempty"`              // want `xml tag name "omitempty" is an option word`
	E string `json:"omitempty,omitzero"`    // want `json tag name "omitempty" is an option word`
	F string `json:"omitempty" yaml:"omitempty"` // want `json tag name "omitempty" is an option word` `yaml tag name "omitempty" is an option word`
}
