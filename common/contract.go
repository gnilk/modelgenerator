package common

//
// Options for generator, takem from command line
//
type Options struct {
	SplitInFiles          bool
	DBTablePrefix         string
	Converters            bool
	Verbose               int
	Filename              string
	OutputName            string
	OutputDBName          string
	AllPersistenceClasses []string
	PersistenceClass      string
	DoPersistence         bool
	IsUpgrade             bool
	GenerateDropStatement bool
	GettersAndSetters     bool
	FromVersion           int    // Always assume from version 0
	DocumentRootDirectory string // This is set by code to the root directory of the first document, relative for all includes
	UseLanguage           string
	Language              Language
	CurrentDoc            *XMLDoc
}

type Generator interface {
	GenerateCode(doc XMLDoc, options *Options) string
}
type Language interface {
	GetModelGenerator() Generator
	GetCrudGenerator() Generator
	GetDBCreateGenerator() Generator
}
