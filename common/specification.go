package common

//
// XML Import structures
//
type XMLDBControl struct {
	Host     string `xml:"host"`
	DBName   string `xml:"dbname"`
	Schema   string `xml:"schema"`
	User     string `xml:"user"`
	Password string `xml:"password"`
}

// XMLDataTypeField declares a variable in an definition (see XmlDefine)
type XMLDataTypeField struct {
	Name            string `xml:"name,attr"`
	Default         string `xml:"default,attr"`
	Value           int    `xml:"value,attr"`
	DBSize          int    `xml:"dbsize,attr"`
	FieldSize       int    `xml:"fieldsize,attr"`
	Type            string `xml:"type,attr"`
	IsPointer       bool   `xml:"ispointer,attr"`
	IsList          bool   `xml:"islist,attr"`
	FromVersion     int    `xml:"fromversion,attr"`
	SkipPersistance bool   `xml:"nopersist,attr"`
	DBAutoID        bool   `xml:"dbautoid,attr"`
	XMLAttrib       string `xml:"xmlattrib,attr"`
}

// XMLDefine declares an object (type/struct)
type XMLDefine struct {
	Type            string             `xml:"type,attr"`
	Name            string             `xml:"name,attr"`
	Inherits        string             `xml:"inherits,attr"`
	DBSchema        string             `xml:"dbschema,attr"`
	Prefix          string             `xml:"prefix,attr"`
	SkipPersistance bool               `xml:"nopersist,attr"`
	Fields          []XMLDataTypeField `xml:"field"`
	Guids           []XMLDataTypeField `xml:"guid"`
	Strings         []XMLDataTypeField `xml:"string"`
	Bools           []XMLDataTypeField `xml:"bool"`
	Times           []XMLDataTypeField `xml:"time"`
	Ints            []XMLDataTypeField `xml:"int"`
	Lists           []XMLDataTypeField `xml:"list"`
	Objects         []XMLDataTypeField `xml:"object"`
	Enums           []XMLDataTypeField `xml:"enum"`

	// private stuff
	Methods []AccessMethod
}

// XMLImport holds import directives
type XMLImport struct {
	DisablePersistence bool   `xml:"no_persistence,attr"`
	Package            string `xml:",innerxml"`
}

// XMLTypeMapping Holds type mappings definitions
type XMLTypeMapping struct {
	Lang      string `xml:"lang,attr"`
	FromType  string `xml:"from,attr"`
	ToType    string `xml:"to,attr"`
	Encode    string `xml:"encode,attr"`
	Decode    string `xml:"decode,attr"`
	FieldSize int    `xml:"fieldsize,attr"`
}

type XMLInclude struct {
	Filename string `xml:",innerxml"`
	Document XMLDoc
}

// XMLDoc holds the document root
type XMLDoc struct {
	Namespace       string           `xml:"namespace,attr"`
	DBSchema        string           `xml:"dbschema,attr"`
	Includes        []XMLInclude     `xml:"include"`
	Imports         []XMLImport      `xml:"imports>package"`
	Defines         []XMLDefine      `xml:"define"`
	DBTypeMappings  []XMLTypeMapping `xml:"dbtypemappings>map"`
	GOTypeMappings  []XMLTypeMapping `xml:"gotypemappings>map"`
	AnyTypeMappings []XMLTypeMapping `xml:"anytypemappings>map"`
	DBControl       XMLDBControl     `xml:"dbcontrol"`
}
