
###
 # @Author: xiao.wei xiaow@suninfo.com
 # @Date: 2024-12-11 15:43:31
 # @LastEditors: xiao.wei xiaow@suninfo.com
 # @LastEditTime: 2024-12-11 15:50:10
 # @FilePath: /adm_s3gateway/build.sh
 # @Description: 
 # 
 # Copyright (c) 2024 by suninfo, All Rights Reserved. 
### 
script_dir="$(cd "$(dirname "$0")" && pwd)"
# echo $script_dir
cd $script_dir
git pull
BranchName=`git rev-parse --abbrev-ref HEAD`

echo "build x86_64"
export GOARCH=amd64 && make build
mv minio cdm-s3gateway
scp cdm-s3gateway root@192.168.216.102:/opt/adm-project/adm-${BranchName}/s3gateway_bin/x86_64

echo "build aarch64"
export GOARCH=arm64 && make build
mv minio cdm-s3gateway
scp cdm-s3gateway root@192.168.216.102:/opt/adm-project/adm-${BranchName}/s3gateway_bin/aarch64

