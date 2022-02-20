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
package godoc

import (
	"context"
	"fmt"
	"os/exec"

	"go.uber.org/zap"
)

// GodocOptions defines the options available for running the godoc
// service.
type GodocOptions struct {
	// The GOROOT value that will be passed to godoc.  Initially set
	// in the config.
	Goroot string
	// The port that godoc will run on. Initially set in the config.
	GodocPort int
	// The logger used by the godoc service. Initially set in the
	// config.
	Logger *zap.Logger
}

type Godoc struct {
	// The GodocOptions that was passed into New.
	options GodocOptions
	// The logger used by the godoc service.
	logger *zap.Logger
}

// New returns an initialized Godoc struct
func New(g GodocOptions) *Godoc {
	return &Godoc{
		options: g,
		logger:  g.Logger,
	}
}

// Start runs the godoc service.  The path of the godoc executable is looked
// up and the argument string created.  The godoc service is started and any
// errors returned to the caller.
func (g *Godoc) Start(ctx context.Context) error {
	godoc, err := exec.LookPath("godoc")
	if err != nil {
		g.logger.Error("unable to find godoc in the path")
		return err
	}

	arg := []string{
		fmt.Sprintf("-http=localhost:%d", g.options.GodocPort),
		fmt.Sprintf("-goroot=%s", g.options.Goroot),
		"-index",
	}
	// Godoc is required to be in the path.
	cmd := exec.CommandContext(ctx, godoc, arg...)
	err = cmd.Start()
	if err != nil {
		g.logger.Error("unable to start godoc server", zap.Error(err))
		return err
	}

	return cmd.Wait()
}
