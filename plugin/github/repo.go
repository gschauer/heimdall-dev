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

package github

import (
	"context"
	"os"
	"strings"

	"github.com/google/go-github/v49/github"
	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/gschauer/heimdall-dev/res"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type RepoInfo struct {
	ID            int64                                `json:"ID"`
	Name          string                               `json:"name"`
	MasterBranch  string                               `json:"master_branch,omitempty"`
	DefaultBranch string                               `json:"default_branch,omitempty"`
	GitURL        string                               `json:"git_url,omitempty"`
	GitCloneURL   string                               `json:"git_clone_url,omitempty"`
	Sec           github.SecurityAndAnalysis           `json:"security_and_analysis"`
	PRRules       github.PullRequestReviewsEnforcement `json:"required_pull_request_reviews"`
}

type RepoPlugin struct {
	client    *github.Client
	repoInfos []RepoInfo
}

func (p *RepoPlugin) Load(o, n release.Info) {
	for _, c := range n.Components {
		c, _, _ = strings.Cut(c, "@")
		c = strings.TrimSuffix(c, ".git")
		ps := strings.Split(c, "/")
		p.repoInfos = append(p.repoInfos, p.GetRepoInfo(ps[len(ps)-2], ps[len(ps)-1]))
	}
}

func (p *RepoPlugin) InitEnv(env map[string]any) {
	env["github"] = map[string]any{"repos": res.ToMap(p.repoInfos[0])}
}

func (p *RepoPlugin) GetRepoInfo(owner, repo string) RepoInfo {
	r, _, err := p.client.Repositories.Get(context.Background(), owner, repo)
	internal.MustNoErr(err)

	prEnf, _, err := p.client.Repositories.GetPullRequestReviewEnforcement(context.Background(), owner, repo, r.GetDefaultBranch())
	internal.MustNoErr(err)

	return RepoInfo{
		r.GetID(),
		r.GetName(),
		r.GetMasterBranch(),
		r.GetDefaultBranch(),
		r.GetGitURL(),
		r.GetCloneURL(),
		*r.GetSecurityAndAnalysis(),
		*prEnf,
	}
}

func init() {
	if !cfg.IsGitEnabled() {
		return
	}

	t := os.Getenv("GITHUB_TOKEN")
	if t == "" {
		log.Warn().Str("name", "GITHUB_TOKEN").Msg("undefined environment variable")
		return
	}
	url := os.Getenv("GITHUB_API_URL")
	if url == "" {
		log.Warn().Str("name", "GITHUB_API_URL").Msg("undefined environment variable")
		return
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t})
	httpClient := oauth2.NewClient(context.Background(), src)
	client := internal.Must(github.NewEnterpriseClient(url+"/v3/", url+"/uploads/", httpClient))

	plugin.Register(&RepoPlugin{client, nil})
}
