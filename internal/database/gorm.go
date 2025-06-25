/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 11:02:24
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2025-06-24 17:09:02
 * @FilePath: /adm_s3gateway/internal/database/gorm.go
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
	DBType string `mapstructure:"db-type" json:"db-type" yaml:"db-type"` // 数据库类型
	Mysql  config.Mysql
	Dm     config.DM
}

func Gorm(dbConfig *DataBase) *gorm.DB {
	if dbConfig.DBType == "mysql" {
		return GormMysql(dbConfig)
	}
	if dbConfig.DBType == "dm" {
		return GormDM(dbConfig)
	}
	panic("unsupported database type: " + dbConfig.DBType)
}
