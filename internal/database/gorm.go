/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 11:02:24
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-05-08 16:55:36
 * @FilePath: /minio/internal/database/gorm.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package database

import (
	"github.com/minio/minio/internal/config"
	"gorm.io/gorm"
)

type DataBase struct {
	Mysql config.Mysql
}

func Gorm(dbConfig *DataBase) *gorm.DB {
	return GormMysql(dbConfig)
}
