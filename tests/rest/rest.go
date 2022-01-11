/*
Copyright © 2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

const (
	apiURL            = "http://localhost:8081/"
	contentTypeHeader = "Content-Type"
)

// ServerTests run all tests for basic REST API endpoints
func ServerTests() {
	// tests for OpenAPI specification that is accessible via its endpoint as well
	// implementation of these tests is stored in openapi.go
	checkOpenAPISpecifications()
}
