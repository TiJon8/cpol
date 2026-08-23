package pgconn

import "os"

// this file should describes the default settings for config e.g. env/user/host/port


func defaultSettings() map[string]string {
	settings := make(map[string]string)

	settings["host"] = "localhost"
	settings["port"] = "5432"
	return settings
}


func envSettings() map[string]string {
	settings := make(map[string]string)

	// local env variables for Postgres, like
	varMap := map[string]string {
		"HOST": "host",
	}

	for envname, mapname := range varMap {
		value := os.Getenv(envname)
		if value != "" {
			settings[mapname] = value
		}
	}

	return settings
}
