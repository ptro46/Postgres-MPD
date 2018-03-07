package main

import (
	"fmt"
)

type PostgresMPDConfig struct {
	Login      string `json:"login,omitempty"`
	Password   string `json:"password,omitempty"`
	Host       string `json:"host,omitempty"`
	Port       int64  `json:"port,omitempty"`
	Parameters string `json:"parameters,omitempty"`
	Db         string `json:"db,omitempty"`
}

func (config *PostgresMPDConfig) dbPostgresConnectString() string {
	return fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d %s", config.Login, config.Password, config.Host, config.Db, config.Port, config.Parameters)
}
