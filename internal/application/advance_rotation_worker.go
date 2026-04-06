package application

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type AdvanceRotationWorker struct {
	rotationRepo domain.RotationRepository
	memberRepo   domain.MemberRepository
	interval     time.Duration
	logger       *slog.Logger
}

func NewAdvanceRotationWorker(
	rotationRepo domain.RotationRepository,
	memberRepo domain.MemberRepository,
	interval time.Duration,
	logger *slog.Logger,
) *AdvanceRotationWorker {
	return &AdvanceRotationWorker{
		rotationRepo: rotationRepo,
		memberRepo:   memberRepo,
		interval:     interval,
		logger:       logger,
	}
}

func (w *AdvanceRotationWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *AdvanceRotationWorker) tick(ctx context.Context) {
	rotations, err := w.rotationRepo.List(ctx)
	if err != nil {
		w.logger.Error("failed to list rotations", "error", err)
		return
	}

	var wg sync.WaitGroup
	for _, rot := range rotations {
		wg.Add(1)
		go func(rot *domain.Rotation) {
			defer wg.Done()
			w.processRotation(ctx, rot)
		}(rot)
	}
	wg.Wait()
}

func (w *AdvanceRotationWorker) processRotation(ctx context.Context, rot *domain.Rotation) {
	needs, handoffTime, err := rot.NeedsAdvance(time.Now().UTC())
	if err != nil {
		w.logger.Error("failed to check if rotation needs advance", "rotation_id", rot.ID, "error", err)
		return
	}
	if !needs {
		return
	}

	fullRot, err := w.rotationRepo.GetByID(ctx, rot.ID)
	if err != nil {
		w.logger.Error("failed to get rotation", "rotation_id", rot.ID, "error", err)
		return
	}

	next, err := fullRot.NextMember()
	if err != nil {
		w.logger.Error("failed to get next member", "rotation_id", rot.ID, "error", err)
		return
	}

	if err := w.memberRepo.SetCurrentMember(ctx, rot.ID, next.ID, handoffTime); err != nil {
		w.logger.Error("failed to set current member", "rotation_id", rot.ID, "member_id", next.ID, "error", err)
		return
	}

	w.logger.Info("rotation advanced", "rotation_id", rot.ID, "new_member_id", next.ID)
}
