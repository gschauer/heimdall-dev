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
	"encoding/csv"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/gschauer/heimdall-dev/res"
)

type CovRec struct {
	Group     string `json:"group"`
	Pkg       string `json:"pkg"`
	Class     string `json:"class"`
	InstrMis  int    `json:"instruction_missed"`
	InstrCov  int    `json:"instruction_covered"`
	BranchMis int    `json:"branch_missed"`
	BranchCov int    `json:"branch_covered"`
	LineMis   int    `json:"line_missed"`
	LineCov   int    `json:"line_covered"`
	ComplMis  int    `json:"complexity_missed"`
	ComplCov  int    `json:"complexity_covered"`
	MethMis   int    `json:"method_missed"`
	MethCov   int    `json:"method_covered"`
}

type JaCoCoPlugin struct {
	rel release.Info
}

func (p *JaCoCoPlugin) Load(o, n release.Info) {
	p.rel = n
}

func (p *JaCoCoPlugin) InitEnv(env map[string]any) {
	for _, c := range p.rel.Components {
		n, _ := res.CompRev(c)
		tag := p.rel.Release
		crs := LoadCovCSV(filepath.Join(cfg.GetArtifactRepoBase(), n, tag, "reports", "jacoco", "test", "jacocoTestReport.csv"))
		env["jacoco"] = res.ToMap(Aggregate(crs...))
	}
}

func LoadCovCSV(uri string) (crs []CovRec) {
	f := internal.Must(res.Open(uri))
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	for i, rec := range internal.Must(r.ReadAll()) {
		if strings.Contains(rec[1], ".generated") || i == 0 {
			continue
		}

		crs = append(crs, CovRec{
			Group:     rec[0],
			Pkg:       rec[1],
			Class:     rec[2],
			InstrMis:  internal.Must(strconv.Atoi(rec[3])),
			InstrCov:  internal.Must(strconv.Atoi(rec[4])),
			BranchMis: internal.Must(strconv.Atoi(rec[5])),
			BranchCov: internal.Must(strconv.Atoi(rec[6])),
			LineMis:   internal.Must(strconv.Atoi(rec[7])),
			LineCov:   internal.Must(strconv.Atoi(rec[8])),
			ComplMis:  internal.Must(strconv.Atoi(rec[9])),
			ComplCov:  internal.Must(strconv.Atoi(rec[10])),
			MethMis:   internal.Must(strconv.Atoi(rec[11])),
			MethCov:   internal.Must(strconv.Atoi(rec[12])),
		})
	}
	return
}

func Aggregate(crs ...CovRec) (tot CovRec) {
	for _, cr := range crs {
		tot.InstrMis += cr.InstrMis
		tot.InstrCov += cr.InstrCov
		tot.BranchMis += cr.BranchMis
		tot.BranchCov += cr.BranchCov
		tot.LineMis += cr.LineMis
		tot.LineCov += cr.LineCov
		tot.ComplMis += cr.ComplMis
		tot.ComplCov += cr.ComplCov
		tot.MethMis += cr.MethMis
		tot.MethCov += cr.MethCov
	}
	return
}

func init() {
	plugin.Register(&JaCoCoPlugin{})
}
