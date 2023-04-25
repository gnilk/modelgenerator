BIN = $(GOPATH)/bin
MODELGEN = modelgenerator
GOIMPORTS = $(BIN)/goimports


GENERATOR_FILES = modelgenerator.go
#GENERATOR_FILES += generators/golang/golangmodelgenerator.go

MODEL_SRC = sample_datamodel.xml
MODEL_OUT = sample.go
CPP_MODEL_OUT = sample.cpp

all:	generator


generator: 	$(GENERATOR_FILES)
	go build -o $(MODELGEN) $(GENERATOR_FILES)


test:	generator $(MODEL_SRC)
	$(MODELGEN) -v -p - -c $(MODEL_SRC) -o $(MODEL_OUT)
	$(GOIMPORTS) -w $(MODEL_OUT)

cpp: generator $(MODEL_SRC)
	$(MODELGEN) -v -l cpp -m ! -c $(MODEL_SRC) -o $(CPP_MODEL_OUT)

clean:
	rm $(MODEL_OUT)
	rm $(CPP_MODEL_OUT)