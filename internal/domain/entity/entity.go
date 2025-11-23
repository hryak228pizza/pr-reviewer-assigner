package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type Team struct {
	Name string `db:"name" json:"team_name"`
}

type User struct {
	ID       string `db:"id" json:"user_id"`
	Username string `db:"username" json:"username"`
	TeamName string `db:"team_name" json:"team_name"`
	IsActive bool   `db:"is_active" json:"is_active"`
}

type PullRequest struct {
	ID        string     `db:"id" json:"pull_request_id"`
	Name      string     `db:"name" json:"pull_request_name"`
	AuthorID  string     `db:"author_id" json:"author_id"`
	Status    PRStatus   `db:"status" json:"status"`
	Reviewers []User     `db:"-" json:"assigned_reviewers"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	MergedAt  *time.Time `db:"merged_at" json:"mergedAt,omitempty"`
}
