// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package acceptance

import (
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/acceptance"
)

func init() {
	acceptance.DefineFlags()
}

func TestAcceptanceDotNet(t *testing.T) {
	builder, cleanup := acceptance.CreateBuilder(t)
	defer cleanup()

	sdk := filepath.Join("/layers", dotnetRuntime, "sdk")

	testCases := []acceptance.Test{
		{
			Name:              "simple dotnet app",
			App:               "dotnet/simple",
			MustUse:           []string{dotnetRuntime, dotnetPublish},
			FilesMustNotExist: []string{sdk},
		},
		{
			Name:              "simple dotnet 6.0 app",
			App:               "dotnet/simple_dotnet6",
			MustUse:           []string{dotnetRuntime, dotnetPublish},
			FilesMustNotExist: []string{sdk},
		},
		{
			Name:              "simple dotnet app with runtime version",
			App:               "dotnet/simple",
			Path:              "/version?want=3.1.1",
			Env:               []string{"GOOGLE_RUNTIME_VERSION=3.1.101"},
			MustUse:           []string{dotnetRuntime, dotnetPublish},
			FilesMustNotExist: []string{sdk},
		},
		{
			Name:              "simple dotnet app with global.json",
			App:               "dotnet/simple_with_global",
			Path:              "/version?want=3.1.0",
			MustUse:           []string{dotnetRuntime, dotnetPublish},
			FilesMustNotExist: []string{sdk},
		},
		{
			Name:              "simple prebuilt dotnet app",
			App:               "dotnet/simple_prebuilt",
			Env:               []string{"GOOGLE_ENTRYPOINT=./simple"},
			MustUse:           []string{dotnetRuntime},
			MustNotUse:        []string{dotnetPublish},
			FilesMustNotExist: []string{sdk},
		},
		{
			Name:                "Dev mode",
			App:                 "dotnet/simple",
			Env:                 []string{"GOOGLE_DEVMODE=1"},
			MustUse:             []string{dotnetRuntime, dotnetPublish},
			FilesMustExist:      []string{sdk, "/workspace/Startup.cs"},
			MustRebuildOnChange: "/workspace/Startup.cs",
		},
		{
			// This is a separate test case from Dev mode above because it has a fixed runtime version.
			// Its only purpose is to test that the metadata is set correctly.
			Name:    "Dev mode metadata",
			App:     "dotnet/simple",
			Env:     []string{"GOOGLE_DEVMODE=1", "GOOGLE_RUNTIME_VERSION=3.1.409"},
			MustUse: []string{dotnetRuntime, dotnetPublish},
			BOM: []acceptance.BOMEntry{
				{
					Name: "sdk",
					Metadata: map[string]interface{}{
						"version": "3.1.409",
					},
				},
				{
					Name: "devmode",
					Metadata: map[string]interface{}{
						"devmode.sync": []interface{}{
							map[string]interface{}{"dest": "/workspace", "src": "**/*.cs"},
							map[string]interface{}{"dest": "/workspace", "src": "*.csproj"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.fs"},
							map[string]interface{}{"dest": "/workspace", "src": "*.fsproj"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.vb"},
							map[string]interface{}{"dest": "/workspace", "src": "*.vbproj"},
							map[string]interface{}{"dest": "/workspace", "src": "**/*.resx"},
						},
					},
				},
			},
		},
		{
			Name:       "dotnet selected via GOOGLE_RUNTIME",
			App:        "override",
			Env:        []string{"GOOGLE_RUNTIME=dotnet"},
			MustUse:    []string{dotnetRuntime},
			MustNotUse: []string{nodeRuntime, pythonRuntime, goRuntime},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			acceptance.TestApp(t, builder, tc)
		})
	}
}

func TestFailuresDotNet(t *testing.T) {
	builder, cleanup := acceptance.CreateBuilder(t)
	t.Cleanup(cleanup)

	testCases := []acceptance.FailureTest{
		{
			Name:      "bad runtime version",
			App:       "dotnet/simple",
			Env:       []string{"GOOGLE_RUNTIME_VERSION=BAD_NEWS_BEARS"},
			MustMatch: "runtime version BAD_NEWS_BEARS does not exist",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			acceptance.TestBuildFailure(t, builder, tc)
		})
	}
}
