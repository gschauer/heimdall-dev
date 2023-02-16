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

package java

import (
	"path/filepath"
	"strings"

	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/gschauer/heimdall-dev/res"
	"github.com/joshdk/go-junit"
	"github.com/rs/zerolog/log"
)

type JUnitPlugin struct {
	suite *junit.Suite
}

func (p *JUnitPlugin) Load(o, n release.Info) {
	c, _ := res.CompRev(n.Components[0])
	path := filepath.Join(cfg.GetArtifactRepoBase(), c, n.Release, "test-results", "test")
	p.suite = &junit.Suite{Suites: loadSuites(path)}
	p.suite.Aggregate()
}

func (p *JUnitPlugin) InitEnv(env map[string]any) {
	env["junit"] = res.ToMap(p.suite.Totals)
}

// loadSuites recursively loads all files in the given directory, recursively.
// If a globing pattern is used, then it ingests only the matching files.
func loadSuites(dir string) []junit.Suite {
	if !strings.ContainsRune(dir, '*') {
		ss, err := junit.IngestDir(dir)
		if err != nil {
			log.Warn().Str("dir", dir).Msg("No JUnit suites found")
		}
		return ss
	}
	fs := internal.Must(filepath.Glob(dir))
	return internal.Must(junit.IngestFiles(fs))
}

func init() {
	plugin.Register(&JUnitPlugin{})
}
