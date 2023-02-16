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

package jira

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/gschauer/heimdall-dev/res"
	"github.com/rs/zerolog/log"
)

type IssuePlugin struct {
	client *jira.Client
	rel    *release.Info
}

func (p *IssuePlugin) Load(o, n release.Info) {
	p.rel = &n
	is := ListIssues(p.client, os.Args[1])
	j := internal.Must(json.Marshal(is))

	d := filepath.Join(cfg.GetArtifactRepoBase(), "jira", n.Release, "issues", time.Now().Format("2006-01-02T15:04:05"), "issues.json")
	internal.MustNoErr(os.MkdirAll(filepath.Dir(d), 0700))
	internal.MustNoErr(os.WriteFile(d, j, 0600))
}

func (p *IssuePlugin) InitEnv(env map[string]any) {
	d := filepath.Join(cfg.GetArtifactRepoBase(), "jira", p.rel.Release, "issues")
	dirs := internal.Must(os.ReadDir(d))
	date := dirs[0].Name() // TODO: pick correct date

	f := internal.Must(res.Open(filepath.Join(d, date, "issues.json")))
	defer func() { _ = f.Close() }()

	var ir IssueRecord
	internal.MustNoErr(json.NewDecoder(f).Decode(&ir))
	env["jira"] = map[string]any{
		"issues": ir.Issues,
	}
}

func init() {
	baseURL := os.Getenv("JIRA_BASE_URL")
	if baseURL == "" {
		log.Warn().Str("name", "JIRA_BASE_URL").Msg("undefined environment variable")
		return
	}

	token := os.Getenv("JIRA_TOKEN")
	if token == "" {
		log.Warn().Str("name", "JIRA_TOKEN").Msg("undefined environment variable")
		return
	}

	client, _ := NewClient(baseURL, token)
	p := IssuePlugin{client, nil}
	plugin.Register(&p)
}

type Issue struct {
	Key     string `json:"key"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

func (i Issue) String() string {
	return i.Key
}

type IssueRecord struct {
	JQL    string  `json:"jql"`
	Issues []Issue `json:"issues"`
}

// NewClient creates a new Jira client.
func NewClient(url, token string) (*jira.Client, error) {
	tp := jira.PATAuthTransport{Token: token}
	return jira.NewClient(tp.Client(), url)
}

// ListIssues searches for Jira issues matching the given JQL query.
func ListIssues(client *jira.Client, jql string) []Issue {
	is, _, _ := client.Issue.Search(jql, nil)
	r := make([]Issue, len(is))
	for k, i := range is {
		r[k] = Issue{i.Key, i.Fields.Type.Name, i.Fields.Summary, i.Fields.Status.Name}
	}
	return r
}
