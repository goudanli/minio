/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-02-20 16:20:40
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-12-31 14:51:09
 * @FilePath: /adm_s3gateway/s3filebackup/main.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil)) // 启动 pprof HTTP 服务器
	}()
	result := 0
	start := time.Now()
	serverPtr := flag.String("server", "", "请输入CDM Ip地址")
	portPtr := flag.Int("port", 44086, "请输入端口号")
	// /zpool231/data/testsystem/
	dataPathPtr := flag.String("data", "", "请输入数据路径")
	// /code/testsystem/
	destPathPtr := flag.String("dest", "", "请输入目标路径")
	typePtr := flag.Int("type", 0, "请输入操作类型（0.下载 1.上传）")
	flag.Parse()

	if *serverPtr == "" || *dataPathPtr == "" || *destPathPtr == "" {
		fmt.Printf("version: %s (commit-id=%s)\n", BranchName, CommitID)
		fmt.Println("缺少必需参数：")
		flag.Usage()
		os.Exit(1)
	}

	ip := *serverPtr
	port := *portPtr
	dataPath := *dataPathPtr
	destPath := *destPathPtr
	opType := *typePtr
	if opType == 0 {
		// todo检查参数
		err := getAllObjects(ip, strconv.Itoa(port), dataPath, destPath)
		if err != nil {
			log.Fatalln(err)
			result = 1
		}
	} else if opType == 1 {
		err := putAllObjects(ip, strconv.Itoa(port), destPath, dataPath)
		if err != nil {
			log.Fatalln(err)
			result = 1
		}
	} else {
		fmt.Printf("version: %s (commit-id=%s)\n", BranchName, CommitID)
		fmt.Println("type参数错误")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("程序已退出")
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("程序运行时间：%s\n", duration)
	os.Exit(result)
}
