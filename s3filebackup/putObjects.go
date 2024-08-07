/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-08-01 17:16:12
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-08-03 11:01:52
 * @FilePath: /s3filebackup/putObjects.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func putAllObjects(ip, port string, destPath string, dataPath string) error {
	hasError = false
	endpoint := ip + ":" + port
	accessKeyID := "root"
	secretAccessKey := "suninfo@123"
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
		},
	})
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// /zpool231/data/testsystem/
	parts := strings.Split(destPath, "/")
	println("bucketName:", parts[1])
	firstSlashIndex := strings.Index(destPath[1:], "/") + 1
	// data/testsystem/
	objectName := destPath[firstSlashIndex+1:]
	// 修正objectName
	if objectName[len(objectName)-1] != '/' {
		objectName = objectName + "/"
	}
	fmt.Println("objectName:", objectName)

	errChan := make(chan error)
	var wg sync.WaitGroup
	sem = make(chan struct{}, 10240)

	wg.Add(1)
	go func() {
		err := putObjects(minioClient, parts[1], objectName, dataPath, &wg)
		errChan <- err
	}()
	wg.Wait()
	err = <-errChan
	return err
}

func putObjects(minioClient *minio.Client, bucketName string, objectName string, dataPath string, wg *sync.WaitGroup) error {
	defer wg.Done()
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// fmt.Println("path:", path)
		if info.IsDir() {
			if path != dataPath {
				// 如果是目录，则上传目录作为一个特殊的对象
				relativePath := strings.TrimPrefix(path, dataPath)
				objKey := objectName + relativePath + "/"
				// fmt.Println("dir:", objKey)
				_, err := minioClient.PutObject(ctx, bucketName, objKey, strings.NewReader(""), 0, minio.PutObjectOptions{})
				if err != nil {
					return err
				}
			}
		} else {
			wg.Add(1)
			sem <- struct{}{}
			relativePath := strings.TrimPrefix(path, dataPath)
			if path == dataPath {
				relativePath = filepath.Base(path)
			}
			objKey := objectName + relativePath
			// fmt.Println("file:", objKey)
			go func() {
				defer func() { <-sem }()
				err := putObject(minioClient, bucketName, objKey, path, wg)
				if err != nil {
					errChan <- err
					hasError = true
				}
			}()
		}
		return nil
	})
	if hasError {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return err
}

func putObject(minioClient *minio.Client, bucketName string, objectName string, filePath string, wg *sync.WaitGroup) error {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()

	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, file, size, minio.PutObjectOptions{})
	return err
}
