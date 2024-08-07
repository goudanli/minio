###
 # @Author: xiao.wei xiaow@suninfo.com
 # @Date: 2024-08-01 10:52:27
 # @LastEditors: xiao.wei xiaow@suninfo.com
 # @LastEditTime: 2024-08-02 11:05:46
 # @FilePath: /s3filebackup/build.sh
 # @Description: 
 # 
 # Copyright (c) 2024 by suninfo, All Rights Reserved. 
### 

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.CommitID=$(git rev-parse HEAD) -X main.BranchName=$(git rev-parse --abbrev-ref HEAD)" -o s3filebackup
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-X main.CommitID=$(git rev-parse HEAD) -X main.BranchName=$(git rev-parse --abbrev-ref HEAD)" -o s3filebackup_arm