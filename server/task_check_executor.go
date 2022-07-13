package server

import (
	"context"

	"github.com/youzi-1122/bytebase/api"
)

// TaskCheckExecutor is the task check executor.
type TaskCheckExecutor interface {
	// Run will be called periodically by the task check scheduler
	Run(ctx context.Context, server *Server, taskCheckRun *api.TaskCheckRun) (result []api.TaskCheckResult, err error)
}
