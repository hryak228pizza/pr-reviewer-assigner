// Package entity defines core domain models
package entity

import "errors"

var (
	ErrNotFound    = errors.New("resource not found")
	ErrTeamExists  = errors.New("team already exists")
	ErrPRMerged    = errors.New("pull request is merged")
	ErrNoCandidate = errors.New("no active candidate available")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrPRExists    = errors.New("pull request with this ID already exists")
)
