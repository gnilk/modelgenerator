BIN = $(GOPATH)/bin
MODELGEN = $(BIN)/modelgenerator

GENERATOR_FILES = modelgenerator.go genmodel.go genpersistence.go gendbcreatecode.go

MODEL_SRC = sample_datamodel.xml
MODEL_OUT = sample.go

all:	generator


generator: 	$(GENERATOR_FILES)
	go build -o $(MODELGEN) -i $(GENERATOR_FILES)


test:	generator $(MODEL_SRC)
	$(MODELGEN) -v -p -c $(MODEL_SRC) -o $(MODEL_OUT)


