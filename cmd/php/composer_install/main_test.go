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

package main

import (
	"fmt"
	"net/http"
	"testing"

	bpt "github.com/GoogleCloudPlatform/buildpacks/internal/buildpacktest"
	"github.com/GoogleCloudPlatform/buildpacks/internal/testserver"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "with composer.json",
			files: map[string]string{
				"index.php":     "",
				"composer.json": "",
			},
			want: 0,
		},
		{
			name: "without composer.json",
			files: map[string]string{
				"index.php": "",
			},
			want: 100,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bpt.TestDetect(t, detectFn, tc.name, tc.files, []string{}, tc.want)
		})
	}
}

func TestBuild(t *testing.T) {
	var (
		expectedHash    = "expected_sha384_hash"
		actualHashCmd   = "php -r"
		runInstallerCmd = fmt.Sprintf("php %s/%s", composerLayer, composerSetup)
	)

	testCases := []struct {
		name                string
		opts                []bpt.Option
		wantExitCode        int // 0 if unspecified
		wantCommands        []string
		skippedCommands     []string
		httpStatusInstaller int
		httpStatusSignature int
	}{
		{
			name: "good composer-setup hash with installation",
			opts: []bpt.Option{
				bpt.WithExecMock("sha384", bpt.MockStdout(expectedHash)),
				bpt.WithExecMock("--filename composer", bpt.MockExitCode(0)),
			},
			wantCommands: []string{
				actualHashCmd,
				runInstallerCmd,
			},
		},
		{
			name: "bad composer-setup hash with installation",
			opts: []bpt.Option{
				bpt.WithExecMock("sha384", bpt.MockStdout("actual_sha384_hash")),
				bpt.WithExecMock("--filename composer", bpt.MockExitCode(0)),
			},
			wantCommands: []string{
				actualHashCmd,
			},
			skippedCommands: []string{
				runInstallerCmd,
			},
			wantExitCode: 1,
		},
		{
			name: "unable to download composer-setup",
			opts: []bpt.Option{
				bpt.WithExecMock("sha384", bpt.MockStdout(expectedHash)),
				bpt.WithExecMock("--filename composer", bpt.MockExitCode(0)),
			},
			skippedCommands: []string{
				actualHashCmd,
				runInstallerCmd,
			},
			httpStatusInstaller: http.StatusInternalServerError,
			wantExitCode:        1,
		},
		{
			name: "unable to get expected hash",
			opts: []bpt.Option{
				bpt.WithExecMock("sha384", bpt.MockStdout(expectedHash)),
				bpt.WithExecMock("--filename composer", bpt.MockExitCode(0)),
			},
			skippedCommands: []string{
				actualHashCmd,
				runInstallerCmd,
			},
			httpStatusSignature: http.StatusInternalServerError,
			wantExitCode:        1,
		},
		{
			name: "unable to get actual hash",
			opts: []bpt.Option{
				bpt.WithExecMock("sha384", bpt.MockExitCode(1)),
				bpt.WithExecMock("--filename composer", bpt.MockExitCode(0)),
			},
			wantCommands: []string{
				actualHashCmd,
			},
			skippedCommands: []string{
				runInstallerCmd,
			},
			wantExitCode: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// stub the installer
			testserver.New(
				t,
				testserver.WithStatus(tc.httpStatusInstaller),
				testserver.WithJSON(`test_file_content`),
				testserver.WithMockURL(&composerSetupURL),
			)

			// stub the signature
			testserver.New(
				t,
				testserver.WithStatus(tc.httpStatusSignature),
				testserver.WithJSON(expectedHash),
				testserver.WithMockURL(&composerSigURL),
			)

			opts := []bpt.Option{
				bpt.WithTestName(tc.name),
			}

			opts = append(opts, tc.opts...)
			result, err := bpt.RunBuild(t, buildFn, opts...)
			if err != nil && tc.wantExitCode == 0 {
				t.Fatalf("error running build: %v, result: %#v", err, result)
			}

			if result.ExitCode != tc.wantExitCode {
				t.Errorf("build exit code mismatch, got: %d, want: %d", result.ExitCode, tc.wantExitCode)
			}

			for _, cmd := range tc.wantCommands {
				if !result.CommandExecuted(cmd) {
					t.Errorf("expected command %q to be executed, but it was not", cmd)
				}
			}

			for _, cmd := range tc.skippedCommands {
				if result.CommandExecuted(cmd) {
					t.Errorf("expected command %q to not be executed, but it was", cmd)
				}
			}
		})
	}
}
