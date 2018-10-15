package common

//
// Meta info for structure access methods (getter/setters)
//
type AccessMethod struct {
	Getter    bool
	Setter    bool
	IsList    bool
	IsPointer bool
	NoPersist bool
	AutoID    bool
	Name      string
	Type      string
}
