// Package services implements business logic and domain rules
package services

import (
	"math/rand"
	"sync"
	"time"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
)

// Assigner is responsible for the business logic of randomly selecting reviewers
type Assigner struct {
	mu         sync.Mutex
	randSource *rand.Rand
}

// NewAssigner creates a new Assigner instance
func NewAssigner() *Assigner {
	return &Assigner{
		// nolint:gosec
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectReviewers selects a specified count of random reviewers from a list of candidates
func (a *Assigner) SelectReviewers(candidates []*entity.User, count int) []entity.User {
	if len(candidates) == 0 || count == 0 {
		return nil // return nil if no candidates or no reviewers needed
	}

	if len(candidates) <= count {
		// if fewer candidates than requested, return all of them
		reviewers := make([]entity.User, len(candidates))
		for i, c := range candidates {
			reviewers[i] = *c
		}
		return reviewers
	}

	// create a shallow copy to prevent modifying the original slice
	tempList := make([]*entity.User, len(candidates))
	copy(tempList, candidates)

	a.mu.Lock() // protect rand.Source during shuffle operation
	// use shuffle for non-repeating random selection
	a.randSource.Shuffle(len(tempList), func(i, j int) {
		tempList[i], tempList[j] = tempList[j], tempList[i]
	})
	a.mu.Unlock()

	// take the first 'count' elements after shuffling
	selected := tempList[:count]

	// convert []*entity.User to []entity.User (value copy)
	reviewers := make([]entity.User, count)
	for i, c := range selected {
		reviewers[i] = *c
	}

	return reviewers
}
