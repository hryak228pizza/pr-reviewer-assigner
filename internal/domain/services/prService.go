package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
)

type PRService interface {
	Create(ctx context.Context, prID, prName, authorID string) (*entity.PullRequest, error)
	Merge(ctx context.Context, prID string) (*entity.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*entity.PullRequest, string, error)
}

type PRUseCase struct {
	prRepo     repository.PRRepository
	userRepo   repository.UserRepository
	transactor repository.Transactor
	assigner   *Assigner
}

func NewPRUseCase(prRepo repository.PRRepository, userRepo repository.UserRepository, transactor repository.Transactor, assigner *Assigner) *PRUseCase {
	return &PRUseCase{
		prRepo:     prRepo,
		userRepo:   userRepo,
		transactor: transactor,
		assigner:   assigner,
	}
}

func (uc *PRUseCase) Create(ctx context.Context, prID, prName, authorID string) (*entity.PullRequest, error) {
	var createdPR *entity.PullRequest

	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		author, err := uc.userRepo.GetByID(txCtx, authorID)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return fmt.Errorf("author not found: %w", entity.ErrNotFound)
			}
			return err
		}

		candidates, err := uc.userRepo.GetActiveCandidatesByTeam(txCtx, author.TeamName, authorID)
		if err != nil {
			return err
		}

		reviewers := uc.assigner.SelectReviewers(candidates, 2)

		pr := &entity.PullRequest{
			ID:        prID,
			Name:      prName,
			AuthorID:  authorID,
			Status:    entity.StatusOpen,
			Reviewers: reviewers,
			CreatedAt: time.Now(),
		}

		if err := uc.prRepo.Create(txCtx, pr); err != nil {
			return err
		}

		createdPR = pr
		return nil
	})

	return createdPR, err
}

func (uc *PRUseCase) Merge(ctx context.Context, prID string) (*entity.PullRequest, error) {
    pr, err := uc.prRepo.GetByID(ctx, prID)
    if err != nil {
        return nil, err
    }

    if pr.Status == entity.StatusMerged {
        return pr, nil
    }

    mergedPR, err := uc.prRepo.UpdateStatus(ctx, prID, entity.StatusMerged)
    if err != nil {
        return nil, err
    }

    reviewersPtrs, err := uc.prRepo.GetReviewersByPRID(ctx, prID)
    if err != nil {
        return nil, fmt.Errorf("failed to load reviewers for merged PR %s: %w", prID, err)
    }

    reviewers := make([]entity.User, len(reviewersPtrs))
    for i, r := range reviewersPtrs {
        reviewers[i] = *r
    }
    
    mergedPR.Reviewers = reviewers 
    return mergedPR, nil
}

func (uc *PRUseCase) Reassign(ctx context.Context, prID, oldReviewerID string) (*entity.PullRequest, string, error) {
	var newReviewerID string
	var updatedPR *entity.PullRequest

	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		pr, err := uc.prRepo.GetByID(txCtx, prID)
		if err != nil {
			return err
		}

		if pr.Status == entity.StatusMerged {
			return entity.ErrPRMerged
		}

		var oldReviewer *entity.User
		var oldReviewerIndex int

		for i, rev := range pr.Reviewers {
			if rev.ID == oldReviewerID {
				oldReviewer = &rev
				oldReviewerIndex = i
				break
			}
		}

		if oldReviewer == nil {
			return entity.ErrNotAssigned
		}

		currentReviewerIDs := make(map[string]bool)
		for _, rev := range pr.Reviewers {
			currentReviewerIDs[rev.ID] = true
		}
		delete(currentReviewerIDs, oldReviewerID)

		candidates, err := uc.userRepo.GetActiveCandidatesByTeam(txCtx, oldReviewer.TeamName, oldReviewerID)
		if err != nil {
			return err
		}

		filteredCandidates := make([]*entity.User, 0)
		for _, c := range candidates {
			if !currentReviewerIDs[c.ID] && c.ID != pr.AuthorID {
				filteredCandidates = append(filteredCandidates, c)
			}
		}

		if len(filteredCandidates) == 0 {
			return entity.ErrNoCandidate
		}

		newReviewer := uc.assigner.SelectReviewers(filteredCandidates, 1)[0]
		newReviewerID = newReviewer.ID

		newReviewers := make([]string, 0, len(pr.Reviewers))

		for i, rev := range pr.Reviewers {
			if i != oldReviewerIndex {
				newReviewers = append(newReviewers, rev.ID)
			}
		}
		newReviewers = append(newReviewers, newReviewerID)

		if setErr := uc.prRepo.SetReviewers(txCtx, prID, newReviewers); setErr != nil {
			return setErr
		}

		updatedPR, err = uc.prRepo.GetByID(txCtx, prID)
		return err
	})

	return updatedPR, newReviewerID, err
}
