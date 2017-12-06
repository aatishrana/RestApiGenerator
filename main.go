package main

import (
	"os"
	"jsonconfig"
	"database"
	"server"
	"encoding/json"
	"fmt"
	"appinfo"
	"generator"
)

type configuration struct {
	Database database.Info
	Server   server.Server
	AppInfo  appinfo.AppInfo
}

func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

var config = &configuration{}

func main() {

	// Load the configuration file
	jsonconfig.Load("config"+string(os.PathSeparator)+"config.json", config)

	// Connect to database
	database.Connect(config.Database)

	// migrate tables
	database.SQL.AutoMigrate(&generator.Entity{},
		&generator.Column{},
		&generator.ColumnType{},
		&generator.Relation{},
		&generator.RelationType{})

	fmt.Println(config.AppInfo.Name, "generated!!")
}
