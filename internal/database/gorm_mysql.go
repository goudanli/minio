/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 10:43:35
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-05-08 16:56:59
 * @FilePath: /minio/internal/database/gorm_mysql.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package database

import (
	"errors"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// GormMysql 初始化Mysql数据库

func GormMysql(dbConfig *DataBase) *gorm.DB {
	m := dbConfig.Mysql
	if m.Dbname == "" {
		log.Fatal(errors.New("please set dbname"))
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig)); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		if err := sqlDB.Ping(); err != nil {
			log.Fatal(err)
		}
		log.Println("Successfully connected to MySQL database")
		return db
	}
}
