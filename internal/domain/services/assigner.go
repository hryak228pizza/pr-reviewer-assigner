package services

import (
	"math/rand"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"sync"
	"time"
)

type Assigner struct {
	mu         sync.Mutex
	randSource *rand.Rand
}

func NewAssigner() *Assigner {
	return &Assigner{
		// nolint:gosec
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (a *Assigner) SelectReviewers(candidates []*entity.User, count int) []entity.User {
	if len(candidates) == 0 || count == 0 {
		return nil
	}

	if len(candidates) <= count {
		reviewers := make([]entity.User, len(candidates))
		for i, c := range candidates {
			reviewers[i] = *c
		}
		return reviewers
	}

	tempList := make([]*entity.User, len(candidates))
	copy(tempList, candidates)

	a.mu.Lock()
	a.randSource.Shuffle(len(tempList), func(i, j int) {
		tempList[i], tempList[j] = tempList[j], tempList[i]
	})
	a.mu.Unlock()

	selected := tempList[:count]

	reviewers := make([]entity.User, count)
	for i, c := range selected {
		reviewers[i] = *c
	}

	return reviewers
}
