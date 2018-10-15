package cpp

import "modelgenerator/common"

type CodeGenerator struct {
	Methods []common.AccessMethod
	Imports []common.XMLImport
}

type CppLangGenerators struct{}

func CreateCppLanguage() common.Language {
	cpplang := CppLangGenerators{}
	return (common.Language)(&cpplang)
}

func (lang *CppLangGenerators) GetModelGenerator() common.Generator {
	return createCppLangGenerator()
}

func (lang *CppLangGenerators) GetCrudGenerator() common.Generator {
	return nil
}

func (lang *CppLangGenerators) GetDBCreateGenerator() common.Generator {
	return nil
}

func createCppLangGenerator() common.Generator {
	codeGen := CodeGenerator{}
	return (common.Generator)(&codeGen)
}
