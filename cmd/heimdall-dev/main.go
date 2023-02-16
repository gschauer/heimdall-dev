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

package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	stdlog "log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/asaskevich/govalidator"
	"github.com/gschauer/heimdall-dev"
	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	_ "github.com/gschauer/heimdall-dev/plugin/git"
	_ "github.com/gschauer/heimdall-dev/plugin/github"
	_ "github.com/gschauer/heimdall-dev/plugin/java"
	_ "github.com/gschauer/heimdall-dev/plugin/jira"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/gschauer/heimdall-dev/res"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var version = 0

func main() {
	os.Args = []string{
		os.Args[0],
		"examples/releases/ZZZ_1.3.yml",
		"examples/releases/ZZZ_1.4.yml",
		"examples/checks.yml",
	}
	if len(os.Args) != 4 {
		stdlog.Fatalf("Usage: %s OLD_RELEASE NEW_RELEASE CHECKS\n", filepath.Base(os.Args[0]))
	}

	oldRel, newRel := loadYAML[release.Info](os.Args[1]), loadYAML[release.Info](os.Args[2])
	envMap := map[string]any{
		"releases": map[string]any{
			"old": res.ToMap(oldRel),
			"new": res.ToMap(newRel),
		},
		"println": fmt.Println,
		"split":   strings.Split,
	}

	for _, p := range plugin.Registry {
		p.Load(oldRel, newRel)
		p.InitEnv(envMap)
	}

	runChecks("examples/checks_without_git.yml", envMap)
}

func runChecks(file string, envMap map[string]any) {
	if stat, err := os.Stat(file); err != nil || !stat.Mode().IsRegular() {
		return
	}

	env := expr.Env(envMap)
	cfg := loadYAML[release.Config](file)

	if cfg.Cond == "" {
		// nothing to do
	} else if ok := internal.Must(expr.Eval(cfg.Cond, envMap)); reflect.ValueOf(ok).Kind() != reflect.Bool {
		log.Fatal().Str("cond", cfg.Cond).Interface("res", ok).Msg("Cannot evaluate condition")
	} else if !reflect.ValueOf(ok).Bool() {
		log.Info().Str("file", file).Msg("Skipping file")
		return
	}

	var cs []release.Check
	for _, s := range cfg.Steps {
		if s.Import != "" {
			fs := internal.Must(filepath.Glob(s.Import))
			if s.Import == "-" {
				fs = append(fs, os.Args[3])
			}
			for _, f := range fs {
				log.Info().Str("file", f).Msg("Importing")
				runChecks(f, envMap)
			}
			continue
		}

		fmt.Println("----")
		log.Info().Str("check", s.Name).Msg("Running")

		for _, c := range strings.Split(s.Cond, "\n") {
			if len(c) == 0 || strings.HasPrefix(c, "#") || strings.HasPrefix(c, "//") {
				continue
			}

			if strings.Contains(c, "valid:") {
				ts := strings.SplitN(c, " ", 3)
				val, exp := ts[0], ts[2]
				m := map[string]any{val: envMap[val]}
				r, err := govalidator.ValidateMap(m, map[string]any{val: exp})
				log.WithLevel(toLevel(r)).Str("val", val).Str("rule", exp).Bool("result", r).AnErr("error", err).Msg("Validating")
				internal.MustNoErr(err)
				cs = append(cs, release.Check{
					Name:      s.Name,
					Status:    release.ToStatus(r),
					Reference: "",
					Comment:   strconv.FormatBool(r),
				})
			} else if len(c) > 0 {
				prg := internal.Must(expr.Compile(c, env))
				r, err := expr.Run(prg, envMap)
				log.WithLevel(toLevel(r)).Str("cond", c).Interface("result", r).AnErr("error", err).Msg("Evaluating")
				cs = append(cs, release.Check{
					Name:      s.Name,
					Status:    release.ToStatus(r),
					Reference: "",
					Comment:   fmt.Sprint(r),
				})
			}
		}
	}

	p := filepath.Join("examples", "report.html")
	f := internal.Must(os.Create(p))
	defer func() { _ = f.Close() }()

	internal.MustNoErr(writeReport(f, cs))
	log.Info().Str("path", p).Msg("Wrote report")
}

func toLevel(ok any) zerolog.Level {
	v := reflect.ValueOf(ok)
	if (v.Kind() == reflect.Bool && v.Bool()) || !v.IsZero() {
		return zerolog.DebugLevel
	}
	return zerolog.WarnLevel
}

func loadYAML[T any](uri string) (i T) {
	log.Debug().Str("file", uri).Msg("Loading YAML")
	r := internal.Must(res.Open(uri))
	defer func() { _ = r.Close() }()
	internal.MustNoErr(yaml.NewDecoder(r).Decode(&i))
	return
}

// writeReport applies the HTML template to variables, including
//   - all environment variables
//   - built-in variables such as DATE and HEIMDALL_VERSION
//   - checks: slices of all checks
func writeReport(w io.Writer, cs []release.Check) error {
	repTmplText := internal.Must(fs.ReadFile(heimdall.StaticFS, "plugin/report/template.html"))
	tmpl := template.New("template.html")
	tmpl = template.Must(tmpl.Parse(string(repTmplText)))

	data := make(map[string]any)
	for _, v := range os.Environ() {
		k, v, _ := strings.Cut(v, "=")
		data[k] = v
	}
	data["HEIMDALL_VERSION"] = version
	data["PRODUCT_NAME"] = cfg.GetProjectKey()
	data["DATE"] = time.Now().Format(time.RFC3339)
	data["checks"] = cs

	return tmpl.Execute(w, data)
}
