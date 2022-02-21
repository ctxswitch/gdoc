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
package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// A personal access token with permissions to access and list the
	// repositories.
	GithubToken string `envconfig:"GITHUB_TOKEN" required:"true"`
	// The user who the token belongs to.  Defaults to the Github user.
	GithubTokenUser string `envconfig:"GITHUB_TOKEN_USER" default:""`
	// The Github user or organization that will be scraped.  Only single
	// values are currently supported.
	GithubUser string `envconfig:"GITHUB_USER" requrired:"true"`
	// The interval to check for changes on Github.  Takes a duration string
	// for the value.  The string is an unsigned decimal number(s), with
	// optional fraction and a unit suffix, such as "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m",
	// "h".
	GithubPollInterval string `envconfig:"GITHUB_POLL_INTERVAL" default:"5m"`
	// The topic that will be used as a filter to identify repositories
	// that will be synchronized.
	GithubTopic string `envconfig:"GITHUB_TOPIC" default:"godoc"`
	// The port that godoc will run on.
	GodocPort int `envconfig:"GODOC_PORT" default:"6060"`
	// The GOROOT value that will be passed to godoc.
	GodocRoot string `envconfig:"GODOC_ROOT" default:"/usr/local/go"`
	// The indexing interval for godoc.  0 for default (5m), negative
	// to only index once at startup.
	GodocIndexInterval string `envconfig:"GODOC_INDEX_INTERVAL" default:"1m"`
	// Changes the verbosity of the logging system.
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"`
}

// New returns a configuration that has been processed and defaulted.
func New() *Config {
	config := &Config{}
	_ = envconfig.Process("", config)

	if config.GithubTokenUser == "" {
		config.GithubTokenUser = config.GithubUser
	}

	return config
}
