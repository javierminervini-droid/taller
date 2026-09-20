package services

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"

	"github.com/xuri/excelize/v2"
)

var templateHeaders = []string{
	"ID_Salesforce", "Nombre", "Telefono", "Email", "Direccion", "Localidad", "Provincia", "Notas",
}

var fieldAliases = map[string][]string{
	"ID_Salesforce": {"id_salesforce", "id", "salesforce id", "account id", "accountid", "case number", "casenumber", "id externo", "external id"},
	"Nombre":        {"nombre", "name", "account name", "accountname", "razon social", "cliente"},
	"Telefono":      {"telefono", "phone", "mobile", "mobilephone", "celular"},
	"Email":         {"email", "correo"},
	"Direccion":     {"direccion", "street", "billingstreet", "address", "domicilio"},
	"Localidad":     {"localidad", "city", "billingcity", "ciudad"},
	"Provincia":     {"provincia", "state", "billingstate"},
	"Notas":         {"notas", "notes", "description", "descripcion"},
}

func normHeader(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch r {
		case 'á', 'à', 'ä', 'â':
			b.WriteRune('a')
		case 'é', 'è', 'ë', 'ê':
			b.WriteRune('e')
		case 'í', 'ì', 'ï', 'î':
			b.WriteRune('i')
		case 'ó', 'ò', 'ö', 'ô':
			b.WriteRune('o')
		case 'ú', 'ù', 'ü', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func (s *Services) BuildClientsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetName(sheet, "Clientes")
	sheet = "Clientes"
	sample := []string{
		"001XX000003ABC", "Juan Pérez", "11 5555-0000", "juan@correo.com",
		"Av. Siempre Viva 742", "Quilmes", "Buenos Aires", "Exportado desde Salesforce",
	}
	for i, h := range templateHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		cell2, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell2, sample[i])
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ImportResult struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Errors  []string `json:"errors"`
	Total   int      `json:"total"`
}

type importRow struct {
	ExternalID string
	Name       string
	Phone      string
	Email      string
	Address    string
	Locality   string
	Province   string
	Notes      string
}

func mapImportRow(headers []string, values []string) importRow {
	lookup := map[string]string{}
	for i, h := range headers {
		val := ""
		if i < len(values) {
			val = strings.TrimSpace(values[i])
		}
		lookup[normHeader(h)] = val
	}
	pick := func(field string) string {
		for _, alias := range fieldAliases[field] {
			if v, ok := lookup[alias]; ok {
				return v
			}
		}
		if v, ok := lookup[normHeader(field)]; ok {
			return v
		}
		return ""
	}
	return importRow{
		ExternalID: pick("ID_Salesforce"),
		Name:       pick("Nombre"),
		Phone:      pick("Telefono"),
		Email:      pick("Email"),
		Address:    pick("Direccion"),
		Locality:   pick("Localidad"),
		Province:   pick("Provincia"),
		Notes:      pick("Notas"),
	}
}

func localityText(name, province string) *string {
	if name == "" {
		return nil
	}
	if province != "" {
		s := name + ", " + province
		return &s
	}
	return &name
}

func (s *Services) ClientsTemplate(ctx context.Context, claims *utils.Claims) ([]byte, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return nil, ErrForbidden
	}
	return s.BuildClientsTemplate()
}

func (s *Services) ImportClients(ctx context.Context, claims *utils.Claims, providerID int64, data []byte) (*ImportResult, error) {
	if !utils.HasRole(claims.Role, "admin", "coordinador") {
		return nil, ErrForbidden
	}
	if providerID <= 0 {
		return nil, BadRequest("Elegí el prestador/proveedor de origen")
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, BadRequest("No se pudo leer el Excel. Usá la plantilla del sistema.")
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, BadRequest("No se pudo leer el Excel. Usá la plantilla del sistema.")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		return nil, BadRequest("No se pudo leer el Excel. Usá la plantilla del sistema.")
	}
	headers := rows[0]
	result := &ImportResult{Errors: []string{}, Total: len(rows) - 1}

	for i, raw := range rows[1:] {
		row := mapImportRow(headers, raw)
		if row.Name == "" {
			result.Errors = append(result.Errors, "Fila "+strconv.Itoa(i+2)+": falta Nombre")
			continue
		}
		loc := localityText(row.Locality, row.Province)
		phone := utils.OptStr(row.Phone)
		email := utils.OptStr(row.Email)
		address := utils.OptStr(row.Address)
		notes := utils.OptStr(row.Notes)
		ext := utils.OptStr(row.ExternalID)

		var existing *models.Client
		if ext != nil {
			existing, err = s.Repos.FindClientByProviderExternal(ctx, providerID, *ext)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, err
				}
				existing = nil
			}
		}
		if existing != nil {
			if err := s.Repos.UpdateClientImport(ctx, existing.ID, row.Name, phone, email, loc, address, notes); err != nil {
				return nil, err
			}
			result.Updated++
		} else {
			pid := providerID
			client := &models.Client{
				ProviderID: &pid,
				ExternalID: ext,
				Name:       row.Name,
				Phone:      phone,
				Email:      email,
				Locality:   loc,
				Address:    address,
				Notes:      notes,
				Source:     "excel",
			}
			if err := s.Repos.InsertClient(ctx, client); err != nil {
				return nil, err
			}
			result.Created++
		}
	}
	return result, nil
}
