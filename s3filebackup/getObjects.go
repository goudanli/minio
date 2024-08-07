/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-07-30 10:46:07
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-08-03 10:24:24
 * @FilePath: /s3filebackup/getObjects.go
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
	// 修正destPath
	if destPath[len(destPath)-1] != '/' {
		destPath = destPath + "/"
	}
	// 指定要下载的对象和文件路径
	parts := strings.Split(dataPath, "/")
	println("bucketName:", parts[1])
	firstSlashIndex := strings.Index(dataPath[1:], "/") + 1
	objectName := dataPath[firstSlashIndex+1:]
	fmt.Println("objectName:", objectName)

	errChan := make(chan error)
	var wg sync.WaitGroup
	sem = make(chan struct{}, 1024)

	wg.Add(1)
	go func() {
		err := getObjects(minioClient, parts[1], objectName, destPath, &wg)
		errChan <- err
	}()
	wg.Wait()
	err = <-errChan
	return err
}

func getObjects(minioClient *minio.Client, bucketName string, objectName string, destPath string, wg *sync.WaitGroup) error {
	defer wg.Done()
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    objectName,
		Recursive: false,
	})
	for object := range objectCh {
		if hasError {
			break
		}
		if object.Err != nil {
			log.Fatalln(object.Err)
			return object.Err
		}
		key := object.Key
		// fmt.Println("object.Key:", key)
		if key[len(key)-1] == '/' {
			// fmt.Println("Folder:", key)
			folderName := key[len(objectName):]
			folderPath := destPath + folderName
			// fmt.Println("folderPath:", folderPath)
			err := os.MkdirAll(folderPath, 0755) // 0755 表示目录权限
			if err != nil {
				log.Fatalln(err)
				return err
			}
			wg.Add(1)
			go func() {
				err := getObjects(minioClient, bucketName, key, folderPath, wg)
				if err != nil {
					errChan <- err
					hasError = true
				}
			}()
		} else {
			wg.Add(1)
			sem <- struct{}{}
			// fmt.Println("File:", key)
			go func() {
				defer func() { <-sem }()
				err := getObject(minioClient, bucketName, key, destPath, wg)
				if err != nil {
					errChan <- err
					hasError = true
				}
			}()
		}
	}

	if hasError {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return nil
}

func getObject(minioClient *minio.Client, bucketName string, objectName string, destPath string, wg *sync.WaitGroup) error {
	defer wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 找到最后一个斜线的索引
	lastSlashIndex := strings.LastIndex(objectName, "/")

	// 如果找到了斜线并且不是最后一个字符
	result := objectName
	if lastSlashIndex != -1 && lastSlashIndex < len(objectName)-1 {
		// 截取最后一个斜线后面的字符串
		result = objectName[lastSlashIndex+1:]
	}

	// 打开文件以写入下载的对象数据
	filepath := destPath + result
	// fmt.Println(filepath)
	file, err := os.Create(filepath)
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
