package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"taller-gestion/backend/src/models"
	"taller-gestion/backend/src/utils"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func Connect(databaseURL string) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	sqldb.SetMaxOpenConns(10)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(30 * time.Minute)
	return bun.NewDB(sqldb, pgdialect.New())
}

func Migrate(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, filename := range files {
		var exists int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE filename = ?`, filename,
		).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}
		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations (filename) VALUES (?)`, filename,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", filename, err)
		}
	}
	return nil
}

func findMigrationsDir() (string, error) {
	candidates := []string{
		"migrations",
		filepath.Join("backend", "migrations"),
		filepath.Join("..", "migrations"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "migrations"),
			filepath.Join(wd, "backend", "migrations"),
		)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "migrations"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}

func SeedIfEmpty(ctx context.Context, db *bun.DB) error {
	n, err := db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return seed(ctx, db)
}

// ForceSeed truncates app tables and reloads the demo dataset (keeps the Postgres volume).
func ForceSeed(ctx context.Context, db *bun.DB) error {
	if _, err := db.ExecContext(ctx, `
		TRUNCATE TABLE
			followups,
			followup_rules,
			service_orders,
			whatsapp_templates,
			tariffs,
			products,
			product_types,
			clients,
			technicians,
			providers,
			service_statuses,
			users
		RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	return seed(ctx, db)
}

func seed(ctx context.Context, db *bun.DB) error {
	hashAdmin, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}
	hashCoord, _ := utils.HashPassword("coord123")
	hashDiego, _ := utils.HashPassword("diego123")
	hashSofia, _ := utils.HashPassword("sofia123")

	admin := &models.User{Username: "admin", PasswordHash: hashAdmin, FullName: "Administrador", Role: "admin", Active: true}
	coord := &models.User{Username: "coord", PasswordHash: hashCoord, FullName: "Laura Coordinación", Role: "coordinador", Active: true}
	diego := &models.User{Username: "diego", PasswordHash: hashDiego, FullName: "Diego Méndez", Role: "tecnico", Active: true}
	sofia := &models.User{Username: "sofia", PasswordHash: hashSofia, FullName: "Sofía Rivas", Role: "tecnico", Active: true}

	for _, u := range []*models.User{admin, coord, diego, sofia} {
		if _, err := db.NewInsert().Model(u).Exec(ctx); err != nil {
			return err
		}
	}

	proveedor := &models.Provider{
		Name: "Repuestos del Sur", Kind: "proveedor",
		Contact: strPtr("11 4567-8900"), Notes: strPtr("Base de clientes y stock de recambios"),
	}
	prestador := &models.Provider{
		Name: "Garantías Andinas", Kind: "prestador",
		Contact: strPtr("11 4789-1122"), Notes: strPtr("Órdenes derivadas de garantía"),
	}
	if _, err := db.NewInsert().Model(proveedor).Exec(ctx); err != nil {
		return err
	}
	if _, err := db.NewInsert().Model(prestador).Exec(ctx); err != nil {
		return err
	}

	t1 := &models.Technician{UserID: &diego.ID, Name: "Diego Méndez", Phone: strPtr("11 5555-1001"), Specialty: strPtr("Línea blanca"), Active: true}
	t2 := &models.Technician{UserID: &sofia.ID, Name: "Sofía Rivas", Phone: strPtr("11 5555-1002"), Specialty: strPtr("Electrónica"), Active: true}
	t3 := &models.Technician{Name: "Martín Acosta", Phone: strPtr("11 5555-1003"), Specialty: strPtr("Motos y herramientas"), Active: true}
	for _, t := range []*models.Technician{t1, t2, t3} {
		if _, err := db.NewInsert().Model(t).Exec(ctx); err != nil {
			return err
		}
	}

	types := map[string]*models.ProductType{
		"lavarropas": {Name: "Lavarropas"},
		"heladera":   {Name: "Heladera"},
		"aire":       {Name: "Aire acondicionado"},
		"moto":       {Name: "Motocicleta"},
	}
	for _, pt := range types {
		if _, err := db.NewInsert().Model(pt).Exec(ctx); err != nil {
			return err
		}
	}

	pCorrea := &models.Product{
		Name: "Correa lavarropas 1270 J5", SKU: strPtr("COR-1270"),
		ProductTypeID: &types["lavarropas"].ID, ProviderID: &proveedor.ID, Stock: 12, Price: 18500,
	}
	products := []*models.Product{
		pCorrea,
		{Name: "Termostato heladera", SKU: strPtr("TER-HL01"), ProductTypeID: &types["heladera"].ID, ProviderID: &proveedor.ID, Stock: 8, Price: 22100},
		{Name: "Capacitor 45uF", SKU: strPtr("CAP-45"), ProductTypeID: &types["aire"].ID, ProviderID: &proveedor.ID, Stock: 20, Price: 9800},
		{Name: "Kit freno delantero", SKU: strPtr("FRN-MOTO"), ProductTypeID: &types["moto"].ID, ProviderID: &prestador.ID, Stock: 4, Price: 45200},
	}
	for _, p := range products {
		if _, err := db.NewInsert().Model(p).Exec(ctx); err != nil {
			return err
		}
	}

	clients := []*models.Client{
		{ProviderID: &proveedor.ID, ExternalID: strPtr("P-1042"), Name: "Ana López", Phone: strPtr("11 6001-2200"), Email: strPtr("ana@correo.com"), Locality: strPtr("CABA"), Address: strPtr("Av. Rivadavia 2100"), Source: "manual"},
		{ProviderID: &prestador.ID, ExternalID: strPtr("G-331"), Name: "Carlos Pérez", Phone: strPtr("11 6002-1188"), Email: strPtr("carlos@correo.com"), Locality: strPtr("Quilmes"), Address: strPtr("Calle Mitre 450"), Source: "manual"},
		{ProviderID: &proveedor.ID, ExternalID: strPtr("P-1188"), Name: "María Suárez", Phone: strPtr("11 6003-4400"), Email: strPtr("maria@correo.com"), Locality: strPtr("La Plata"), Address: strPtr("Calle 12 n° 800"), Source: "manual"},
		{ProviderID: &prestador.ID, ExternalID: strPtr("G-402"), Name: "Jorge Díaz", Phone: strPtr("11 6004-9900"), Email: strPtr("jorge@correo.com"), Locality: strPtr("Morón"), Address: strPtr("Belgrano 90"), Source: "manual"},
	}
	for _, c := range clients {
		if _, err := db.NewInsert().Model(c).Exec(ctx); err != nil {
			return err
		}
	}

	statuses := map[string]*models.ServiceStatus{
		"ingresado":   {Name: "Ingresado", SortOrder: 1, Color: "#64748b", IsClosed: false},
		"diagnostico": {Name: "Diagnóstico", SortOrder: 2, Color: "#2563eb", IsClosed: false},
		"espera":      {Name: "Esperando repuesto", SortOrder: 3, Color: "#d97706", IsClosed: false},
		"reparacion":  {Name: "En reparación", SortOrder: 4, Color: "#7c3aed", IsClosed: false},
		"listo":       {Name: "Listo para entregar", SortOrder: 5, Color: "#059669", IsClosed: false},
		"entregado":   {Name: "Entregado", SortOrder: 6, Color: "#334155", IsClosed: true},
		"cancelado":   {Name: "Cancelado", SortOrder: 7, Color: "#dc2626", IsClosed: true},
	}
	for _, s := range []*models.ServiceStatus{
		statuses["ingresado"], statuses["diagnostico"], statuses["espera"],
		statuses["reparacion"], statuses["listo"], statuses["entregado"], statuses["cancelado"],
	} {
		if _, err := db.NewInsert().Model(s).Exec(ctx); err != nil {
			return err
		}
	}

	coordRole := "coordinador"
	rules := []*models.FollowupRule{
		{Name: "Avisar cliente: listo para entregar", TriggerType: "status", StatusID: &statuses["listo"].ID, AssignRole: &coordRole, Active: true},
		{Name: "Gestionar compra de repuesto", TriggerType: "status", StatusID: &statuses["espera"].ID, AssignRole: &coordRole, Active: true},
		{Name: "Diagnóstico demorado (48 h)", TriggerType: "elapsed_hours", StatusID: &statuses["diagnostico"].ID, Hours: intPtr(48), AssignRole: &coordRole, Active: true},
		{Name: "Repuesto demorado (72 h)", TriggerType: "elapsed_hours", StatusID: &statuses["espera"].ID, Hours: intPtr(72), AssignRole: &coordRole, Active: true},
	}
	for _, r := range rules {
		if _, err := db.NewInsert().Model(r).Exec(ctx); err != nil {
			return err
		}
	}

	iso := func(offset int) string {
		d := time.Now().AddDate(0, 0, offset)
		return d.Format("2006-01-02")
	}
	twoDaysAgo := time.Now().Add(-50 * time.Hour)
	twoDaysAgoTrunc := time.Date(twoDaysAgo.Year(), twoDaysAgo.Month(), twoDaysAgo.Day(), twoDaysAgo.Hour(), twoDaysAgo.Minute(), twoDaysAgo.Second(), 0, time.UTC)

	parseTS := func(s string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		return t
	}

	pastMonth := time.Now().AddDate(0, -1, -3)
	pastMonthDay := pastMonth.Format("2006-01-02")
	completedPast := pastMonth.Add(6 * time.Hour)

	orders := []*models.ServiceOrder{
		{
			ClientID: clients[0].ID, TechnicianID: &t1.ID, ProductTypeID: &types["lavarropas"].ID, ProductID: &pCorrea.ID,
			ProductLabel: strPtr("Lavarropas Drean 8kg"), Locality: strPtr("CABA"), ProviderID: &proveedor.ID,
			StatusID: statuses["espera"].ID, Title: "Cambio de correa",
			Description: strPtr("No centrifuga. Cliente de Repuestos del Sur."),
			ScheduledDate: strPtr(iso(0)), ScheduledTime: strPtr("09:00"),
			Hours: 1.5, Km: 12, PartsCost: 12000, PartsSale: 18500,
			StartedAt: &twoDaysAgoTrunc, CreatedAt: twoDaysAgoTrunc, UpdatedAt: time.Now(),
		},
		{
			ClientID: clients[1].ID, TechnicianID: &t2.ID, ProductTypeID: &types["heladera"].ID,
			ProductLabel: strPtr("Heladera Patrick 300L"), Locality: strPtr("Quilmes"), ProviderID: &prestador.ID,
			StatusID: statuses["diagnostico"].ID, Title: "Heladera no enfría",
			Description: strPtr("Orden de garantía G-331."),
			ScheduledDate: strPtr(iso(0)), ScheduledTime: strPtr("11:30"),
			Hours: 1, Km: 28,
			StartedAt: &twoDaysAgoTrunc, CreatedAt: twoDaysAgoTrunc, UpdatedAt: time.Now(),
		},
		{
			ClientID: clients[2].ID, TechnicianID: &t1.ID, ProductTypeID: &types["aire"].ID,
			ProductLabel: strPtr("Split BGH 3000"), Locality: strPtr("La Plata"), ProviderID: &proveedor.ID,
			StatusID: statuses["listo"].ID, Title: "Carga de gas y limpieza",
			Description: strPtr("Unidad lista. Coordinar entrega."),
			ScheduledDate: strPtr(iso(0)), ScheduledTime: strPtr("15:00"),
			Hours: 2, Km: 45, PartsCost: 5000, PartsSale: 9800,
			StartedAt: timePtr(parseTS(iso(-1) + " 10:00:00")), CreatedAt: parseTS(iso(-2) + " 09:00:00"), UpdatedAt: time.Now(),
		},
		{
			ClientID: clients[3].ID, TechnicianID: &t3.ID, ProductTypeID: &types["moto"].ID,
			ProductLabel: strPtr("Motomel Skua 150"), Locality: strPtr("Morón"), ProviderID: &prestador.ID,
			StatusID: statuses["ingresado"].ID, Title: "Frenos delanteros",
			Description: strPtr("Prestador derivó inspección."),
			ScheduledDate: strPtr(iso(1)), ScheduledTime: strPtr("10:00"),
			CreatedAt: parseTS(iso(0) + " 08:00:00"), UpdatedAt: time.Now(),
		},
		{
			ClientID: clients[0].ID, TechnicianID: &t2.ID, ProductTypeID: &types["heladera"].ID,
			ProductLabel: strPtr("Freezer horizontal"), Locality: strPtr("CABA"), ProviderID: &proveedor.ID,
			StatusID: statuses["entregado"].ID, Title: "Cambio de termostato",
			Description: strPtr("Orden histórica para probar el listado anual."),
			ScheduledDate: strPtr(pastMonthDay), ScheduledTime: strPtr("14:00"),
			Hours: 1.5, Km: 8, PartsCost: 15000, PartsSale: 22100,
			StartedAt: &pastMonth, CompletedAt: &completedPast,
			CreatedAt: pastMonth.Add(-24 * time.Hour), UpdatedAt: completedPast,
		},
		{
			ClientID: clients[2].ID, TechnicianID: &t1.ID, ProductTypeID: &types["lavarropas"].ID,
			ProductLabel: strPtr("Lavarropas Samsung"), Locality: strPtr("La Plata"), ProviderID: &prestador.ID,
			StatusID: statuses["cancelado"].ID, Title: "Ruido en centrifugado",
			Description: strPtr("Cliente canceló visita."),
			ScheduledDate: strPtr(iso(-7)), ScheduledTime: strPtr("16:00"),
			CreatedAt: parseTS(iso(-10) + " 11:00:00"), UpdatedAt: parseTS(iso(-7) + " 09:00:00"),
		},
	}
	for _, o := range orders {
		if _, err := db.NewInsert().Model(o).Exec(ctx); err != nil {
			return err
		}
	}

	waTemplates := []*models.WhatsAppTemplate{
		{
			Name: "Confirmación de visita",
			Body: "Hola {cliente}, soy de Instal Service S.A. Te confirmamos la visita del {fecha} a las {hora} por “{trabajo}”. Técnico: {tecnico}. Ante dudas respondé este mensaje.",
			StatusID: &statuses["ingresado"].ID, AutoOpen: false, Active: true,
		},
		{
			Name: "Esperando repuesto",
			Body: "Hola {cliente}: tu orden #{orden} ({trabajo}) está en espera de repuesto. Te avisamos apenas avancemos.",
			StatusID: &statuses["espera"].ID, AutoOpen: true, Active: true,
		},
		{
			Name: "Listo para entregar",
			Body: "Hola {cliente}: tu equipo ya está listo para entregar (orden #{orden}: {trabajo}). Coordinamos retiro o entrega?",
			StatusID: &statuses["listo"].ID, AutoOpen: true, Active: true,
		},
		{
			Name: "Mensaje libre",
			Body: "Hola {cliente}, te escribimos por tu servicio “{trabajo}” (orden #{orden}).",
			AutoOpen: false, Active: true,
		},
	}
	for _, t := range waTemplates {
		if _, err := db.NewInsert().Model(t).Exec(ctx); err != nil {
			return err
		}
	}

	tariffs := []*models.Tariff{
		{
			Name: "Tarifa general del taller", Scope: "general",
			IncomeFixed: 10000, CostFixed: 7000, CostPartsPct: 100, Active: true,
		},
		{
			Name: "Contrato proveedor (fijo + % repuestos)", Scope: "proveedor", ProviderID: &proveedor.ID,
			IncomeFixed: 12000, IncomePartsPct: 20, CostPartsPct: 0, Active: true,
		},
		{
			Name: "Contrato prestador (fijo + km)", Scope: "proveedor", ProviderID: &prestador.ID,
			IncomeFixed: 18000, IncomePerKm: 250, CostPartsPct: 0, Active: true,
		},
		{
			Name: "Costo técnico titular", Scope: "tecnico", TechnicianID: &t1.ID,
			CostFixed: 8000, CostPerHour: 2500, CostPerKm: 80, CostPartsPct: 100, Active: true,
		},
		{
			Name: "Costo técnico electrónica", Scope: "tecnico", TechnicianID: &t2.ID,
			CostFixed: 7500, CostPerHour: 2400, CostPerKm: 80, CostPartsPct: 100, Active: true,
		},
		{
			Name: "Costo técnico motos", Scope: "tecnico", TechnicianID: &t3.ID,
			CostFixed: 7000, CostPerHour: 2200, CostPerKm: 80, CostPartsPct: 100, Active: true,
		},
	}
	for _, t := range tariffs {
		if _, err := db.NewInsert().Model(t).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }
func intPtr(n int) *int       { return &n }
func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
