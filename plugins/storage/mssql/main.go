package main

import (
	"github.com/openfga/openfga/pkg/plugin"

	"github.com/openfga/openfga-contrib/storage/mssql"
)

func InitPlugin(pm *plugin.PluginManager) error {
	mssqlDriver := &mssql.MSSQLDriver{}
	return pm.RegisterOpenFGADatastore("mssql", mssqlDriver)
}
