// Package services implements business logic and domain rules
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
)

// PRService defines the interface for pull request operations
type PRService interface {
	Create(ctx context.Context, prID, prName, authorID string) (*entity.PullRequest, error)
	Merge(ctx context.Context, prID string) (*entity.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*entity.PullRequest, string, error)
}

// PRUseCase implements the prservice interface
type PRUseCase struct {
	prRepo     repository.PRRepository
	userRepo   repository.UserRepository
	transactor repository.Transactor
	assigner   *Assigner
}

// NewPRUseCase is the constructor for prusecase
func NewPRUseCase(prRepo repository.PRRepository, userRepo repository.UserRepository, transactor repository.Transactor, assigner *Assigner) *PRUseCase {
	return &PRUseCase{
		prRepo:     prRepo,
		userRepo:   userRepo,
		transactor: transactor,
		assigner:   assigner,
	}
}

// Create handles the creation of a pr and initial reviewer assignment
func (uc *PRUseCase) Create(ctx context.Context, prID, prName, authorID string) (*entity.PullRequest, error) {
	var createdPR *entity.PullRequest

	// wrap all database operations in a transaction
	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		// fetch author to determine the team
		author, err := uc.userRepo.GetByID(txCtx, authorID)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return fmt.Errorf("author not found: %w", entity.ErrNotFound)
			}
			return err
		}

		// get active candidates from the author's team, excluding the author
		candidates, err := uc.userRepo.GetActiveCandidatesByTeam(txCtx, author.TeamName, authorID)
		if err != nil {
			return err
		}

		// select up to 2 reviewers randomly
		reviewers := uc.assigner.SelectReviewers(candidates, 2)

		// build the new pull request entity
		pr := &entity.PullRequest{
			ID:        prID,
			Name:      prName,
			AuthorID:  authorID,
			Status:    entity.StatusOpen,
			Reviewers: reviewers,
			CreatedAt: time.Now(),
		}

		// save the pr and its reviewers to the database
		if err := uc.prRepo.Create(txCtx, pr); err != nil {
			return err
		}

		createdPR = pr
		return nil
	})

	return createdPR, err
}

// Merge sets the pr status to merged
func (uc *PRUseCase) Merge(ctx context.Context, prID string) (*entity.PullRequest, error) {
	// check if pr exists
	pr, err := uc.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	// exit early if already merged (idempotency)
	if pr.Status == entity.StatusMerged {
		return pr, nil
	}

	// update the status to merged
	mergedPR, err := uc.prRepo.UpdateStatus(ctx, prID, entity.StatusMerged)
	if err != nil {
		return nil, err
	}

	// load reviewers for the final response entity
	// since update status might not return the full pr struct
	reviewersPtrs, err := uc.prRepo.GetReviewersByPRID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reviewers for merged PR %s: %w", prID, err)
	}

	// convert from pointer slice to value slice
	reviewers := make([]entity.User, len(reviewersPtrs))
	for i, r := range reviewersPtrs {
		reviewers[i] = *r
	}

	mergedPR.Reviewers = reviewers // attach reviewers to the response entity
	return mergedPR, nil
}

// Reassign replaces one reviewer with a random new one from the same team
func (uc *PRUseCase) Reassign(ctx context.Context, prID, oldReviewerID string) (*entity.PullRequest, string, error) {
	var newReviewerID string
	var updatedPR *entity.PullRequest

	// wrap all operations in a transaction
	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		// load pr
		pr, err := uc.prRepo.GetByID(txCtx, prID)
		if err != nil {
			return err
		}

		// cannot reassign if merged
		if pr.Status == entity.StatusMerged {
			return entity.ErrPRMerged
		}

		// validate the old reviewer is currently assigned
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

		// collect current reviewer ids to exclude them from candidates
		currentReviewerIDs := make(map[string]bool)
		for _, rev := range pr.Reviewers {
			currentReviewerIDs[rev.ID] = true
		}
		delete(currentReviewerIDs, oldReviewerID) // exclude the one we are replacing from the check for existing reviewers

		// get active candidates from the old reviewer's team (excluding the old reviewer)
		candidates, err := uc.userRepo.GetActiveCandidatesByTeam(txCtx, oldReviewer.TeamName, oldReviewerID)
		if err != nil {
			return err
		}

		// filter candidates to exclude existing reviewers and the pr author
		filteredCandidates := make([]*entity.User, 0)
		for _, c := range candidates {
			if !currentReviewerIDs[c.ID] && c.ID != pr.AuthorID {
				filteredCandidates = append(filteredCandidates, c)
			}
		}

		if len(filteredCandidates) == 0 {
			return entity.ErrNoCandidate // no one available to replace the reviewer
		}

		// select one random replacement
		newReviewer := uc.assigner.SelectReviewers(filteredCandidates, 1)[0]
		newReviewerID = newReviewer.ID

		// build the new list of reviewer ids
		newReviewers := make([]string, 0, len(pr.Reviewers))

		for i, rev := range pr.Reviewers {
			// copy all ids except the one being replaced
			if i != oldReviewerIndex {
				newReviewers = append(newReviewers, rev.ID)
			}
		}
		newReviewers = append(newReviewers, newReviewerID) // add the new reviewer

		// update the pr_reviewers table
		if setErr := uc.prRepo.SetReviewers(txCtx, prID, newReviewers); setErr != nil {
			return setErr
		}

		// fetch the final pr entity for the response
		updatedPR, err = uc.prRepo.GetByID(txCtx, prID)
		return err
	})

	return updatedPR, newReviewerID, err
}
