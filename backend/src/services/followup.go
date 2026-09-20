package services

import (
	"context"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) GenerateFollowups(ctx context.Context) error {
	rules, err := s.Repos.ListActiveFollowupRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		var orderIDs []int64
		switch rule.TriggerType {
		case "status":
			if rule.StatusID == nil {
				continue
			}
			orderIDs, err = s.Repos.OrderIDsForStatusTrigger(ctx, *rule.StatusID)
			if err != nil {
				return err
			}
		case "elapsed_hours":
			if rule.StatusID == nil || rule.Hours == nil {
				continue
			}
			orderIDs, err = s.Repos.OrderIDsForElapsedTrigger(ctx, *rule.StatusID, *rule.Hours)
			if err != nil {
				return err
			}
		default:
			continue
		}

		assigneeID := s.pickAssignee(ctx, rule.AssignRole)
		for _, orderID := range orderIDs {
			exists, err := s.Repos.PendingFollowupExists(ctx, orderID, rule.ID)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			now := time.Now()
			notes := rule.Name
			f := &models.Followup{
				ServiceOrderID: orderID,
				RuleID:         &rule.ID,
				AssignedUserID: assigneeID,
				DueAt:          &now,
				Notes:          &notes,
				Status:         "pendiente",
				CreatedAt:      now,
			}
			if err := s.Repos.InsertFollowup(ctx, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Services) pickAssignee(ctx context.Context, role *string) *int64 {
	if role == nil || *role == "" {
		return nil
	}
	user, err := s.Repos.FindFirstActiveUserByRole(ctx, *role)
	if err != nil {
		return nil
	}
	return &user.ID
}

func (s *Services) ListFollowups(ctx context.Context, claims *utils.Claims, status string) ([]map[string]any, error) {
	if claims.Role == "tecnico" {
		return nil, ErrFollowupsAdmin
	}
	_ = s.GenerateFollowups(ctx)
	if status == "" {
		status = "pendiente"
	}
	return s.Repos.ListFollowups(ctx, status)
}

func (s *Services) PatchFollowup(ctx context.Context, id int64, statusIn string) error {
	status := "pendiente"
	if statusIn == "hecho" {
		status = "hecho"
	}
	if status == "hecho" {
		now := time.Now()
		return s.Repos.UpdateFollowupStatus(ctx, id, status, &now)
	}
	return s.Repos.UpdateFollowupStatus(ctx, id, status, nil)
}
