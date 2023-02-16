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

package git

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gschauer/heimdall-dev/cfg"
	"github.com/gschauer/heimdall-dev/internal"
	"github.com/gschauer/heimdall-dev/plugin"
	"github.com/gschauer/heimdall-dev/release"
	"github.com/rs/zerolog/log"
)

var user, pass string

type CommitPlugin struct {
	branch  string
	commits []*object.Commit
}

func init() {
	if !cfg.IsGitEnabled() {
		return
	}

	if user = os.Getenv("GIT_USERNAME"); user == "" {
		log.Warn().Str("name", "GIT_USERNAME").Msg("undefined environment variable")
		return
	}
	if pass = os.Getenv("GIT_PASSWORD"); pass == "" {
		log.Warn().Str("name", "GIT_PASSWORD").Msg("undefined environment variable")
		return
	}

	plugin.Register(&CommitPlugin{})
}

func (p *CommitPlugin) Load(o, n release.Info) {
	old := map[string]string{}
	for _, c := range o.Components {
		url, rev, _ := strings.Cut(c, "@")
		old[url] = rev
	}

	for _, c := range n.Components {
		url, newRev, _ := strings.Cut(c, "@")
		if oldRev, ok := old[url]; ok && oldRev != newRev {
			r := internal.Must(open(url, user, pass))
			base := mergeBase(r, oldRev, newRev)
			p.commits = append(p.commits, p.loadCommits(r, base)...)
			findBranches(r, resolve(r, newRev)) // TODO: implement correct remote resolution without clone
			p.branch = newRev
		}
	}
}
func (p *CommitPlugin) InitEnv(env map[string]any) {
	env["git"] = map[string]any{
		"branch":         p.branch,
		"commits":        p.commits,
		"validCommitMsg": validCommitMsg,
	}
}

func validCommitMsg(c *object.Commit) bool {
	return strings.HasPrefix(c.Message, cfg.GetProjectKey()+"-")
}

func open(uri, user, pass string) (*git.Repository, error) {
	if strings.HasPrefix(uri, "https://") {
		log.Info().Str("URL", uri).Msg("Cloning Git repo")
		return git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:  uri,
			Auth: &http.BasicAuth{Username: user, Password: pass},
		})
	}

	log.Info().Str("dir", uri).Msg("Opening Git repo")
	return git.PlainOpen(uri)
}

func mergeBase(r *git.Repository, revs ...string) *object.Commit {
	hashes := make([]plumbing.Hash, len(revs))
	for _, rev := range revs {
		hashes = append(hashes, resolve(r, rev))
	}

	commits := make([]*object.Commit, len(hashes))
	for _, hash := range hashes {
		commits = append(commits, internal.Must(r.CommitObject(hash)))
	}

	if len(commits) != 2 {
		internal.MustOkMsgf[any](nil, false, "expected 2 commits, got %d", len(commits))
	}
	bs := internal.Must(commits[0].MergeBase(commits[1]))
	if len(bs) != 1 {
		internal.MustOkMsgf[any](nil, false, "expected 1 common ancestor, got %d", len(bs))
	}
	log.Debug().Stringer("old", commits[0].Hash).Stringer("new", commits[1].Hash).
		Stringer("hash", bs[0].Hash).Msg("Resolved Git merge base")
	return bs[0]
}

func resolve(r *git.Repository, rev string) plumbing.Hash {
	if hash, err := r.ResolveRevision(plumbing.Revision("refs/heads/" + rev)); err == nil {
		return *hash
	} else if hash, err = r.ResolveRevision(plumbing.Revision("refs/tags/" + rev)); err == nil {
		return *hash
	} else if hash, err = r.ResolveRevision(plumbing.Revision("refs/remotes/origin/" + rev)); err == nil {
		return *hash
	} else {
		internal.MustOkMsgf[any](nil, false, "cannot resolve revision %s", rev)
	}
	return plumbing.ZeroHash
}

func (p *CommitPlugin) loadCommits(r *git.Repository, base *object.Commit) (cs []*object.Commit) {
	log.Debug().Str("repository", internal.Must(r.Remote("origin")).Config().URLs[0]).Msg("Loading commits")
	l := internal.Must(r.Log(&git.LogOptions{From: base.Hash}))
	defer l.Close()
	internal.MustNoErr(l.ForEach(func(c *object.Commit) error {
		cs = append(cs, c)
		return nil
	}))
	return
}

func findBranches(r *git.Repository, c plumbing.Hash) (bs []plumbing.ReferenceName) {
	log.Debug().Stringer("hash", c).Msg("Resolve branch(es) for commit")
	visited := make(map[plumbing.Hash]bool)
	refs := listRemoteRefs(r)

	for _, ref := range refs {
		n := ref.Name()
		if n.IsBranch() {
			internal.MustNoErr(r.Fetch(&git.FetchOptions{
				RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
				Auth:     &http.BasicAuth{Username: user, Password: pass},
			}))

			b := internal.Must(r.Reference(n, true))
			ok := internal.Must(reaches(r, b.Hash(), c, visited))
			if ok {
				bs = append(bs, n)
			}
		}
	}
	return
}

func listRemoteRefs(r *git.Repository) (bs []*plumbing.Reference) {
	rem := internal.Must(r.Remote("origin"))
	refs := internal.Must(rem.List(&git.ListOptions{
		Auth: &http.BasicAuth{Username: user, Password: pass},
	}))

	refPrefix := "refs/heads/"
	for _, ref := range refs {
		if !strings.HasPrefix(ref.Name().String(), refPrefix) {
			continue
		}
		bs = append(bs, ref)
	}
	return
}

func reaches(r *git.Repository, start, c plumbing.Hash, visited map[plumbing.Hash]bool) (bool, error) {
	if v, ok := visited[start]; ok {
		return v, nil
	}
	if start == c {
		visited[start] = true
		return true, nil
	}

	co, err := r.CommitObject(start)
	if err != nil {
		return false, err
	}

	for _, p := range co.ParentHashes {
		v, err := reaches(r, p, c, visited)
		if err != nil {
			return false, err
		}
		if v {
			visited[start] = true
			return true, nil
		}
	}

	visited[start] = false
	return false, nil
}
