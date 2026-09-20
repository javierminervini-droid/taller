package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/repositories"
	"taller-gestion/backend/src/utils"
)

type AgendaResult struct {
	Date        string              `json:"date"`
	Technicians []models.Technician `json:"technicians"`
	Requests    []map[string]any    `json:"requests"`
	Orders      []map[string]any    `json:"orders"` // alias for older clients
}

type OrdersListResult struct {
	Requests []map[string]any `json:"requests"`
	Orders   []map[string]any `json:"orders"` // alias
	Years    []string         `json:"years"`
	Year     any              `json:"year"`
	Month    any              `json:"month"`
	Total    int              `json:"total"`
}

func scopeFilters(q *repositories.OrderFilters, claims *utils.Claims, techID *int64) {
	if claims != nil && claims.Role == "tecnico" {
		if techID != nil {
			q.TechnicianID = strconv.FormatInt(*techID, 10)
		} else {
			q.TechnicianID = "-1"
		}
	}
}

func (s *Services) Agenda(ctx context.Context, claims *utils.Claims, q repositories.OrderFilters) (*AgendaResult, error) {
	tech := s.TechnicianOf(ctx, claims)
	if claims.Role == "tecnico" && tech == nil {
		return nil, ErrTechNotLinked
	}
	var techID *int64
	if tech != nil {
		id := tech.ID
		techID = &id
	}
	scopeFilters(&q, claims, techID)
	date := q.Date
	if date == "" {
		date = utils.LocalDate("")
	}
	q.Date = date
	where, params := q.Where()
	_ = s.GenerateFollowups(ctx)

	orders, err := s.Repos.QueryOrders(ctx, where, params, "ORDER BY o.scheduled_time, o.id")
	if err != nil {
		return nil, err
	}

	var technicians []models.Technician
	if claims.Role == "tecnico" {
		if tech != nil {
			technicians = []models.Technician{*tech}
		} else {
			technicians = []models.Technician{}
		}
	} else {
		technicians, _ = s.Repos.ListActiveTechnicians(ctx)
	}
	return &AgendaResult{Date: date, Technicians: technicians, Requests: orders, Orders: orders}, nil
}

func (s *Services) ListOrders(ctx context.Context, claims *utils.Claims, q repositories.OrderFilters) (*OrdersListResult, error) {
	tech := s.TechnicianOf(ctx, claims)
	var techID *int64
	if tech != nil {
		id := tech.ID
		techID = &id
	}
	scopeFilters(&q, claims, techID)
	where, params := q.Where()

	orders, err := s.Repos.QueryOrders(ctx, where, params, "ORDER BY COALESCE(o.received_at, o.created_at::date) DESC, o.id DESC")
	if err != nil {
		return nil, err
	}
	years, _ := s.Repos.ListOrderYears(ctx)

	var year any
	var month any
	if q.Year != "" {
		year = q.Year
	}
	if q.Month != "" {
		month = q.Month
	}
	return &OrdersListResult{
		Requests: orders,
		Orders:   orders,
		Years:    years,
		Year:     year,
		Month:    month,
		Total:    len(orders),
	}, nil
}

func (s *Services) GetOrder(ctx context.Context, claims *utils.Claims, id int64) (map[string]any, error) {
	order, err := s.Repos.QueryOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrNotFound
	}
	if claims.Role == "tecnico" {
		tech := s.TechnicianOf(ctx, claims)
		oid, _ := utils.AsInt64(order["technician_id"])
		if tech == nil || oid != tech.ID {
			return nil, ErrTechViewOrders
		}
	}
	return order, nil
}

func requestTitle(b map[string]any) string {
	if t := strings.TrimSpace(utils.StrOr(b["title"], "")); t != "" {
		return t
	}
	model := strings.TrimSpace(utils.StrOr(b["appliance_model"], utils.StrOr(b["product_label"], "")))
	fail := strings.TrimSpace(utils.StrOr(b["reported_failure"], ""))
	switch {
	case model != "" && fail != "":
		return model + " - " + fail
	case model != "":
		return model
	case fail != "":
		return fail
	default:
		return "Servicio"
	}
}

func normalizeRequestKind(v any) *string {
	s := strings.TrimSpace(strings.ToUpper(utils.StrOr(v, "")))
	if s == "G" || s == "FG" {
		return &s
	}
	return nil
}

func (s *Services) CreateOrder(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	now := time.Now()
	title := requestTitle(b)
	receivedAt := utils.OptStr(b["received_at"])
	if receivedAt == nil {
		d := now.Format("2006-01-02")
		receivedAt = &d
	}
	unitTypeID := utils.OptInt64(b["unit_type_id"])
	if unitTypeID == nil {
		unitTypeID = utils.OptInt64(b["product_type_id"])
	}
	visitDate := utils.OptStr(b["visit_date"])
	if visitDate == nil {
		visitDate = utils.OptStr(b["scheduled_date"])
	}
	applianceModel := utils.OptStr(b["appliance_model"])
	if applianceModel == nil {
		applianceModel = utils.OptStr(b["product_label"])
	}
	reportedFailure := utils.OptStr(b["reported_failure"])
	if reportedFailure == nil {
		reportedFailure = utils.OptStr(b["description"])
	}

	order := &models.ServiceRequest{
		ClientID:         utils.MustInt64(b["client_id"]),
		TechnicianID:     utils.OptInt64(b["technician_id"]),
		UnitTypeID:       unitTypeID,
		ProductID:        utils.OptInt64(b["product_id"]),
		ProductLabel:     applianceModel,
		Locality:         utils.OptStr(b["locality"]),
		ProviderID:       utils.OptInt64(b["provider_id"]),
		StatusID:         utils.MustInt64(b["status_id"]),
		Title:            title,
		Description:      reportedFailure,
		ReceivedAt:       receivedAt,
		ProviderOrderRef: utils.OptStr(b["provider_order_ref"]),
		InternalOrderNo:  utils.OptStr(b["internal_order_no"]),
		RequestKind:      normalizeRequestKind(b["request_kind"]),
		ApplianceModel:   applianceModel,
		ReportedFailure:  reportedFailure,
		VisitDate:        visitDate,
		ScheduledTime:    utils.OptStr(b["scheduled_time"]),
		OpsNotes:         utils.OptStr(b["ops_notes"]),
		DiagnosisNotes:   utils.OptStr(b["diagnosis_notes"]),
		Hours:            utils.FloatOr(b["hours"], 0),
		Km:               utils.FloatOr(b["km"], 0),
		PartsCost:        utils.FloatOr(b["parts_cost"], 0),
		PartsSale:        utils.FloatOr(b["parts_sale"], 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.Repos.InsertOrder(ctx, order); err != nil {
		return 0, err
	}
	_ = s.GenerateFollowups(ctx)
	return order.ID, nil
}

func (s *Services) PatchOrder(ctx context.Context, claims *utils.Claims, id int64, b map[string]any) (*WhatsAppMessage, error) {
	current, err := s.Repos.FindOrderByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if claims.Role == "tecnico" {
		tech := s.TechnicianOf(ctx, claims)
		if tech == nil || current.TechnicianID == nil || *current.TechnicianID != tech.ID {
			return nil, ErrTechUpdateOrders
		}
	}
	if b == nil {
		b = map[string]any{}
	}

	// Accept legacy aliases from older clients.
	if _, ok := b["unit_type_id"]; !ok {
		if v, ok := b["product_type_id"]; ok {
			b["unit_type_id"] = v
		}
	}
	if _, ok := b["visit_date"]; !ok {
		if v, ok := b["scheduled_date"]; ok {
			b["visit_date"] = v
		}
	}
	if _, ok := b["appliance_model"]; !ok {
		if v, ok := b["product_label"]; ok {
			b["appliance_model"] = v
		}
	}

	allowed := map[string]any{
		"client_id":          current.ClientID,
		"technician_id":      current.TechnicianID,
		"unit_type_id":       current.UnitTypeID,
		"product_id":         current.ProductID,
		"product_label":      current.ProductLabel,
		"locality":           current.Locality,
		"provider_id":        current.ProviderID,
		"status_id":          current.StatusID,
		"title":              current.Title,
		"description":        current.Description,
		"received_at":        current.ReceivedAt,
		"provider_order_ref": current.ProviderOrderRef,
		"internal_order_no":  current.InternalOrderNo,
		"request_kind":       current.RequestKind,
		"appliance_model":    current.ApplianceModel,
		"reported_failure":   current.ReportedFailure,
		"visit_date":         current.VisitDate,
		"scheduled_time":     current.ScheduledTime,
		"ops_notes":          current.OpsNotes,
		"diagnosis_notes":    current.DiagnosisNotes,
		"hours":              current.Hours,
		"km":                 current.Km,
		"parts_cost":         current.PartsCost,
		"parts_sale":         current.PartsSale,
	}

	techKeys := []string{
		"title", "description", "status_id", "appliance_model", "product_label",
		"unit_type_id", "scheduled_time", "locality", "reported_failure",
		"diagnosis_notes", "ops_notes",
	}
	if claims.Role == "tecnico" {
		for _, key := range techKeys {
			if v, ok := b[key]; ok {
				allowed[key] = v
			}
		}
	} else {
		for k, v := range b {
			allowed[k] = v
		}
	}

	if v, ok := allowed["request_kind"]; ok {
		allowed["request_kind"] = normalizeRequestKind(v)
	}
	if am := utils.OptStr(allowed["appliance_model"]); am != nil {
		allowed["product_label"] = am
	}
	if rf := utils.OptStr(allowed["reported_failure"]); rf != nil {
		allowed["description"] = rf
	}
	allowed["title"] = requestTitle(allowed)

	startedAt := current.StartedAt
	completedAt := current.CompletedAt
	newStatusID := utils.MustInt64(allowed["status_id"])
	statusChanged := newStatusID != 0 && newStatusID != current.StatusID
	if statusChanged {
		if startedAt == nil {
			now := time.Now()
			startedAt = &now
		}
		if st, err := s.Repos.FindStatusByID(ctx, newStatusID); err == nil && st.IsClosed {
			now := time.Now()
			completedAt = &now
		}
	}

	allowed["status_id"] = newStatusID
	if err := s.Repos.UpdateOrder(ctx, id, allowed, startedAt, completedAt); err != nil {
		return nil, err
	}
	_ = s.GenerateFollowups(ctx)

	var wa *WhatsAppMessage
	if statusChanged {
		order, err := s.Repos.QueryOrderByID(ctx, id)
		if err == nil && order != nil {
			wa, _ = s.messageForOrder(ctx, order, nil)
			if wa != nil && wa.URL == "" {
				wa = nil
			}
		}
	}
	return wa, nil
}
