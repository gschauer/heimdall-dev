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

package res

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gschauer/heimdall-dev/internal"
)

var client http.Client

func init() {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/"))) //nolint:gosec
	client = http.Client{Transport: t}
}

func Open(uri string) (io.ReadCloser, error) {
	if strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://") {
		resp, err := client.Get(uri)
		if err != nil {
			return nil, err
		}
		return resp.Body, err
	}
	return os.Open(uri)
}

func CompRev(uri string) (string, string) {
	n, rev, _ := strings.Cut(uri, "@")
	n = path.Base(n)
	return strings.TrimSuffix(n, ".git"), rev
}

func ToMap(a any) (m map[string]any) {
	bs := internal.Must(json.Marshal(a))
	internal.MustNoErr(json.Unmarshal(bs, &m))
	return
}
