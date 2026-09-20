package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"
)

type WhatsAppMessage struct {
	URL        string `json:"url,omitempty"`
	Text       string `json:"text,omitempty"`
	Template   string `json:"template,omitempty"`
	TemplateID int64  `json:"template_id,omitempty"`
	AutoOpen   bool   `json:"auto_open,omitempty"`
	Tel        string `json:"tel,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (s *Services) ListWhatsAppTemplates(ctx context.Context, claims *utils.Claims) ([]map[string]any, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return nil, ErrForbidden
	}
	list, err := s.Repos.ListWhatsAppTemplates(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []map[string]any{}
	}
	return list, nil
}

func (s *Services) CreateWhatsAppTemplate(ctx context.Context, claims *utils.Claims, b map[string]any) (int64, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return 0, ErrForbidden
	}
	t := &models.WhatsAppTemplate{
		Name:     utils.StrOr(b["name"], ""),
		Body:     utils.StrOr(b["body"], ""),
		StatusID: utils.OptInt64(b["status_id"]),
		AutoOpen: truthy(b["auto_open"]),
		Active:   true,
	}
	if t.Name == "" || t.Body == "" {
		return 0, BadRequest("nombre y cuerpo requeridos")
	}
	if err := s.Repos.InsertWhatsAppTemplate(ctx, t); err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (s *Services) PatchWhatsAppTemplate(ctx context.Context, claims *utils.Claims, id int64, b map[string]any) error {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return ErrForbidden
	}
	current, err := s.Repos.FindWhatsAppTemplateByID(ctx, id)
	if err != nil {
		return ErrTemplateNotFound
	}
	if b == nil {
		b = map[string]any{}
	}
	name := current.Name
	if v, ok := b["name"]; ok {
		name = utils.StrOr(v, name)
	}
	body := current.Body
	if v, ok := b["body"]; ok {
		body = utils.StrOr(v, body)
	}
	statusID := current.StatusID
	if v, ok := b["status_id"]; ok {
		statusID = utils.OptInt64(v)
	}
	autoOpen := current.AutoOpen
	if v, ok := b["auto_open"]; ok {
		autoOpen = truthy(v)
	}
	active := current.Active
	if v, ok := b["active"]; ok {
		active = truthy(v)
		if n, ok := utils.AsInt64(v); ok && n == 0 {
			active = false
		}
	}
	return s.Repos.UpdateWhatsAppTemplate(ctx, id, name, body, statusID, autoOpen, active)
}

func (s *Services) OrderWhatsApp(ctx context.Context, claims *utils.Claims, orderID int64, templateID *int64) (*WhatsAppMessage, error) {
	order, err := s.Repos.QueryOrderByID(ctx, orderID)
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
			return nil, ErrTechContactOrders
		}
	}
	msg, err := s.messageForOrder(ctx, order, templateID)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, BadRequest("No hay plantilla o teléfono")
	}
	if msg.Error != "" && msg.URL == "" {
		return nil, BadRequest(msg.Error)
	}
	if tel := utils.TelHref(fmt.Sprint(order["client_phone"])); tel != nil {
		msg.Tel = *tel
	}
	return msg, nil
}

func (s *Services) messageForOrder(ctx context.Context, order map[string]any, templateID *int64) (*WhatsAppMessage, error) {
	var tpl *models.WhatsAppTemplate
	var err error
	if templateID != nil && *templateID > 0 {
		tpl, err = s.Repos.FindActiveTemplateByID(ctx, *templateID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if tpl == nil {
		if sid, ok := utils.AsInt64(order["status_id"]); ok && sid > 0 {
			tpl, err = s.Repos.FindActiveTemplateByStatus(ctx, sid)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
		}
	}
	if tpl == nil {
		tpl, err = s.Repos.FindActiveGenericTemplate(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if tpl == nil {
		return nil, nil
	}
	text := utils.FillTemplate(tpl.Body, utils.OrderWhatsAppVars(order))
	phone := ""
	if v, ok := order["client_phone"]; ok && v != nil {
		phone = fmt.Sprint(v)
	}
	url := utils.BuildWhatsAppURL(phone, text)
	msg := &WhatsAppMessage{
		Text:       text,
		Template:   tpl.Name,
		TemplateID: tpl.ID,
		AutoOpen:   tpl.AutoOpen,
		URL:        url,
	}
	if url == "" {
		msg.Error = "El cliente no tiene teléfono válido"
	}
	return msg, nil
}

func truthy(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case int64:
		return t != 0
	case string:
		return t == "1" || t == "true" || t == "True"
	default:
		return false
	}
}
