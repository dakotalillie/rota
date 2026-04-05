package application

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type CreateMemberInput struct {
	RotationID string
	UserID     string    // non-empty = existing user; mutually exclusive with UserName/UserEmail
	UserName   string    // non-empty = new user
	UserEmail  string    // non-empty = new user
	Now        time.Time // used to set became_current_at when adding the first member
}

type CreateMemberUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	userRepo     domain.UserRepository
	memberRepo   domain.MemberRepository
}

func NewCreateMemberUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	userRepo domain.UserRepository,
	memberRepo domain.MemberRepository,
) *CreateMemberUseCase {
	return &CreateMemberUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		userRepo:     userRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *CreateMemberUseCase) Execute(ctx context.Context, input CreateMemberInput) (*domain.Member, error) {
	if input.UserID == "" && (input.UserName == "" || input.UserEmail == "") {
		return nil, domain.ErrMissingUserFields
	}

	var member *domain.Member
	err := uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		if _, err := uc.rotationRepo.GetByID(ctx, input.RotationID); err != nil {
			return err
		}

		count, err := uc.memberRepo.CountByRotationID(ctx, input.RotationID)
		if err != nil {
			return err
		}
		if count >= 20 {
			return domain.ErrRotationMembershipFull
		}

		var user *domain.User
		if input.UserID != "" {
			user, err = uc.userRepo.GetByID(ctx, input.UserID)
		} else {
			user, err = uc.userRepo.Create(ctx, input.UserName, input.UserEmail)
		}
		if err != nil {
			return err
		}

		member, err = uc.memberRepo.Create(ctx, input.RotationID, user.ID, count+1)
		if err != nil {
			return err
		}
		member.User = *user

		if count == 0 {
			if err = uc.memberRepo.SetCurrentMember(ctx, member.ID, input.Now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return member, nil
}
