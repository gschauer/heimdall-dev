//  Copyright 2023 The heimdall-dev authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package cfg mocks a more dynamic config mgmt (external configs, API, etc.).
// For the the sake of simplicity, values are hardcoded.
package cfg

func GetArtifactRepoBase() string {
	return "examples/artifacts/zzz-raw-host"
}

func GetProjectKey() string {
	return "ZZZ"
}

func IsGitEnabled() bool {
	return false
}
