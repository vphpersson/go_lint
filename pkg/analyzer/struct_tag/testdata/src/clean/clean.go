package clean

type Clean struct {
	A string `json:",omitempty"`
	B *int   `json:"b,omitzero"`
	C string `json:"-"`
	D string `json:"name"`
	E string `yaml:",omitempty"`
	F string `xml:"name,omitempty"`
	G string
	H string `json:"omitemptyish"`
	I string `db:"omitempty"`
}
