
###
 # @Author: xiao.wei xiaow@suninfo.com
 # @Date: 2024-07-31 11:20:53
 # @LastEditors: xiao.wei xiaow@suninfo.com
 # @LastEditTime: 2024-08-02 09:20:46
 # @FilePath: /s3filebackup/get_file_from_cdm.sh
 # @Description: 
 # 
 # Copyright (c) 2024 by suninfo, All Rights Reserved. 
### 

restore_path=""
data_path=""
cdm_server=""
while getopts ":c:r:d:" opt
do
    case $opt in
        c)
        cdm_server=$OPTARG
        ;;
        r)
        restore_path=$OPTARG
        ;;
        d)
        data_path=$OPTARG
        ;;
        ?)
        echo "未知参数"
        exit 1;;
        :)         
        echo "没有输入任何选项 $OPTARG";;
esac done

if [ -z "$cdm_server" ]; then
  echo "Option -c is required."
  exit 1
fi

if [ -z "$restore_path" ]; then
  echo "Option -r is required."
  exit 1
fi

if [ -z "$data_path" ]; then
  echo "Option -d is required."
  exit 1
fi

[ ! -d $restore_path ] &&  mkdir -p $restore_path

script_dir="$(cd "$(dirname "$0")" && pwd)"

recover_exe_path=${script_dir}/../../../s3filebackup
result=(`${recover_exe_path} -server ${cdm_server} -data ${data_path} -dest ${restore_path}`)
if [ $? -ne 0 ];then
  exit $?
fi

chmod 777 -R ${restore_path}