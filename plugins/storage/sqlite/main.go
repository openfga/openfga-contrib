package main

import (
	"github.com/openfga/openfga-contrib/storage/sqlite"
	"github.com/openfga/openfga/pkg/plugin"
)

func InitPlugin(pm *plugin.PluginManager) error {
	sqliteDriver := &sqlite.SQLiteDriver{}
	return pm.RegisterOpenFGADatastore("sqlite", sqliteDriver)
}
