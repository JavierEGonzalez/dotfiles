package types

import "time"

type Worktree struct {
	Ticket    string
	Branch    string
	Path      string
	CreatedAt time.Time
}

func (w Worktree) FilterValue() string {
	return w.Ticket + " " + w.Branch
}
