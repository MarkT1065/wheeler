package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"stonks/internal/database"
	"stonks/internal/models"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	db                  *sql.DB
	optionService       *models.OptionService
	symbolService       *models.SymbolService
	treasuryService     *models.TreasuryService
	longPositionService *models.LongPositionService
	dividendService     *models.DividendService
	templates           *template.Template
}

func NewServer() (*Server, error) {
	log.Printf("[SERVER] Initializing Wheeler web server")

	dbPath, err := database.GetCurrentDatabasePath()
	if err != nil {
		log.Printf("[SERVER] ERROR: Failed to get current database path: %v", err)
		return nil, fmt.Errorf("failed to get current database path: %w", err)
	}
	log.Printf("[SERVER] Connecting to database: %s", dbPath)
	dbWrapper, err := database.NewDB(dbPath)
	if err != nil {
		log.Printf("[SERVER] ERROR: Failed to initialize database: %v", err)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Printf("[SERVER] Database connection established successfully")

	// Load templates with custom functions
	templatePath := filepath.Join("internal", "web", "templates", "*.html")
	log.Printf("[SERVER] Loading HTML templates from: %s", templatePath)
	
	// Create template with custom functions
	funcMap := template.FuncMap{
		"groupByExpiration": groupPositionsByExpiration,
		"replace": func(old, new, src string) string {
			return strings.Replace(src, old, new, -1)
		},
		"add": func(a, b interface{}) interface{} {
			switch aVal := a.(type) {
			case int:
				if bVal, ok := b.(int); ok {
					return aVal + bVal
				}
			case float64:
				if bVal, ok := b.(float64); ok {
					return aVal + bVal
				}
			}
			return 0
		},
		"mul": func(a, b interface{}) interface{} {
			switch aVal := a.(type) {
			case int:
				if bVal, ok := b.(int); ok {
					return float64(aVal * bVal)
				}
				if bVal, ok := b.(float64); ok {
					return float64(aVal) * bVal
				}
			case float64:
				if bVal, ok := b.(int); ok {
					return aVal * float64(bVal)
				}
				if bVal, ok := b.(float64); ok {
					return aVal * bVal
				}
			}
			return 0.0
		},
		"formatCurrency": func(value interface{}) string {
			var floatVal float64
			switch v := value.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			default:
				return "$0"
			}
			
			// Round to nearest whole number
			rounded := int64(floatVal + 0.5)
			if floatVal < 0 {
				rounded = int64(floatVal - 0.5)
			}
			
			// Format with commas
			str := fmt.Sprintf("%d", rounded)
			if rounded < 0 {
				str = str[1:] // Remove negative sign temporarily
			}
			
			// Add commas
			if len(str) > 3 {
				var result string
				for i, digit := range str {
					if i > 0 && (len(str)-i)%3 == 0 {
						result += ","
					}
					result += string(digit)
				}
				str = result
			}
			
			if floatVal < 0 {
				return "-$" + str
			}
			return "$" + str
		},
	}
	
	templates, err := template.New("").Funcs(funcMap).ParseGlob(templatePath)
	if err != nil {
		log.Printf("[SERVER] ERROR: Failed to parse templates: %v", err)
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	log.Printf("[SERVER] HTML templates loaded successfully")

	log.Printf("[SERVER] Initializing service layers")
	server := &Server{
		db:                  dbWrapper.DB,
		optionService:       models.NewOptionService(dbWrapper.DB),
		symbolService:       models.NewSymbolService(dbWrapper.DB),
		treasuryService:     models.NewTreasuryService(dbWrapper.DB),
		longPositionService: models.NewLongPositionService(dbWrapper.DB),
		dividendService:     models.NewDividendService(dbWrapper.DB),
		templates:           templates,
	}

	log.Printf("[SERVER] All services initialized successfully")
	log.Printf("[SERVER] Server creation completed")

	return server, nil
}

func (s *Server) setupRoutes() {
	log.Printf("[SERVER] Setting up HTTP routes")

	// Serve static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))
	log.Printf("[SERVER] Route registered: /static/ -> file server")

	http.HandleFunc("/", s.dashboardHandler)
	log.Printf("[SERVER] Route registered: / -> dashboardHandler")

	http.HandleFunc("/monthly", s.monthlyHandler)
	log.Printf("[SERVER] Route registered: /monthly -> monthlyHandler")

	http.HandleFunc("/options", s.optionsHandler)
	log.Printf("[SERVER] Route registered: /options -> optionsHandler")

	http.HandleFunc("/treasuries", s.treasuriesHandler)
	log.Printf("[SERVER] Route registered: /treasuries -> treasuriesHandler")

	http.HandleFunc("/symbol/", s.symbolHandler)
	log.Printf("[SERVER] Route registered: /symbol/ -> symbolHandler")

	http.HandleFunc("/api/premium-data", s.premiumDataHandler)
	log.Printf("[SERVER] Route registered: /api/premium-data -> premiumDataHandler")

	http.HandleFunc("/api/options", s.optionAPIHandler)
	log.Printf("[SERVER] Route registered: /api/options -> optionAPIHandler")

	http.HandleFunc("/api/options/", s.individualOptionAPIHandler)
	log.Printf("[SERVER] Route registered: /api/options/ -> individualOptionAPIHandler")

	http.HandleFunc("/api/symbols/", s.symbolAPIHandler)
	log.Printf("[SERVER] Route registered: /api/symbols/ -> symbolAPIHandler")

	http.HandleFunc("/api/dividends", s.dividendsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/dividends -> dividendsAPIHandler")

	http.HandleFunc("/api/long-positions", s.longPositionsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/long-positions -> longPositionsAPIHandler")

	http.HandleFunc("/api/treasuries/", s.treasuryAPIHandler)
	log.Printf("[SERVER] Route registered: /api/treasuries/ -> treasuryAPIHandler")

	http.HandleFunc("/add-option", s.addOptionHandler)
	log.Printf("[SERVER] Route registered: /add-option -> addOptionHandler")

	http.HandleFunc("/add-treasury", s.addTreasuryHandler)
	log.Printf("[SERVER] Route registered: /add-treasury -> addTreasuryHandler")

	http.HandleFunc("/api/allocation-data", s.allocationDataHandler)
	log.Printf("[SERVER] Route registered: /api/allocation-data -> allocationDataHandler")

	http.HandleFunc("/import", s.HandleImport)
	log.Printf("[SERVER] Route registered: /import -> HandleImport")

	http.HandleFunc("/backup", s.HandleBackup)
	log.Printf("[SERVER] Route registered: /backup -> HandleBackup")

	http.HandleFunc("/backup/", s.HandleBackupFile)
	log.Printf("[SERVER] Route registered: /backup/ -> HandleBackupFile")

	http.HandleFunc("/database/set-current", s.handleSetCurrentDatabase)
	log.Printf("[SERVER] Route registered: /database/set-current -> handleSetCurrentDatabase")

	http.HandleFunc("/database/create", s.handleCreateDatabase)
	log.Printf("[SERVER] Route registered: /database/create -> handleCreateDatabase")

	http.HandleFunc("/database/delete/", s.handleDeleteDatabase)
	log.Printf("[SERVER] Route registered: /database/delete/ -> handleDeleteDatabase")

	http.Handle("/backups/", http.StripPrefix("/backups/", http.FileServer(http.Dir("./data/backups"))))
	log.Printf("[SERVER] Route registered: /backups/ -> file server for backup directory")

	http.HandleFunc("/import/upload", s.HandleImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload -> HandleImportUpload")

	http.HandleFunc("/import/upload/stocks", s.HandleStocksImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload/stocks -> HandleStocksImportUpload")

	http.HandleFunc("/import/upload/dividends", s.HandleDividendsImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload/dividends -> HandleDividendsImportUpload")

	http.HandleFunc("/api/generate-test-data", s.HandleGenerateTestData)
	log.Printf("[SERVER] Route registered: /api/generate-test-data -> HandleGenerateTestData")

	http.HandleFunc("/help", s.helpHandler)
	log.Printf("[SERVER] Route registered: /help -> helpHandler")

	log.Printf("[SERVER] All routes registered successfully")
}

func (s *Server) Start(port string) error {
	s.setupRoutes()

	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Wheeler web application starting on http://localhost:%s\n", port)
	fmt.Printf("   📈 Dashboard:    http://localhost:%s/\n", port)

	return http.ListenAndServe(":"+port, nil)
}

// ExpirationGroup represents a group of positions with the same expiration date
type ExpirationGroup struct {
	Expiration time.Time
	DateStr    string
	Positions  []*models.OpenPositionData
}

// groupPositionsByExpiration groups open positions by expiration date and returns them sorted
func groupPositionsByExpiration(positions []*models.OpenPositionData) []ExpirationGroup {
	grouped := make(map[time.Time][]*models.OpenPositionData)
	
	// Group positions by expiration date
	for _, position := range positions {
		expDate := position.Expiration
		grouped[expDate] = append(grouped[expDate], position)
	}
	
	// Convert to slice and sort by expiration date
	var groups []ExpirationGroup
	for expDate, posGroup := range grouped {
		// Sort positions within each group by symbol, then by strike price
		sort.Slice(posGroup, func(i, j int) bool {
			posI, posJ := posGroup[i], posGroup[j]
			if posI.Symbol != posJ.Symbol {
				return posI.Symbol < posJ.Symbol
			}
			return posI.Strike < posJ.Strike
		})
		
		groups = append(groups, ExpirationGroup{
			Expiration: expDate,
			DateStr:    expDate.Format("01/02/2006"),
			Positions:  posGroup,
		})
	}
	
	// Sort groups by expiration date (earliest first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Expiration.Before(groups[j].Expiration)
	})
	
	return groups
}

func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	log.Printf("[TEMPLATE] Starting template execution for: %s", templateName)
	err := s.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("[TEMPLATE] ERROR: Template execution failed for %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	} else {
		log.Printf("[TEMPLATE] Successfully rendered template: %s", templateName)
	}
}

// getCurrentDatabaseName returns the current database name for template rendering
func (s *Server) getCurrentDatabaseName() string {
	dbName, err := database.GetCurrentDatabase()
	if err != nil {
		log.Printf("[SERVER] Error getting current database name: %v", err)
		return "wheeler.db"
	}
	return dbName
}
