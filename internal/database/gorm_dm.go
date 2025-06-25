/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 10:43:35
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2025-06-25 09:58:21
 * @FilePath: /adm_s3gateway/internal/database/gorm_dm.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package database

import (
	"errors"
	"log"

	"github.com/minio/minio/internal/dm"
	"gorm.io/gorm"
)

// GormMysql 初始化Mysql数据库

func GormDM(dbConfig *DataBase) *gorm.DB {
	m := dbConfig.Dm
	if m.Dbname == "" {
		log.Fatal(errors.New("please set dbname"))
	}
	if db, err := gorm.Open(dm.Open(m.Dsn())); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		// sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		// sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		if err := sqlDB.Ping(); err != nil {
			log.Fatal(err)
		}
		// 执行 ALTER SESSION 语句切换模式
		err = db.Exec("ALTER SESSION SET CURRENT_SCHEMA = " + m.Dbname).Error
		if err != nil {
			panic(err)
		}
		log.Println("Successfully connected to DM database")
		return db
	}
}
