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
package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ctxswitch/godoc-web/internal/config"
	"github.com/ctxswitch/godoc-web/internal/godoc"
	"github.com/ctxswitch/godoc-web/internal/logger"
	"github.com/ctxswitch/godoc-web/internal/syncer"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	logger := logger.New(cfg.LogLevel)

	logger.Info("Starting up godoc service")
	logger.Debug("Using configuration", zap.Any("config", cfg))

	var wg sync.WaitGroup

	gsync := syncer.New(syncer.SyncerOptions{
		GithubToken:        cfg.GithubToken,
		GithubTokenUser:    cfg.GithubTokenUser,
		GithubUser:         cfg.GithubUser,
		GithubTopic:        cfg.GithubTopic,
		GithubPollInterval: cfg.GithubPollInterval,
		GoRoot:             cfg.GoRoot,
		Logger:             logger,
	})

	godoc := godoc.New(godoc.GodocOptions{
		Goroot:    cfg.GoRoot,
		GodocPort: cfg.GodocPort,
		Logger:    logger,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		logger.Info("starting the syncer service")
		err := gsync.Start(ctx)
		logger.Error("syncer exited", zap.Error(err))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		logger.Info("starting the godoc service")
		err := godoc.Start(ctx)
		logger.Error("godoc exited", zap.Error(err))
	}()

	wg.Wait()
}
