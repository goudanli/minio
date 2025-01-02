/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-07-30 10:46:07
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2025-01-02 14:17:35
 * @FilePath: /adm_s3gateway/s3filebackup/getObjects.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var hasError bool
var sem chan struct{}

func getAllObjects(ip, port string, dataPath string, destPath string) error {
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

	// 指定要下载的对象和文件路径
	parts := strings.Split(dataPath, "/")
	println("bucketName:", parts[1])
	firstSlashIndex := strings.Index(dataPath[1:], "/") + 1
	prefix := dataPath[firstSlashIndex+1:]
	fmt.Println("prefix:", prefix)

	concurrencyLimit := 10
	errChan := make(chan error, concurrencyLimit+10)
	var wg sync.WaitGroup
	sem = make(chan struct{}, concurrencyLimit)

	getObjects(minioClient, parts[1], prefix, destPath, &wg, errChan)
	wg.Wait()
	close(errChan)
	for err := range errChan {
		fmt.Println("Error occurred:", err)
		return err
	}
	return nil
}

func getObjects(minioClient *minio.Client, bucketName string, prefix string, destPath string, wg *sync.WaitGroup, errChan chan<- error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var mu sync.Mutex
	for object := range objectCh {
		if hasError {
			break
		}
		if object.Err != nil {
			mu.Lock()
			errChan <- object.Err
			hasError = true
			mu.Unlock()
			return
		}
		key := object.Key
		if key[len(key)-1] != '/' {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				err := getObject(minioClient, bucketName, prefix, key, destPath, wg)
				if err != nil {
					mu.Lock()
					errChan <- err
					hasError = true
					mu.Unlock()
				}
			}()
		}
	}
}

func getObject(minioClient *minio.Client, bucketName string, prefix string, objectName string, destPath string, wg *sync.WaitGroup) error {
	if hasError {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tmpName := objectName[len(prefix):]
	fileName := destPath + tmpName
	// fmt.Println("TrimSuffix:", filepath.Dir(tmpName))
	folderName := destPath + filepath.Dir(tmpName)
	// 创建文件目录
	// fmt.Println("folderName:", folderName)
	err := os.MkdirAll(folderName, os.ModePerm)
	if err != nil {
		fmt.Printf("unable to create directory %v, %v", destPath, err)
		return err
	}

	// fmt.Println(filepath)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer file.Close()

	// 使用GetObject方法下载对象
	object, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer object.Close()

	// 将对象内容写入文件
	if _, err := io.Copy(file, object); err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}
