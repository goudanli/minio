/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-07 17:01:41
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-06-03 15:07:23
 * @FilePath: /minio/cmd/adm-router.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */

package cmd

import (
	"net/http"
	"time"

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
	mapTimer = make(map[string]*time.Timer)
	admApiRouter := router.PathPrefix(SlashSeparator).Subrouter()

	// apiRouter := router.PathPrefix(SlashSeparator).Subrouter()
	gz, err := gzhttp.NewWrapper(gzhttp.MinSize(1000), gzhttp.CompressionLevel(gzip.BestSpeed))
	if err != nil {
		// Static params, so this is very unlikely.
		logger.Fatal(err, "Unable to initialize server")
	}
	// 更新状态，内部会移除心跳timer
	admApiRouter.Methods(http.MethodPost).Path("/updateRecordStatus").HandlerFunc(gz(httpTraceAll(UpdateRecordStatusHandler)))
	// 查看状态
	admApiRouter.Methods(http.MethodGet).Path("/progress").HandlerFunc(gz(httpTraceAll(ProgressHandler)))
	// 添加心跳timer或更新
	admApiRouter.Methods(http.MethodPost).Path("/addHeartBeatTimer").HandlerFunc(gz(httpTraceAll(AddHeartBeatTimerHandler)))
	// 更新心跳timer
	admApiRouter.Methods(http.MethodPost).Path("/updateHeartBeatTimer").HandlerFunc(gz(httpTraceAll(UpdateHeartBeatTimerHandler)))

	admApiRouter.Methods(http.MethodPost).Path("/updateStatistics").HandlerFunc(gz(httpTraceAll(UpdateStatisticsHandler)))
}
