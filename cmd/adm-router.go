/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-07 17:01:41
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-05-08 15:07:01
 * @FilePath: /minio/cmd/adm-router.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */

package cmd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/gzip"
	"github.com/minio/minio/internal/logger"
)

const (
	admAPIPathPrefix = "/minio/adm"
)

// registerHealthCheckRouter - add handler functions for liveness and readiness routes.
func registerADMAPIRouter(router *mux.Router) {
	// Healthcheck router
	admApiRouter := router.PathPrefix(SlashSeparator).Subrouter()

	// apiRouter := router.PathPrefix(SlashSeparator).Subrouter()
	gz, err := gzhttp.NewWrapper(gzhttp.MinSize(1000), gzhttp.CompressionLevel(gzip.BestSpeed))
	if err != nil {
		// Static params, so this is very unlikely.
		logger.Fatal(err, "Unable to initialize server")
	}
	//update adm backup job
	admApiRouter.Methods(http.MethodPost).Path("/updatebackupstatus").HandlerFunc(gz(httpTraceAll(UpdateBackupStatusHandler)))
	admApiRouter.Methods(http.MethodPost).Path("/updaterestorestatus").HandlerFunc(gz(httpTraceAll(UpdateRestoreStatusHandler)))
}
