// Copyright (C) 2022, Rob Lyon <rob@ctxswitch.com>
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package syncer

import (
	"context"
	"fmt"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v42/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// SyncerOptions defines the options available for running the
// Syncer service.
type SyncerOptions struct {
	// A personal access token with permissions to access and list the
	// repositories.  Initially set in the config.
	GithubToken string
	// The user who the token belongs to.  Defaults to the Github user.
	// Initially set in the config.
	GithubTokenUser string
	// The Github user or organization that will be scraped.  Only single
	// values are currently supported.  Initially set in the config.
	GithubUser string
	// The topic that will be used as a filter to identify repositories
	// that will be synchronized.  Initially set in the config.
	GithubTopic string
	// The interval to check for changes on Github.  Takes a duration string
	// for the value.  The string is an unsigned decimal number(s), with
	// optional fraction and a unit suffix, such as "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m",
	// "h".  Initially set in the config.
	GithubPollInterval string
	// Changes the verbosity of the logging system.  Initially set in the config.
	GoRoot string
	// The logger used by the godoc service. Initially set in the
	// config.
	Logger *zap.Logger
}

// Syncer is a service that polls Github looking for repositories that have been
// tagged with a specific topic as defined for GithubTopic.  The list of
// repositories is returned and the latest commit sha is gathered.  If a repo
// does not exist locally, it is cloned using the username and token and if the
// repo exists and has been updated as seen by comparing the commit sha, the
// changes are pulled in.
type Syncer struct {
	options SyncerOptions
	repos   map[string]*Repo
	logger  *zap.Logger
}

func New(options SyncerOptions) *Syncer {
	return &Syncer{
		options: options,
		repos:   make(map[string]*Repo),
		logger:  options.Logger,
	}
}

// Start runs the synchronization process.  The process is repeated at an interval
// equal to the configured poll interval.
func (rs *Syncer) Start(ctx context.Context) error {
	// BUG(d) Negative values are not checked before the poll interval is passed
	// to the ParseDuration function.
	// BUG(d) Small values should not be allowed.  We need to set a minimun value
	// of potentially 1 minute.  If not, we run the risk of exceeding limits which
	// could cause the service to behave poorly.
	d, err := time.ParseDuration(rs.options.GithubPollInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rs.sync(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

// update checks to see if the repository has changed since the last
// cycle.
func (rs *Syncer) update(r *Repo) bool {
	name := r.Name + "/" + r.Owner
	if _, has := rs.repos[name]; !has {
		// We've not seen the repo before.  Add it and return true
		// signifying that we've seen a change.
		rs.repos[name] = r
		return true
	}

	if r.CommitSHA == rs.repos[name].CommitSHA {
		// The repo has not changed.
		return false
	}

	// The commit shas are different.  Replace the repo and return that
	// we have changed.
	rs.repos[name] = r
	return true
}

// sync utilizes the Github API though a personal access token and
// queries for repositories that have a configured topic set.  Once
// the list has returned, it iterates through and gathers the latest
// commit sha by getting detailed information about the default branch.
// If there has been an update to the repository, the local repo is
// updated.
func (rs *Syncer) sync(ctx context.Context) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: rs.options.GithubToken},
	)
	auth := oauth2.NewClient(ctx, ts)
	client := github.NewClient(auth)

	q := fmt.Sprintf("language:go user:%s topic:%s", rs.options.GithubUser, rs.options.GithubTopic)
	rs.logger.Debug("query string", zap.String("query", q))

	result, _, err := client.Search.Repositories(ctx, q, &github.SearchOptions{})
	if err != nil {
		rs.logger.Error("search failed", zap.Error(err))
		return
	}
	rs.logger.Debug("search", zap.Int("total", *result.Total))

	for _, repo := range result.Repositories {
		r := &Repo{
			Owner:     *repo.Owner.Login,
			Name:      *repo.Name,
			CloneURL:  *repo.CloneURL,
			LocalPath: fmt.Sprintf("%s/src/github.com/%s", rs.options.GoRoot, *repo.FullName),
		}

		branch, _, err := client.Repositories.GetBranch(ctx, r.Owner, r.Name, *repo.DefaultBranch, true)
		if err != nil {
			rs.logger.Error("unable to get commit", zap.Error(err))
			continue
		}

		r.CommitSHA = *branch.Commit.SHA
		if changed := rs.update(r); !changed {
			rs.logger.Debug("repository has not changed", zap.Any("repo", r), zap.Any("sha", branch.Commit.SHA))
			continue
		}

		rs.logger.Info("processing repository update", zap.Any("repo", r), zap.Any("sha", branch.Commit.SHA))
		if err = rs.get(r); err != nil {
			rs.logger.Error("unable to update repository", zap.Error(err))
		}
	}
}

// get determines whether or not a repository has already been cloned.  If it
// does not yet exist, it is cloned.  Otherwise a pull is performed.
func (rs *Syncer) get(r *Repo) error {
	// if the path already exists and is a git repo, then pull otherwise clone
	if _, err := os.Stat(r.LocalPath); os.IsNotExist(err) {
		return rs.clone(r)
	} else {
		return rs.pull(r)
	}
}

// clone performs a git clone of the repository passed to is as an argument
// using token based authentication.
func (rs *Syncer) clone(r *Repo) error {
	rs.logger.Info("cloning repository", zap.Any("repo", r))
	_, err := git.PlainClone(r.LocalPath, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: rs.options.GithubTokenUser,
			Password: rs.options.GithubToken,
		},
		URL:      r.CloneURL,
		Progress: nil,
	})

	return err
}

// pull performs a git pull of the provided repository
func (rs *Syncer) pull(r *Repo) error {
	rs.logger.Info("pulling repository", zap.Any("repo", r))
	p, err := git.PlainOpen(r.LocalPath)
	if err != nil {
		return err
	}

	w, err := p.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: rs.options.GithubTokenUser,
			Password: rs.options.GithubToken,
		},
		RemoteName: "origin",
		Depth:      1,
	})
	return err
}
