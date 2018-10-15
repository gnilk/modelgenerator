BIN = $(GOPATH)/bin
MODELGEN = $(BIN)/modelgenerator
GOIMPORTS = $(BIN)/goimports


GENERATOR_FILES = modelgenerator.go
#GENERATOR_FILES += generators/golang/golangmodelgenerator.go

MODEL_SRC = sample_datamodel.xml
MODEL_OUT = sample.go

all:	generator


generator: 	$(GENERATOR_FILES)
	go build -o $(MODELGEN) -i $(GENERATOR_FILES)


test:	generator $(MODEL_SRC)
	$(MODELGEN) -v -p - -c $(MODEL_SRC) -o $(MODEL_OUT)
	$(GOIMPORTS) -w $(MODEL_OUT)

