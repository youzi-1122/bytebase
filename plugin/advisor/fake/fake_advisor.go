package fake

import (
	"github.com/youzi-1122/bytebase/plugin/advisor"
)

var (
	_ advisor.Advisor = (*Advisor)(nil)
)

func init() {
	advisor.Register(advisor.MySQL, advisor.Fake, &Advisor{})
	advisor.Register(advisor.Postgres, advisor.Fake, &Advisor{})
	advisor.Register(advisor.TiDB, advisor.Fake, &Advisor{})
}

// Advisor is the fake sql advisor.
type Advisor struct {
}

// Check is a fake advisor check reporting 1 advice for each severity.
func (adv *Advisor) Check(ctx advisor.Context, statement string) ([]advisor.Advice, error) {
	return []advisor.Advice{
		{
			Status:  advisor.Success,
			Code:    advisor.Ok,
			Title:   "INFO check",
			Content: statement,
		},
		{
			Status:  advisor.Warn,
			Code:    advisor.Internal,
			Title:   "WARN check",
			Content: statement,
		},
		{
			Status:  advisor.Error,
			Code:    advisor.Internal,
			Title:   "ERROR check",
			Content: statement,
		},
	}, nil
}
