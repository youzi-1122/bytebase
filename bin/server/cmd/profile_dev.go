//go:build !release
// +build !release

package cmd

import (
	"fmt"
	"time"

	"github.com/youzi-1122/bytebase/common"
	"github.com/youzi-1122/bytebase/server"
)

func activeProfile(dataDir string) server.Profile {
	// `flags.demo` always be true in dev mode
	demoName := string(common.ReleaseModeDev)
	if flags.demoName != "" {
		demoName = flags.demoName
	}
	demoDataDir := fmt.Sprintf("demo/%s", demoName)
	// Using flags.port + 1 as our datastore port
	datastorePort := flags.port + 1

	return server.Profile{
		Mode:                 common.ReleaseModeDev,
		BackendHost:          flags.host,
		BackendPort:          flags.port,
		FrontendHost:         flags.frontendHost,
		FrontendPort:         flags.frontendPort,
		DatastorePort:        datastorePort,
		PgUser:               "bbdev",
		Readonly:             flags.readonly,
		Debug:                flags.debug,
		Demo:                 flags.demo,
		DataDir:              dataDir,
		DemoDataDir:          demoDataDir,
		BackupRunnerInterval: 10 * time.Second,
		Version:              version,
		GitCommit:            gitcommit,
		PgURL:                flags.pgURL,
		MetricConnectionKey:  "3zcZLeX3ahvlueEJqNyJysGfVAErsjjT",
	}
}
