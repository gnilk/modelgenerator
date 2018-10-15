package golang

import "modelgenerator/common"

type CrudGenerator struct {
	Imports []common.XMLImport
}

type DBGenerator struct{}

type CodeGenerator struct {
	Methods []common.AccessMethod
	Imports []common.XMLImport
}

type GoLangGenerators struct{}

func (lang *GoLangGenerators) GetModelGenerator() common.Generator {
	return createGoLangGenerator()
}

func (lang *GoLangGenerators) GetCrudGenerator() common.Generator {
	return createCrudGenerator()
}

func (lang *GoLangGenerators) GetDBCreateGenerator() common.Generator {
	return createDBGenerator()
}

func CreateGoLanguage() common.Language {
	golang := GoLangGenerators{}
	return (common.Language)(&golang)
}

func createGoLangGenerator() common.Generator {
	codeGen := CodeGenerator{}
	return (common.Generator)(&codeGen)
}

func createCrudGenerator() common.Generator {
	generator := CrudGenerator{}
	return (common.Generator)(&generator)
}

func createDBGenerator() common.Generator {
	generator := DBGenerator{}
	return (common.Generator)(&generator)
}
