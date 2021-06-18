// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package internal

// This directory contains the OpenAPI 3.0 specification which defines our
// server. The file petstore.gen.go is automatically generated from the schema

// Run oapi-codegen to regenerate the petstore boilerplate
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --package=petstore --generate types,client -o ../petstore-client.gen.go ../petstore-expanded.yaml
