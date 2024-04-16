package main

import (
	"github.com/openfga/openfga/pkg/plugin"
)

func InitPlugin(pm *plugin.PluginManager) error {
	sqliteDriver := &SQLiteDriver{}
	return pm.RegisterOpenFGADatastore("sqlite", sqliteDriver)
}
