package validator

import (
	"bytes"
	"log"
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
)

func TestReusingValidatorRace(t *testing.T) {
	doc, err := libopenapi.NewDocument(pestore)
	assert.NoError(t, err)

	docv3, _ := doc.BuildV3Model()
	v := NewValidatorFromV3Model(&docv3.Model)

	for i := 0; i < 2; i++ {
		t.Run("ValidateHttpRequest", func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 10; i++ {
				req, _ := http.NewRequest("POST", "/pets", bytes.NewReader([]byte("{}")))
				req.Header.Set("Content-Type", "application/json")

				_, _ = v.ValidateHttpRequest(req)
			}
		})
	}
}

func TestReusingValidator(t *testing.T) {
	doc, err := libopenapi.NewDocument(pestore)
	assert.NoError(t, err)

	docv3, _ := doc.BuildV3Model()
	v := NewValidatorFromV3Model(&docv3.Model)

	t.Run("ValidateHttpRequest", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/pets", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		ok, errsz := v.ValidateHttpRequest(req)
		log.Println(ok, errsz[0].SchemaValidationErrors[0].OriginalError.Causes[0].Message)
	})
}

func TestValidatorRace(t *testing.T) {
	doc, err := libopenapi.NewDocument(pestore)
	assert.NoError(t, err)

	docv3, _ := doc.BuildV3Model()

	for i := 0; i < 2; i++ {
		v := NewValidatorFromV3Model(&docv3.Model)
		t.Run("ValidateHttpRequest", func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 10; i++ {
				req, _ := http.NewRequest("POST", "/pets", bytes.NewReader([]byte("{}")))
				req.Header.Set("Content-Type", "application/json")

				_, _ = v.ValidateHttpRequest(req)
			}
		})
	}
}

var pestore = []byte(`openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    post:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      requestBody: 
        content: 
          application/json:
            schema:
              type: object
              properties:
                id:
                  type: integer
                  format: int64
                  required: 
                    - id
              required: 
                - id
      responses:
        '200':
          description: A paged array of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:    
              schema:
                $ref: "#/components/schemas/Pets"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
        created_at:
          type: string
          format: date-time
          readOnly: true
    Pets:
      type: array
      maxItems: 100
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string`)
