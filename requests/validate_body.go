// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package requests

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/utils"
)

func (v *requestBodyValidator) ValidateRequestBody(request *http.Request) (bool, []*errors.ValidationError) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	// find path

	pathItem, errs, _ := paths.FindPath(request, v.document)
	if pathItem == nil || errs != nil {
		return false, errs
	}

	var validationErrors []*errors.ValidationError
	operation := helpers.ExtractOperation(request, pathItem)

	var contentType string
	// extract the content type from the request

	if contentType = request.Header.Get(helpers.ContentTypeHeader); contentType != "" {

		// extract the media type from the content type header.
		ct, _, _ := helpers.ExtractContentType(contentType)
		if operation.RequestBody != nil {
			if mediaType, ok := operation.RequestBody.Content[ct]; ok {
				// we currently only support JSON validation for request bodies
				// this will capture *everything* that contains some form of 'json' in the content type
				if strings.Contains(strings.ToLower(contentType), helpers.JSONType) {
					// extract schema from media type
					if mediaType.Schema != nil {

						var schema *base.Schema
						var renderedInline, renderedJSON []byte

						// have we seen this schema before? let's hash it and check the cache.
						hash := mediaType.GoLow().Schema.Value.Hash()

						if cacheHit, ch := v.schemaCache[hash]; ch { // might wanna error here if not found, but if buildCache is called, this should never happen
							schema = cacheHit.schema
							renderedInline = cacheHit.renderedInline
							renderedJSON = cacheHit.renderedJSON
						}

						// render the schema, to be used for validation
						valid, vErrs := ValidateRequestSchema(request, schema, renderedInline, renderedJSON)
						if !valid {
							validationErrors = append(validationErrors, vErrs...)
						}
					}
				}
			} else {
				// content type not found in the contract
				validationErrors = append(validationErrors, errors.RequestContentTypeNotFound(operation, request))
			}
		}
	}
	if len(validationErrors) > 0 {
		return false, validationErrors
	}
	return true, nil
}

func (v *requestBodyValidator) buildCache() error {
	v.mux.Lock()
	defer v.mux.Unlock()
	if v.document.Paths == nil {
		return nil
	}

	for _, pathItem := range v.document.Paths.PathItems {
		// build cache for each path item
		for _, operation := range extractOperations(pathItem) {
			for _, body := range operation.RequestBody.Content {
				// build cache for each media type
				if body.Schema != nil {
					// build cache for each schema

					// have we seen this schema before? let's hash it and check the cache.

					hash := body.GoLow().Schema.Value.Hash()

					// perform work only once and cache the result in the validator.

					// render the schema inline and perform the intensive work of rendering and converting
					// this is only performed once per schema and cached in the validator.
					schema := body.Schema.Schema()
					renderedInline, err := schema.RenderInline()
					if err != nil {
						return fmt.Errorf("failed to render inline schema: %w", err)
					}

					renderedJSON, err := utils.ConvertYAMLtoJSON(renderedInline)
					if err != nil {
						return fmt.Errorf("failed to convert rendered inline schema to JSON: %w", err)
					}

					v.schemaCache[hash] = &schemaCache{
						schema:         schema,
						renderedInline: renderedInline,
						renderedJSON:   renderedJSON,
					}

				}
			}
		}
	}

	return nil
}

func extractOperations(pathItem *v3.PathItem) []*v3.Operation {
	operations := make([]*v3.Operation, 0, 8)

	if pathItem.Get != nil && pathItem.Get.RequestBody != nil && pathItem.Get.RequestBody.Content != nil {
		operations = append(operations, pathItem.Get)
	}

	if pathItem.Put != nil && pathItem.Put.RequestBody != nil && pathItem.Put.RequestBody.Content != nil {
		operations = append(operations, pathItem.Put)
	}

	if pathItem.Post != nil && pathItem.Post.RequestBody != nil && pathItem.Post.RequestBody.Content != nil {
		operations = append(operations, pathItem.Post)
	}

	if pathItem.Delete != nil && pathItem.Delete.RequestBody != nil && pathItem.Delete.RequestBody.Content != nil {
		operations = append(operations, pathItem.Delete)
	}

	if pathItem.Options != nil && pathItem.Options.RequestBody != nil && pathItem.Options.RequestBody.Content != nil {
		operations = append(operations, pathItem.Options)
	}

	if pathItem.Head != nil && pathItem.Head.RequestBody != nil && pathItem.Head.RequestBody.Content != nil {
		operations = append(operations, pathItem.Head)
	}

	if pathItem.Patch != nil && pathItem.Patch.RequestBody != nil && pathItem.Patch.RequestBody.Content != nil {
		operations = append(operations, pathItem.Patch)
	}

	if pathItem.Trace != nil && pathItem.Trace.RequestBody != nil && pathItem.Trace.RequestBody.Content != nil {
		operations = append(operations, pathItem.Trace)
	}

	return operations
}
