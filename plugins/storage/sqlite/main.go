package main

import (
	"github.com/openfga/openfga/pkg/plugin"

	"github.com/openfga/openfga-contrib/storage/sqlite"
)

func InitPlugin(pm *plugin.PluginManager) error {
	sqliteDriver := &sqlite.SQLiteDriver{}
	return pm.RegisterOpenFGADatastore("sqlite", sqliteDriver)
}
