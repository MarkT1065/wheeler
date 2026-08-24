package web

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"stonks/internal/database"
	"stonks/internal/models"
	"stonks/internal/providers"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	servicesMu                   sync.RWMutex
	db                           *sql.DB
	optionService                *models.OptionService
	symbolService                *models.SymbolService
	treasuryService              *models.TreasuryService
	longPositionService          *models.LongPositionService
	dividendService              *models.DividendService
	settingService               *models.SettingService
	metricService                *models.MetricService
	providerService              *providers.Service
	serviceLeases                map[*sql.DB]int
	retiredDatabases             map[*sql.DB]struct{}
	providerValidation           map[string]providerValidationCacheEntry
	providerValidationInFlight   map[string]*providerValidationCall
	providerValidationGeneration uint64
	providerValidationMu         sync.Mutex
	providerValidationWaitHook   func()
	templates                    *template.Template
}

const providerValidationCacheTTL = time.Minute

type providerValidationCacheEntry struct {
	checkedAt time.Time
	valid     bool
	err       string
}

type providerValidationCall struct {
	done       chan struct{}
	generation uint64
}

type providerServicesSnapshot struct {
	providerService *providers.Service
	symbolService   *models.SymbolService
}

type providerServicesContextKey struct{}

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
		"div": func(a, b interface{}) interface{} {
			switch aVal := a.(type) {
			case int:
				if bVal, ok := b.(int); ok && bVal != 0 {
					return float64(aVal) / float64(bVal)
				}
				if bVal, ok := b.(float64); ok && bVal != 0.0 {
					return float64(aVal) / bVal
				}
			case float64:
				if bVal, ok := b.(int); ok && bVal != 0 {
					return aVal / float64(bVal)
				}
				if bVal, ok := b.(float64); ok && bVal != 0.0 {
					return aVal / bVal
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
		"formatCurrencyWithDecimals": func(value interface{}) string {
			var floatVal float64
			var err error
			switch v := value.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case string:
				floatVal, err = strconv.ParseFloat(v, 64)
				if err != nil {
					return "$0.00"
				}
			default:
				return "$0.00"
			}

			// Format to 2 decimal places
			formatted := fmt.Sprintf("%.2f", floatVal)

			// Split into integer and decimal parts
			parts := strings.Split(formatted, ".")
			intPart := parts[0]
			decPart := parts[1]

			// Handle negative numbers
			isNegative := false
			if strings.HasPrefix(intPart, "-") {
				isNegative = true
				intPart = intPart[1:]
			}

			// Add commas to integer part
			if len(intPart) > 3 {
				var result string
				for i, digit := range intPart {
					if i > 0 && (len(intPart)-i)%3 == 0 {
						result += ","
					}
					result += string(digit)
				}
				intPart = result
			}

			// Combine with decimals
			formatted = intPart + "." + decPart
			if isNegative {
				return "-$" + formatted
			}
			return "$" + formatted
		},
		"formatInt": func(value interface{}) string {
			var intVal int
			switch v := value.(type) {
			case int:
				intVal = v
			case float64:
				intVal = int(v)
			default:
				return "0"
			}

			str := fmt.Sprintf("%d", intVal)
			isNegative := false
			if intVal < 0 {
				isNegative = true
				str = str[1:]
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

			if isNegative {
				return "-" + str
			}
			return str
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
		db:        dbWrapper.DB,
		templates: templates,
	}
	server.initializeServices(dbWrapper.DB)

	log.Printf("[SERVER] All services initialized successfully")
	log.Printf("[SERVER] Server creation completed")

	return server, nil
}

// initializeServices binds every server service to the supplied database.
// It is used at startup and after a database switch to avoid stale service
// pointers, including provider configuration from the previous database.
func (s *Server) initializeServices(db *sql.DB) {
	s.servicesMu.Lock()
	defer s.servicesMu.Unlock()
	s.initializeServicesLocked(db)
	s.invalidateProviderValidation()
}

func (s *Server) initializeServicesLocked(db *sql.DB) {
	s.db = db
	s.optionService = models.NewOptionService(db)
	s.symbolService = models.NewSymbolService(db)
	s.treasuryService = models.NewTreasuryService(db)
	s.longPositionService = models.NewLongPositionService(db)
	s.dividendService = models.NewDividendService(db)
	s.settingService = models.NewSettingService(db)
	s.metricService = models.NewMetricService(db)
	s.providerService = providers.NewService(s.symbolService, s.settingService)
}

func (s *Server) cachedProviderValidation(provider string, validate func() error) (bool, string) {
	for {
		s.providerValidationMu.Lock()
		generation := s.providerValidationGeneration
		// A fresh entry is safe to return immediately. Cache invalidation bumps
		// the generation, so entries from an earlier provider configuration are
		// never reused after a settings or database change.
		if entry, ok := s.providerValidation[provider]; ok && time.Since(entry.checkedAt) < providerValidationCacheTTL {
			s.providerValidationMu.Unlock()
			return entry.valid, entry.err
		}
		// One caller validates outside the mutex. Other callers wait for that
		// result instead of serializing multiple provider HTTP requests.
		if call, ok := s.providerValidationInFlight[provider]; ok && call.generation == generation {
			done := call.done
			waitHook := s.providerValidationWaitHook
			s.providerValidationMu.Unlock()
			if waitHook != nil {
				waitHook()
			}
			<-done
			continue
		}
		if s.providerValidationInFlight == nil {
			s.providerValidationInFlight = make(map[string]*providerValidationCall)
		}
		call := &providerValidationCall{done: make(chan struct{}), generation: generation}
		s.providerValidationInFlight[provider] = call
		s.providerValidationMu.Unlock()

		// This request owns the validation for this provider and generation.
		// Never hold providerValidationMu across the potentially slow HTTP call.
		entry := providerValidationCacheEntry{checkedAt: time.Now()}
		if err := validate(); err != nil {
			entry.err = err.Error()
		} else {
			entry.valid = true
		}

		s.providerValidationMu.Lock()
		if s.providerValidationGeneration == generation {
			if s.providerValidation == nil {
				s.providerValidation = make(map[string]providerValidationCacheEntry)
			}
			s.providerValidation[provider] = entry
		}
		if s.providerValidationInFlight[provider] == call {
			delete(s.providerValidationInFlight, provider)
			close(call.done)
		}
		isCurrentGeneration := s.providerValidationGeneration == generation
		s.providerValidationMu.Unlock()
		if isCurrentGeneration {
			return entry.valid, entry.err
		}
		// Configuration changed while validation was in progress. Repeat using
		// the current generation rather than returning a stale result.
	}
}

// invalidateProviderValidation makes cached and in-flight validations from a
// previous configuration obsolete. In-flight callers wake their waiters, which
// retry against this newer generation.
func (s *Server) invalidateProviderValidation() {
	s.providerValidationMu.Lock()
	defer s.providerValidationMu.Unlock()
	s.providerValidation = make(map[string]providerValidationCacheEntry)
	s.providerValidationGeneration++
}

// Close closes the database connection
func (s *Server) Close() error {
	s.servicesMu.Lock()
	defer s.servicesMu.Unlock()
	if s.db != nil {
		log.Printf("[SERVER] Closing database connection")
		return s.db.Close()
	}
	return nil
}

func (s *Server) handleServiceFunc(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		s.servicesMu.RLock()
		defer s.servicesMu.RUnlock()
		handler(w, r)
	})
}

// handleProviderServiceFunc leases the current provider services for long bulk
// requests. The lease keeps their database open after a switch, but does not
// hold servicesMu while network and rate-limit waits are in progress.
func (s *Server) handleProviderServiceFunc(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		snapshot, release := s.acquireProviderServices()
		defer release()
		ctx := context.WithValue(r.Context(), providerServicesContextKey{}, snapshot)
		handler(w, r.WithContext(ctx))
	})
}

func (s *Server) acquireProviderServices() (providerServicesSnapshot, func()) {
	s.servicesMu.Lock()
	db := s.db
	if s.serviceLeases == nil {
		s.serviceLeases = make(map[*sql.DB]int)
	}
	s.serviceLeases[db]++
	snapshot := providerServicesSnapshot{providerService: s.providerService, symbolService: s.symbolService}
	s.servicesMu.Unlock()

	return snapshot, func() {
		s.servicesMu.Lock()
		defer s.servicesMu.Unlock()
		s.serviceLeases[db]--
		if s.serviceLeases[db] != 0 {
			return
		}
		delete(s.serviceLeases, db)
		if _, retired := s.retiredDatabases[db]; retired {
			delete(s.retiredDatabases, db)
			if err := db.Close(); err != nil {
				log.Printf("[SERVER] Error closing retired database: %v", err)
			}
		}
	}
}

func providerServicesFromContext(ctx context.Context) (providerServicesSnapshot, bool) {
	snapshot, ok := ctx.Value(providerServicesContextKey{}).(providerServicesSnapshot)
	return snapshot, ok
}

// switchServices atomically replaces the service set. A database leased by a
// long provider request is closed by that request's release function instead
// of delaying the switch or invalidating the in-flight request.
func (s *Server) switchServices(db *sql.DB) error {
	s.servicesMu.Lock()
	defer s.servicesMu.Unlock()

	oldDB := s.db
	s.initializeServicesLocked(db)
	s.invalidateProviderValidation()
	if oldDB != nil && oldDB != db {
		if s.serviceLeases[oldDB] > 0 {
			if s.retiredDatabases == nil {
				s.retiredDatabases = make(map[*sql.DB]struct{})
			}
			s.retiredDatabases[oldDB] = struct{}{}
			return nil
		}
		return oldDB.Close()
	}
	return nil
}

func (s *Server) setupRoutes() {
	log.Printf("[SERVER] Setting up HTTP routes")

	// Serve static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))))
	log.Printf("[SERVER] Route registered: /static/ -> file server")

	s.handleServiceFunc("/", s.dashboardHandler)
	log.Printf("[SERVER] Route registered: / -> dashboardHandler")

	s.handleServiceFunc("/monthly", s.monthlyHandler)
	log.Printf("[SERVER] Route registered: /monthly -> monthlyHandler")

	s.handleServiceFunc("/options", s.optionsHandler)
	log.Printf("[SERVER] Route registered: /options -> optionsHandler")

	s.handleServiceFunc("/all-options", s.allOptionsHandler)
	log.Printf("[SERVER] Route registered: /all-options -> allOptionsHandler")

	s.handleServiceFunc("/treasuries", s.treasuriesHandler)
	log.Printf("[SERVER] Route registered: /treasuries -> treasuriesHandler")

	s.handleServiceFunc("/dividends", s.dividendsHandler)
	log.Printf("[SERVER] Route registered: /dividends -> dividendsHandler")

	s.handleServiceFunc("/metrics", s.metricsHandler)
	log.Printf("[SERVER] Route registered: /metrics -> metricsHandler")

	s.handleServiceFunc("/zen", s.zenHandler)
	log.Printf("[SERVER] Route registered: /zen -> zenHandler")

	s.handleServiceFunc("/symbol/", s.symbolHandler)
	log.Printf("[SERVER] Route registered: /symbol/ -> symbolHandler")

	s.handleServiceFunc("/api/premium-data", s.premiumDataHandler)
	log.Printf("[SERVER] Route registered: /api/premium-data -> premiumDataHandler")

	s.handleServiceFunc("/api/options", s.optionAPIHandler)
	log.Printf("[SERVER] Route registered: /api/options -> optionAPIHandler")

	s.handleServiceFunc("/api/options/", s.individualOptionAPIHandler)
	log.Printf("[SERVER] Route registered: /api/options/ -> individualOptionAPIHandler")

	s.handleServiceFunc("/api/options/filter", s.optionsFilterHandler)
	log.Printf("[SERVER] Route registered: /api/options/filter -> optionsFilterHandler")

	s.handleServiceFunc("/api/symbols/", s.symbolAPIHandler)
	log.Printf("[SERVER] Route registered: /api/symbols/ -> symbolAPIHandler")

	s.handleServiceFunc("/api/dividends", s.dividendsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/dividends -> dividendsAPIHandler")

	s.handleServiceFunc("/api/long-positions", s.longPositionsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/long-positions -> longPositionsAPIHandler")

	s.handleServiceFunc("/api/treasuries/", s.treasuryAPIHandler)
	log.Printf("[SERVER] Route registered: /api/treasuries/ -> treasuryAPIHandler")

	s.handleServiceFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getMetricsHandler(w, r)
		case http.MethodPost:
			s.createMetricHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Printf("[SERVER] Route registered: /api/metrics -> metrics API handler")

	s.handleServiceFunc("/api/metrics/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			s.updateMetricHandler(w, r)
		case http.MethodDelete:
			s.deleteMetricHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Printf("[SERVER] Route registered: /api/metrics/ -> individual metric API handler")

	s.handleServiceFunc("/api/metrics/snapshot", s.createMetricsSnapshotHandler)
	log.Printf("[SERVER] Route registered: /api/metrics/snapshot -> createMetricsSnapshotHandler")

	s.handleServiceFunc("/api/metrics/chart-data", s.getMetricsChartDataHandler)
	log.Printf("[SERVER] Route registered: /api/metrics/chart-data -> getMetricsChartDataHandler")

	s.handleServiceFunc("/add-option", s.addOptionHandler)
	log.Printf("[SERVER] Route registered: /add-option -> addOptionHandler")

	s.handleServiceFunc("/add-treasury", s.addTreasuryHandler)
	log.Printf("[SERVER] Route registered: /add-treasury -> addTreasuryHandler")

	s.handleServiceFunc("/api/allocation-data", s.allocationDataHandler)
	log.Printf("[SERVER] Route registered: /api/allocation-data -> allocationDataHandler")

	s.handleServiceFunc("/api/optionable-positions", s.optionablePositionsHandler)
	log.Printf("[SERVER] Route registered: /api/optionable-positions -> optionablePositionsHandler")

	s.handleServiceFunc("/import", s.HandleImport)
	log.Printf("[SERVER] Route registered: /import -> HandleImport")

	s.handleServiceFunc("/backup", s.HandleBackup)
	log.Printf("[SERVER] Route registered: /backup -> HandleBackup")

	s.handleServiceFunc("/backup/", s.HandleBackupFile)
	log.Printf("[SERVER] Route registered: /backup/ -> HandleBackupFile")

	http.HandleFunc("/database/set-current", s.handleSetCurrentDatabase)
	log.Printf("[SERVER] Route registered: /database/set-current -> handleSetCurrentDatabase")

	s.handleServiceFunc("/database/create", s.handleCreateDatabase)
	log.Printf("[SERVER] Route registered: /database/create -> handleCreateDatabase")

	s.handleServiceFunc("/database/delete/", s.handleDeleteDatabase)
	log.Printf("[SERVER] Route registered: /database/delete/ -> handleDeleteDatabase")

	http.Handle("/backups/", http.StripPrefix("/backups/", http.FileServer(http.Dir("./data/backups"))))
	log.Printf("[SERVER] Route registered: /backups/ -> file server for backup directory")

	s.handleServiceFunc("/import/upload", s.HandleImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload -> HandleImportUpload")

	s.handleServiceFunc("/import/upload/stocks", s.HandleStocksImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload/stocks -> HandleStocksImportUpload")

	s.handleServiceFunc("/import/upload/dividends", s.HandleDividendsImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload/dividends -> HandleDividendsImportUpload")

	s.handleServiceFunc("/import/upload/treasuries", s.HandleTreasuriesImportUpload)
	log.Printf("[SERVER] Route registered: /import/upload/treasuries -> HandleTreasuriesImportUpload")

	s.handleServiceFunc("/api/generate-test-data", s.HandleGenerateTestData)
	log.Printf("[SERVER] Route registered: /api/generate-test-data -> HandleGenerateTestData")

	s.handleServiceFunc("/help", s.helpHandler)
	log.Printf("[SERVER] Route registered: /help -> helpHandler")

	s.handleServiceFunc("/settings", s.settingsHandler)
	log.Printf("[SERVER] Route registered: /settings -> settingsHandler")

	s.handleServiceFunc("/api/settings", s.settingsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/settings -> settingsAPIHandler")

	s.handleServiceFunc("/api/settings/", s.individualSettingAPIHandler)
	log.Printf("[SERVER] Route registered: /api/settings/ -> individualSettingAPIHandler")

	s.handleServiceFunc("/api/provider/configuration", s.providerConfigurationAPIHandler)
	log.Printf("[SERVER] Route registered: /api/provider/configuration -> providerConfigurationAPIHandler")

	s.handleServiceFunc("/api/markets", s.marketsAPIHandler)
	log.Printf("[SERVER] Route registered: /api/markets -> marketsAPIHandler")

	s.handleServiceFunc("/api/provider/test", s.providerTestHandler)
	log.Printf("[SERVER] Route registered: /api/provider/test -> providerTestHandler")

	s.handleProviderServiceFunc("/api/provider/update-prices", s.providerUpdatePricesHandler)
	log.Printf("[SERVER] Route registered: /api/provider/update-prices -> providerUpdatePricesHandler")

	s.handleServiceFunc("/api/provider/symbol-info/", s.providerSymbolInfoHandler)
	log.Printf("[SERVER] Route registered: /api/provider/symbol-info/ -> providerSymbolInfoHandler")

	s.handleServiceFunc("/api/provider/status", s.providerStatusHandler)
	log.Printf("[SERVER] Route registered: /api/provider/status -> providerStatusHandler")

	s.handleProviderServiceFunc("/api/provider/fetch-dividends", s.providerFetchDividendsHandler)
	log.Printf("[SERVER] Route registered: /api/provider/fetch-dividends -> providerFetchDividendsHandler")

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

// SetupTestRoutes sets up routes for testing purposes
func (s *Server) SetupTestRoutes() {
	s.setupRoutes()
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

	// Use a buffer to execute template first, then write to response if successful
	var buf bytes.Buffer
	err := s.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		log.Printf("[TEMPLATE] ERROR: Template execution failed for %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Template executed successfully, write to response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Printf("[TEMPLATE] ERROR: Failed to write response for %s: %v", templateName, err)
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

// getAllSymbolsList returns a list of all distinct symbols for navigation
func (s *Server) getAllSymbolsList() []string {
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[SERVER] Error getting symbols list: %v", err)
		return []string{}
	}
	return symbols
}
