/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 10:41:29
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2024-05-08 14:02:33
 * @FilePath: /minio/internal/config/gorm_mysql.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package config

type Mysql struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

func (m *Mysql) GetLogMode() string {
	return m.LogMode
}
