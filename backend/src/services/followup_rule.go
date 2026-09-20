package services

import (
	"context"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

func (s *Services) ListFollowupRules(ctx context.Context, claims *utils.Claims) ([]map[string]any, error) {
	if !utils.CanManage(claims.Role) {
		return nil, ErrForbidden
	}
	return s.Repos.ListFollowupRulesJoined(ctx)
}

func (s *Services) CreateFollowupRule(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	assignRole := utils.StrOr(b["assign_role"], "coordinador")
	var hours *int
	if v := utils.OptInt64(b["hours"]); v != nil {
		h := int(*v)
		hours = &h
	}
	rule := &models.FollowupRule{
		Name:        utils.StrOr(b["name"], ""),
		TriggerType: utils.StrOr(b["trigger_type"], ""),
		StatusID:    utils.OptInt64(b["status_id"]),
		Hours:       hours,
		AssignRole:  &assignRole,
		Active:      true,
	}
	if err := s.Repos.InsertFollowupRule(ctx, rule); err != nil {
		return 0, err
	}
	return rule.ID, nil
}
