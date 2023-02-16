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

package release

type Config struct {
	Cond  string `json:"cond" yaml:"cond"`
	Steps []Step `json:"steps" yaml:"steps"`
}

type Step struct {
	Name   string `json:"name" yaml:"name"`
	Desc   string `json:"description" yaml:"description"`
	Cond   string `json:"condition" yaml:"condition"`
	Out    string `json:"output" yaml:"output"`
	Type   string `json:"type" yaml:"type"`
	Import string `json:"import" yaml:"import"`
}
