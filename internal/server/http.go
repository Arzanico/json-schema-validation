package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"net/http"
	"strings"
)

var schemaText = `
{
  "$id": "https://example.com/product.schema.json",
  "title": "Product",
  "description": "A product from Acme's catalog",
  "type": "object",
  "properties": {
    "productId": {
      "description": "The unique identifier for a product",
      "type": "integer"
    },
    "productName": {
      "description": "Name of the product",
      "type": "string"
    },
    "price": {
      "description": "The price of the product",
      "type": "number",
      "exclusiveMinimum": 0
    },
    "tags": {
      "description": "Tags for the product",
      "type": "array",
      "items": {
        "type": "string"
      },
      "minItems": 1,
      "uniqueItems": true
    }
  },
  "required": [ "productId", "productName", "price" ]
}
`

func NewHttpServer(addr string) *http.Server {
	httpsrv := newHttpServer()
	r := mux.NewRouter()
	r.HandleFunc("/validate", httpsrv.validate).Methods(http.MethodPost)
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type httpServer struct {
	Payload *PayloadValidationRequest
}

func newHttpServer() *httpServer {
	return &httpServer{
		Payload: NewPayloadValidationRequest(),
	}
}

func (s *httpServer) validate(w http.ResponseWriter, r *http.Request) {
	var m interface{}
	err := json.NewDecoder(r.Body).Decode(m)
	if err != nil {
		panic(err)
	}

	compiler := jsonschema.NewCompiler()
	//compiler.Draft = jsonschema.Draft2019

	if err := compiler.AddResource("schema.json", strings.NewReader(schemaText)); err != nil {
		panic(err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		panic(err)
	}
	error := schema.Validate(r.Body)
	fmt.Println(error)
	//err := json.NewEncoder(w).Encode(resp)
	return
}
