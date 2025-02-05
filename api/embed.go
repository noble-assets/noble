// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"embed"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

//go:embed gen
var SwaggerUI embed.FS

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router, swaggerEnabled bool) error {
	if !swaggerEnabled {
		return nil
	}

	index, err := SwaggerUI.ReadFile("gen/index.html")
	if err != nil {
		return err
	}
	rtr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(index)
	})

	swagger, err := SwaggerUI.ReadFile("gen/swagger.yaml")
	if err != nil {
		return err
	}
	rtr.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(swagger)
	})

	return nil
}
