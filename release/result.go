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

import (
	"fmt"
	"reflect"
)

type Check struct {
	Name      string
	Status    Status
	Reference string
	Comment   string
}

type Status string

const (
	OK     Status = "OK"
	Warn   Status = "Warn"
	Failed Status = "Failed"
)

func ToStatus(ok any) Status {
	v := reflect.ValueOf(ok)
	fmt.Println("STATUS", ok, v.Kind(), v.IsZero())
	if (v.Kind() == reflect.Bool && v.Bool()) || (v.Kind() != reflect.Bool && v.IsZero()) {
		return OK
	}
	return Failed
}
