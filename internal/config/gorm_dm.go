package config

import (
	"net/url"
)

type DM struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

func (m *DM) Dsn() string {
	password := url.QueryEscape(m.Password)
	return "dm://" + m.Username + ":" + password + "@" + m.Path + ":" + m.Port
}

func (m *DM) GetLogMode() string {
	return m.LogMode
}
