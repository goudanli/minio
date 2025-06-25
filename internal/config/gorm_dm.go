/*
 * @Author: xiao.wei xiaow@suninfo.com
 * @Date: 2024-05-08 10:41:29
 * @LastEditors: xiao.wei xiaow@suninfo.com
 * @LastEditTime: 2025-06-24 16:50:25
 * @FilePath: /adm_s3gateway/internal/config/gorm_dm.go
 * @Description:
 *
 * Copyright (c) 2024 by suninfo, All Rights Reserved.
 */
package config

type DM struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *DM) Dsn() string {
	return "dm://" + m.Username + ":" + m.Password + "@" + m.Path + ":" + m.Port
}

func (m *DM) GetLogMode() string {
	return m.LogMode
}
