package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
	"io"
	server "json-schema-validation/internal/server/canonicalSpecification"
	"json-schema-validation/lib/tkt"
	"net/http"
)

func NewHttpServer(addr string) *http.Server {
	validator := newCanonicalJsonValidator()
	r := mux.NewRouter()
	r.HandleFunc("/validate", validator.validate).Methods(http.MethodPost)
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type CanonicalJsonValidator struct {
	Document *string
	Schema   []byte
}

//THIS MAY BE PART OF THE SCHEMA, CAUSE IS THE VALIDATION CAPABILITY
func newCanonicalJsonValidator() *CanonicalJsonValidator {
	schema := server.InitializeCanonicalSchema()
	return &CanonicalJsonValidator{
		Schema: schema.Schema,
	}
}

func (v *CanonicalJsonValidator) validate(w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	//buf := bytes.Buffer{}
	//err = json.Compact(&buf, data)
	//tkt.CheckErr(err)

	//jsonDoc := buf.String()

	//schemaLoader := gojsonschema.NewStringLoader(v.Schema)
	//documentLoader := gojsonschema.NewStringLoader(jsonDoc)
	schemaLoader := gojsonschema.NewBytesLoader(v.Schema)
	documentLoader := gojsonschema.NewBytesLoader(data)

	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err)
	}

	result, err := schema.Validate(documentLoader)
	if err != nil {
		panic(err)
	}

	fmt.Println("Validation succesfully")
	fmt.Printf("Results:\n%v", result.Errors())
	tkt.JsonResponse(result, w)
	return
}
