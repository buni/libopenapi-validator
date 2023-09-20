// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package requests

import (
	"net/http"
	"sync"

	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RequestBodyValidator is an interface that defines the methods for validating request bodies for Operations.
//
//	ValidateRequestBody method accepts an *http.Request and returns true if validation passed,
//	                    false if validation failed and a slice of ValidationError pointers.
type RequestBodyValidator interface {
	// ValidateRequestBody will validate the request body for an operation. The first return value will be true if the
	// request body is valid, false if it is not. The second return value will be a slice of ValidationError pointers if
	// the body is not valid.
	ValidateRequestBody(request *http.Request) (bool, []*errors.ValidationError)

	// SetPathItem will set the pathItem for the RequestBodyValidator, all validations will be performed
	// against this pathItem otherwise if not set, each validation will perform a lookup for the pathItem
	// based on the *http.Request
	SetPathItem(path *v3.PathItem, pathValue string)
}

// NewRequestBodyValidator will create a new RequestBodyValidator from an OpenAPI 3+ document
func NewRequestBodyValidator(document *v3.Document) RequestBodyValidator {
	validator := &requestBodyValidator{document: document, schemaCache: make(map[[32]byte]*schemaCache), mux: &sync.RWMutex{}}
	_ = validator.buildCache() // ignore the error for now

	return validator
}

func (v *requestBodyValidator) SetPathItem(path *v3.PathItem, pathValue string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.pathItem = path
	v.pathValue = pathValue
}

type schemaCache struct {
	schema         *base.Schema
	renderedInline []byte
	renderedJSON   []byte
}

type requestBodyValidator struct {
	document    *v3.Document
	pathItem    *v3.PathItem
	pathValue   string
	errors      []*errors.ValidationError
	schemaCache map[[32]byte]*schemaCache
	mux         *sync.RWMutex
}
