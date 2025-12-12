package web

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"stonks/internal/database"
	"stonks/internal/models"
	"strconv"
	"strings"
	"time"
)

// HandleImport renders the import page
func (s *Server) HandleImport(w http.ResponseWriter, r *http.Request) {
	log.Printf("[IMPORT] Rendering import page")

	// Get all symbols for navigation
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[IMPORT] Error getting symbols: %v", err)
		symbols = []string{}
	}

	data := ImportData{
		Symbols:    symbols,
		AllSymbols: symbols, // For navigation compatibility
		CurrentDB:  s.getCurrentDatabaseName(),
		ActivePage: "import",
	}

	s.renderTemplate(w, "import.html", data)
}

// HandleBackup renders the backup page and lists available database files
func (s *Server) HandleBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.handleCreateBackup(w, r)
		return
	}

	log.Printf("[BACKUP] Rendering backup page")

	// Get all symbols for navigation
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[BACKUP] Error getting symbols: %v", err)
		symbols = []string{}
	}

	// Get current working directory and list .db files (active databases)
	dbFiles, err := s.getAvailableDbFiles()
	if err != nil {
		log.Printf("[BACKUP] Error getting database files: %v", err)
		dbFiles = []string{}
	}

	// Get backup files from ./backups directory
	backupFiles, err := s.getBackupFiles()
	if err != nil {
		log.Printf("[BACKUP] Error getting backup files: %v", err)
		backupFiles = []string{}
	}

	// Get current database name
	currentDB := s.getCurrentDatabaseName()

	data := BackupData{
		AllSymbols:  symbols,
		DbFiles:     dbFiles,
		BackupFiles: backupFiles,
		CurrentDB:   currentDB,
		ActivePage:  "backup",
	}

	if err := s.templates.ExecuteTemplate(w, "backup.html", data); err != nil {
		log.Printf("[BACKUP] Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleImportUpload processes the CSV file upload and imports options
func (s *Server) HandleImportUpload(w http.ResponseWriter, r *http.Request) {
	log.Printf("[IMPORT] Processing CSV upload")

	// Set response header for JSON
	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form (10MB max)
	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		log.Printf("[IMPORT] Error parsing multipart form: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to parse upload form",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the uploaded file
	file, fileHeader, err := r.FormFile("csvFile")
	if err != nil {
		log.Printf("[IMPORT] Error getting uploaded file: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "No file uploaded or invalid file",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()

	log.Printf("[IMPORT] Processing file: %s (size: %d bytes)", fileHeader.Filename, fileHeader.Size)

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".csv") {
		log.Printf("[IMPORT] Invalid file type: %s", fileHeader.Filename)
		response := ImportResponse{
			Success: false,
			Error:   "File must be a CSV file",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse CSV and import options
	importedCount, skippedCount, err := s.importOptionsFromCSV(file)
	if err != nil {
		log.Printf("[IMPORT] Error importing options: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to import options",
			Details: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("[IMPORT] Import completed: %d imported, %d skipped", importedCount, skippedCount)
	response := ImportResponse{
		Success:       true,
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleStocksImportUpload processes the stocks CSV file upload and imports long positions
func (s *Server) HandleStocksImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[STOCKS_IMPORT] Starting stocks CSV import")

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("[STOCKS_IMPORT] Error parsing multipart form: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to parse form data",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	file, _, err := r.FormFile("csvFile")
	if err != nil {
		log.Printf("[STOCKS_IMPORT] Error getting form file: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "No file provided or error reading file",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()

	// Import stocks from CSV
	importedCount, skippedCount, err := s.importStocksFromCSV(file)
	if err != nil {
		log.Printf("[STOCKS_IMPORT] Import failed: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to import stocks from CSV",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("[STOCKS_IMPORT] Import completed: %d imported, %d skipped", importedCount, skippedCount)
	response := ImportResponse{
		Success:       true,
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDividendsImportUpload processes the dividends CSV file upload and imports dividend records
func (s *Server) HandleDividendsImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[DIVIDENDS_IMPORT] Starting dividends CSV import")

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("[DIVIDENDS_IMPORT] Error parsing multipart form: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to parse form data",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	file, _, err := r.FormFile("csvFile")
	if err != nil {
		log.Printf("[DIVIDENDS_IMPORT] Error getting form file: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "No file provided or error reading file",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()

	// Import dividends from CSV
	importedCount, skippedCount, err := s.importDividendsFromCSV(file)
	if err != nil {
		log.Printf("[DIVIDENDS_IMPORT] Import failed: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to import dividends from CSV",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("[DIVIDENDS_IMPORT] Import completed: %d imported, %d skipped", importedCount, skippedCount)
	response := ImportResponse{
		Success:       true,
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTreasuriesImportUpload processes the treasuries CSV file upload and imports treasury records
func (s *Server) HandleTreasuriesImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[TREASURIES_IMPORT] Starting treasuries CSV import")

	// Parse multipart form (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("[TREASURIES_IMPORT] Error parsing multipart form: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to parse form data",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	file, _, err := r.FormFile("csvFile")
	if err != nil {
		log.Printf("[TREASURIES_IMPORT] Error getting form file: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "No file provided or error reading file",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer file.Close()

	// Import treasuries from CSV
	importedCount, skippedCount, err := s.importTreasuriesFromCSV(file)
	if err != nil {
		log.Printf("[TREASURIES_IMPORT] Import failed: %v", err)
		response := ImportResponse{
			Success: false,
			Error:   "Failed to import treasuries from CSV",
			Details: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("[TREASURIES_IMPORT] Import completed: %d imported, %d skipped", importedCount, skippedCount)
	response := ImportResponse{
		Success:       true,
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// importOptionsFromCSV parses the CSV file and imports options
func (s *Server) importOptionsFromCSV(file io.Reader) (importedCount int, skippedCount int, err error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 10 // Expect exactly 10 fields

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Validate headers (accept both 'commission' and 'total_commission' for backward compatibility)
	expectedHeaders := []string{"symbol", "opened", "closed", "type", "strike", "expiration", "premium", "contracts", "exit_price", "commission"}
	if len(headers) != len(expectedHeaders) {
		return 0, 0, fmt.Errorf("CSV must have exactly %d columns, got %d", len(expectedHeaders), len(headers))
	}

	for i, expected := range expectedHeaders {
		header := strings.TrimSpace(strings.ToLower(headers[i]))
		// Accept both 'commission' and 'total_commission' for the last column
		if i == 9 && (header == "commission" || header == "total_commission") {
			continue
		}
		if header != expected {
			return 0, 0, fmt.Errorf("column %d should be '%s', got '%s'", i+1, expected, headers[i])
		}
	}

	log.Printf("[IMPORT] CSV headers validated successfully")

	// Process data rows
	rowNumber := 1 // Start at 1 since we already read the header
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return importedCount, skippedCount, fmt.Errorf("error reading row %d: %w", rowNumber+1, err)
		}
		rowNumber++

		// Parse and validate the record
		csvRecord := CSVOptionRecord{
			Symbol:     strings.TrimSpace(strings.ToUpper(record[0])),
			Opened:     strings.TrimSpace(record[1]),
			Closed:     strings.TrimSpace(record[2]),
			Type:       strings.TrimSpace(record[3]),
			Strike:     strings.TrimSpace(record[4]),
			Expiration: strings.TrimSpace(record[5]),
			Premium:    strings.TrimSpace(record[6]),
			Contracts:  strings.TrimSpace(record[7]),
			ExitPrice:  strings.TrimSpace(record[8]),
			Commission: strings.TrimSpace(record[9]),
		}

		// Convert to Option struct
		option, err := s.convertCSVRecordToOption(csvRecord, rowNumber)
		if err != nil {
			return importedCount, skippedCount, fmt.Errorf("error processing row %d: %w", rowNumber, err)
		}

		// Ensure symbol exists (create if it doesn't)
		err = s.ensureSymbolExists(option.Symbol)
		if err != nil {
			return importedCount, skippedCount, fmt.Errorf("error ensuring symbol exists for row %d: %w", rowNumber, err)
		}

		// Try to create the option (skip if duplicate) - use CreateWithCommission to set custom commission
		_, err = s.optionService.CreateWithCommission(option.Symbol, option.Type, option.Opened, option.Strike, option.Expiration, option.Premium, option.Contracts, option.Commission)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate") {
				log.Printf("[IMPORT] Skipping duplicate option at row %d: %s %s %v", rowNumber, option.Symbol, option.Type, option.Opened)
				skippedCount++
				continue
			}
			return importedCount, skippedCount, fmt.Errorf("error creating option at row %d: %w", rowNumber, err)
		}

		// If the option was closed, update it with exit information
		if option.Closed != nil {
			// We need to get the created option to update it
			options, err := s.optionService.GetBySymbol(option.Symbol)
			if err == nil {
				// Find the option we just created (match by key fields)
				for _, opt := range options {
					if opt.Symbol == option.Symbol && opt.Type == option.Type &&
						opt.Opened.Equal(option.Opened) && opt.Strike == option.Strike &&
						opt.Expiration.Equal(option.Expiration) && opt.Premium == option.Premium &&
						opt.Contracts == option.Contracts {
						_, updateErr := s.optionService.UpdateByID(opt.ID, opt.Symbol, opt.Type, opt.Opened, opt.Strike, opt.Expiration, opt.Premium, opt.Contracts, opt.Commission, option.Closed, option.ExitPrice)
						if updateErr != nil {
							log.Printf("[IMPORT] Warning: Failed to update option exit info for row %d: %v", rowNumber, updateErr)
						}
						break
					}
				}
			}
		}

		importedCount++
		if importedCount%10 == 0 {
			log.Printf("[IMPORT] Progress: %d options imported so far", importedCount)
		}
	}

	return importedCount, skippedCount, nil
}

// importStocksFromCSV parses the CSV file and imports stock positions
func (s *Server) importStocksFromCSV(file io.Reader) (importedCount int, skippedCount int, err error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 6 // Expect exactly 6 fields

	records, err := reader.ReadAll()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return 0, 0, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	if len(records) <= 1 {
		return 0, 0, fmt.Errorf("CSV file must contain data rows beyond the header")
	}

	log.Printf("[STOCKS_IMPORT] Processing %d stock records", len(records)-1)

	for i, record := range records[1:] { // Skip header row
		if len(record) != 6 {
			log.Printf("[STOCKS_IMPORT] Row %d: Invalid column count (expected 6, got %d)", i+2, len(record))
			return importedCount, skippedCount, fmt.Errorf("row %d: expected 6 columns, got %d", i+2, len(record))
		}

		csvRecord := CSVStockRecord{
			Symbol:     strings.TrimSpace(record[0]),
			Purchased:  strings.TrimSpace(record[1]),
			ClosedDate: strings.TrimSpace(record[2]),
			Shares:     strings.TrimSpace(record[3]),
			BuyPrice:   strings.TrimSpace(record[4]),
			ExitPrice:  strings.TrimSpace(record[5]),
		}

		// Convert CSV record to LongPosition
		position, err := s.csvStockRecordToLongPosition(csvRecord)
		if err != nil {
			log.Printf("[STOCKS_IMPORT] Row %d: Failed to convert record: %v", i+2, err)
			return importedCount, skippedCount, fmt.Errorf("row %d: %w", i+2, err)
		}

		// Ensure symbol exists
		if err := s.ensureSymbolExists(position.Symbol); err != nil {
			log.Printf("[STOCKS_IMPORT] Row %d: Failed to ensure symbol exists: %v", i+2, err)
			return importedCount, skippedCount, fmt.Errorf("row %d: failed to create symbol: %w", i+2, err)
		}

		// Create long position
		_, err = s.longPositionService.Create(
			position.Symbol,
			position.Opened,
			position.Shares,
			position.BuyPrice,
		)

		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				log.Printf("[STOCKS_IMPORT] Row %d: Duplicate stock position skipped", i+2)
				skippedCount++
				continue
			}
			log.Printf("[STOCKS_IMPORT] Row %d: Failed to create position: %v", i+2, err)
			return importedCount, skippedCount, fmt.Errorf("row %d: failed to create position: %w", i+2, err)
		}

		// If position was closed, update with exit data
		if position.Closed != nil && position.ExitPrice != nil {
			// We need to find the position we just created and update it
			positions, err := s.longPositionService.GetBySymbol(position.Symbol)
			if err == nil && len(positions) > 0 {
				// Find the most recent position (highest ID)
				var latestPosition *models.LongPosition
				for _, p := range positions {
					if latestPosition == nil || p.ID > latestPosition.ID {
						latestPosition = p
					}
				}

				if latestPosition != nil {
					_, updateErr := s.longPositionService.UpdateByID(
						latestPosition.ID,
						position.Symbol,
						position.Opened,
						position.Shares,
						position.BuyPrice,
						position.Closed,
						position.ExitPrice,
					)
					if updateErr != nil {
						log.Printf("[STOCKS_IMPORT] Row %d: Failed to update position with exit data: %v", i+2, updateErr)
					}
				}
			}
		}

		importedCount++
		log.Printf("[STOCKS_IMPORT] Row %d: Successfully imported %s position", i+2, position.Symbol)
	}

	return importedCount, skippedCount, nil
}

// importDividendsFromCSV parses the CSV file and imports dividend records
func (s *Server) importDividendsFromCSV(file io.Reader) (importedCount int, skippedCount int, err error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3 // Expect exactly 3 fields: Symbol, Date Received, Amount

	records, err := reader.ReadAll()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return 0, 0, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	if len(records) <= 1 {
		return 0, 0, fmt.Errorf("CSV file must contain data rows beyond the header")
	}

	log.Printf("[DIVIDENDS_IMPORT] Processing %d dividend records", len(records)-1)

	for i, record := range records[1:] { // Skip header row
		if len(record) != 3 {
			log.Printf("[DIVIDENDS_IMPORT] Row %d: Invalid column count (expected 3, got %d)", i+2, len(record))
			return importedCount, skippedCount, fmt.Errorf("row %d: expected 3 columns, got %d", i+2, len(record))
		}

		csvRecord := CSVDividendRecord{
			Symbol:       strings.TrimSpace(record[0]),
			DateReceived: strings.TrimSpace(record[1]),
			Amount:       strings.TrimSpace(record[2]),
		}

		dividend, created, err := s.processDividendRecord(csvRecord, i+2)
		if err != nil {
			log.Printf("[DIVIDENDS_IMPORT] Row %d: %v", i+2, err)
			return importedCount, skippedCount, err
		}

		if created {
			importedCount++
			log.Printf("[DIVIDENDS_IMPORT] Row %d: Created dividend %s %.2f on %s",
				i+2, dividend.Symbol, dividend.Amount, dividend.Received.Format("2006-01-02"))
		} else {
			skippedCount++
			log.Printf("[DIVIDENDS_IMPORT] Row %d: Skipped duplicate dividend %s %.2f on %s",
				i+2, dividend.Symbol, dividend.Amount, dividend.Received.Format("2006-01-02"))
		}
	}

	return importedCount, skippedCount, nil
}

// importTreasuriesFromCSV parses the CSV file and imports treasury records
func (s *Server) importTreasuriesFromCSV(file io.Reader) (importedCount int, skippedCount int, err error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 8 // Expect exactly 8 fields: CUSPID, Purchased, Maturity, Amount, Yield, BuyPrice, CurrentValue, ExitPrice

	records, err := reader.ReadAll()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return 0, 0, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	if len(records) <= 1 {
		return 0, 0, fmt.Errorf("CSV file must contain data rows beyond the header")
	}

	log.Printf("[TREASURIES_IMPORT] Processing %d treasury records", len(records)-1)

	for i, record := range records[1:] { // Skip header row
		if len(record) != 8 {
			log.Printf("[TREASURIES_IMPORT] Row %d: Invalid column count (expected 8, got %d)", i+2, len(record))
			return importedCount, skippedCount, fmt.Errorf("row %d: expected 8 columns, got %d", i+2, len(record))
		}

		csvRecord := CSVTreasuryRecord{
			CUSPID:       strings.TrimSpace(record[0]),
			Purchased:    strings.TrimSpace(record[1]),
			Maturity:     strings.TrimSpace(record[2]),
			Amount:       strings.TrimSpace(record[3]),
			Yield:        strings.TrimSpace(record[4]),
			BuyPrice:     strings.TrimSpace(record[5]),
			CurrentValue: strings.TrimSpace(record[6]),
			ExitPrice:    strings.TrimSpace(record[7]),
		}

		treasury, created, err := s.processTreasuryRecord(csvRecord, i+2)
		if err != nil {
			log.Printf("[TREASURIES_IMPORT] Row %d: %v", i+2, err)
			return importedCount, skippedCount, err
		}

		if created {
			importedCount++
			log.Printf("[TREASURIES_IMPORT] Row %d: Created treasury %s %.2f purchased on %s",
				i+2, treasury.CUSPID, treasury.Amount, treasury.Purchased.Format("2006-01-02"))
		} else {
			skippedCount++
			log.Printf("[TREASURIES_IMPORT] Row %d: Skipped duplicate treasury %s %.2f purchased on %s",
				i+2, treasury.CUSPID, treasury.Amount, treasury.Purchased.Format("2006-01-02"))
		}
	}

	return importedCount, skippedCount, nil
}

// convertCSVRecordToOption converts a CSV record to an Option struct
func (s *Server) convertCSVRecordToOption(record CSVOptionRecord, rowNumber int) (*models.Option, error) {
	// Validate required fields
	if record.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if record.Opened == "" {
		return nil, fmt.Errorf("opened date is required")
	}
	if record.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if record.Type != "Put" && record.Type != "Call" {
		return nil, fmt.Errorf("type must be 'Put' or 'Call', got '%s'", record.Type)
	}

	// Parse dates
	opened, err := time.Parse("2006-01-02", record.Opened)
	if err != nil {
		return nil, fmt.Errorf("invalid opened date format (must be YYYY-MM-DD): %w", err)
	}

	expiration, err := time.Parse("2006-01-02", record.Expiration)
	if err != nil {
		return nil, fmt.Errorf("invalid expiration date format (must be YYYY-MM-DD): %w", err)
	}

	var closed *time.Time
	if record.Closed != "" {
		closedDate, err := time.Parse("2006-01-02", record.Closed)
		if err != nil {
			return nil, fmt.Errorf("invalid closed date format (must be YYYY-MM-DD): %w", err)
		}
		closed = &closedDate
	}

	// Parse numeric fields
	strike, err := strconv.ParseFloat(record.Strike, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid strike price: %w", err)
	}

	premium, err := strconv.ParseFloat(record.Premium, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid premium: %w", err)
	}

	contracts, err := strconv.Atoi(record.Contracts)
	if err != nil {
		return nil, fmt.Errorf("invalid contracts count: %w", err)
	}

	commission, err := strconv.ParseFloat(record.Commission, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid commission: %w", err)
	}

	var exitPrice *float64
	if record.ExitPrice != "" {
		price, err := strconv.ParseFloat(record.ExitPrice, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid exit price: %w", err)
		}
		exitPrice = &price
	}

	// Validate business logic
	if strike <= 0 {
		return nil, fmt.Errorf("strike price must be positive")
	}
	if premium < 0 {
		return nil, fmt.Errorf("premium cannot be negative")
	}
	if contracts <= 0 {
		return nil, fmt.Errorf("contracts must be positive")
	}
	if commission < 0 {
		return nil, fmt.Errorf("commission cannot be negative")
	}
	if expiration.Before(opened) {
		return nil, fmt.Errorf("expiration date cannot be before opened date")
	}
	if closed != nil && closed.Before(opened) {
		return nil, fmt.Errorf("closed date cannot be before opened date")
	}

	option := &models.Option{
		Symbol:     record.Symbol,
		Type:       record.Type,
		Opened:     opened,
		Closed:     closed,
		Strike:     strike,
		Expiration: expiration,
		Premium:    premium,
		Contracts:  contracts,
		ExitPrice:  exitPrice,
		Commission: commission,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return option, nil
}

// csvStockRecordToLongPosition converts a CSV stock record to a LongPosition
func (s *Server) csvStockRecordToLongPosition(record CSVStockRecord) (*models.LongPosition, error) {
	// Validate required fields
	if record.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if record.Purchased == "" {
		return nil, fmt.Errorf("purchased date is required")
	}
	if record.Shares == "" {
		return nil, fmt.Errorf("shares is required")
	}
	if record.BuyPrice == "" {
		return nil, fmt.Errorf("buy price is required")
	}

	// Parse purchased date (MM/DD/YYYY format)
	purchased, err := time.Parse("1/2/2006", record.Purchased)
	if err != nil {
		// Try MM/DD/YYYY format
		purchased, err = time.Parse("01/02/2006", record.Purchased)
		if err != nil {
			return nil, fmt.Errorf("invalid purchased date format (expected MM/DD/YYYY): %s", record.Purchased)
		}
	}

	// Parse shares (decimal representing hundreds of shares)
	sharesFloat, err := strconv.ParseFloat(record.Shares, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid shares format: %s", record.Shares)
	}
	shares := int(sharesFloat * 100) // Convert to actual shares count

	// Parse buy price
	buyPrice, err := strconv.ParseFloat(record.BuyPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid buy price format: %s", record.BuyPrice)
	}

	// Parse optional closed date
	var closed *time.Time
	if record.ClosedDate != "" {
		closedDate, err := time.Parse("1/2/2006", record.ClosedDate)
		if err != nil {
			closedDate, err = time.Parse("01/02/2006", record.ClosedDate)
			if err != nil {
				return nil, fmt.Errorf("invalid closed date format (expected MM/DD/YYYY): %s", record.ClosedDate)
			}
		}
		closed = &closedDate
	}

	// Parse optional exit price
	var exitPrice *float64
	if record.ExitPrice != "" {
		price, err := strconv.ParseFloat(record.ExitPrice, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid exit price format: %s", record.ExitPrice)
		}
		exitPrice = &price
	}

	position := &models.LongPosition{
		Symbol:    strings.ToUpper(record.Symbol),
		Opened:    purchased,
		Closed:    closed,
		Shares:    shares,
		BuyPrice:  buyPrice,
		ExitPrice: exitPrice,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return position, nil
}

// processDividendRecord processes a single dividend record from CSV
func (s *Server) processDividendRecord(csvRecord CSVDividendRecord, rowNum int) (*models.Dividend, bool, error) {
	// Validate symbol
	if csvRecord.Symbol == "" {
		return nil, false, fmt.Errorf("symbol cannot be empty")
	}

	// Parse date (accepting multiple formats: MM/DD/YYYY, M/D/YYYY, MM/DD/YY, M/D/YY)
	var receivedDate time.Time
	var err error

	// Try different date formats
	dateFormats := []string{
		"1/2/2006",   // M/D/YYYY
		"01/02/2006", // MM/DD/YYYY
		"1/2/06",     // M/D/YY
		"01/02/06",   // MM/DD/YY
	}

	for _, format := range dateFormats {
		receivedDate, err = time.Parse(format, csvRecord.DateReceived)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, false, fmt.Errorf("invalid date format '%s' (expected MM/DD/YYYY, M/D/YYYY, MM/DD/YY, or M/D/YY)", csvRecord.DateReceived)
	}

	// Parse amount (handle dollar signs)
	amountStr := strings.TrimSpace(csvRecord.Amount)
	amountStr = strings.TrimPrefix(amountStr, "$")
	amountStr = strings.ReplaceAll(amountStr, ",", "")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, false, fmt.Errorf("invalid amount '%s'", csvRecord.Amount)
	}

	if amount <= 0 {
		return nil, false, fmt.Errorf("amount must be positive, got %.2f", amount)
	}

	// Ensure symbol exists
	symbol, err := s.symbolService.GetBySymbol(csvRecord.Symbol)
	if err != nil {
		// Create symbol if it doesn't exist
		log.Printf("[DIVIDENDS_IMPORT] Creating new symbol: %s", csvRecord.Symbol)
		symbol, err = s.symbolService.Create(csvRecord.Symbol)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create symbol '%s': %v", csvRecord.Symbol, err)
		}
	}

	// Check if dividend already exists (to avoid duplicates)
	existingDividends, err := s.dividendService.GetBySymbol(symbol.Symbol)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check existing dividends: %v", err)
	}

	for _, existing := range existingDividends {
		if existing.Received.Equal(receivedDate) && existing.Amount == amount {
			return existing, false, nil // Already exists, skip
		}
	}

	// Create the dividend
	dividend, err := s.dividendService.Create(symbol.Symbol, receivedDate, amount)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create dividend: %v", err)
	}

	return dividend, true, nil
}

// processTreasuryRecord processes a single treasury record from CSV
func (s *Server) processTreasuryRecord(csvRecord CSVTreasuryRecord, rowNum int) (*models.Treasury, bool, error) {
	// Validate CUSPID
	if csvRecord.CUSPID == "" {
		return nil, false, fmt.Errorf("CUSPID cannot be empty")
	}

	// Parse purchased date (accepting multiple formats: YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY)
	var purchasedDate time.Time
	var err error

	// Try different date formats
	dateFormats := []string{
		"2006-01-02", // YYYY-MM-DD
		"1/2/2006",   // M/D/YYYY
		"01/02/2006", // MM/DD/YYYY
		"1/2/06",     // M/D/YY
		"01/02/06",   // MM/DD/YY
	}

	for _, format := range dateFormats {
		purchasedDate, err = time.Parse(format, csvRecord.Purchased)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, false, fmt.Errorf("invalid purchased date format '%s' (expected YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY, MM/DD/YY, or M/D/YY)", csvRecord.Purchased)
	}

	// Parse maturity date
	var maturityDate time.Time
	for _, format := range dateFormats {
		maturityDate, err = time.Parse(format, csvRecord.Maturity)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, false, fmt.Errorf("invalid maturity date format '%s' (expected YYYY-MM-DD, MM/DD/YYYY, M/D/YYYY, MM/DD/YY, or M/D/YY)", csvRecord.Maturity)
	}

	// Parse amount
	amount, err := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(csvRecord.Amount, ",", ""), "$"), 64)
	if err != nil {
		return nil, false, fmt.Errorf("invalid amount '%s'", csvRecord.Amount)
	}

	if amount <= 0 {
		return nil, false, fmt.Errorf("amount must be positive, got %.2f", amount)
	}

	// Parse yield
	yieldStr := strings.TrimSuffix(csvRecord.Yield, "%")
	yield, err := strconv.ParseFloat(yieldStr, 64)
	if err != nil {
		return nil, false, fmt.Errorf("invalid yield '%s'", csvRecord.Yield)
	}

	// Parse buy price
	buyPrice, err := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(csvRecord.BuyPrice, ",", ""), "$"), 64)
	if err != nil {
		return nil, false, fmt.Errorf("invalid buy price '%s'", csvRecord.BuyPrice)
	}

	if buyPrice <= 0 {
		return nil, false, fmt.Errorf("buy price must be positive, got %.2f", buyPrice)
	}

	// Parse optional current value
	var currentValue *float64
	if csvRecord.CurrentValue != "" {
		value, err := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(csvRecord.CurrentValue, ",", ""), "$"), 64)
		if err != nil {
			return nil, false, fmt.Errorf("invalid current value '%s'", csvRecord.CurrentValue)
		}
		currentValue = &value
	}

	// Parse optional exit price
	var exitPrice *float64
	if csvRecord.ExitPrice != "" {
		price, err := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(csvRecord.ExitPrice, ",", ""), "$"), 64)
		if err != nil {
			return nil, false, fmt.Errorf("invalid exit price '%s'", csvRecord.ExitPrice)
		}
		exitPrice = &price
	}

	// Check if treasury already exists (to avoid duplicates)
	existingTreasury, err := s.treasuryService.GetByCUSPID(csvRecord.CUSPID)
	if err == nil && existingTreasury != nil {
		// Treasury already exists, check if it's the same one
		if existingTreasury.Purchased.Equal(purchasedDate) && 
		   existingTreasury.Maturity.Equal(maturityDate) && 
		   existingTreasury.Amount == amount {
			return existingTreasury, false, nil // Already exists, skip
		}
	}

	// Create the treasury
	treasury, err := s.treasuryService.Create(csvRecord.CUSPID, purchasedDate, maturityDate, amount, yield, buyPrice)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create treasury: %v", err)
	}

	// Update with optional fields if provided
	if currentValue != nil || exitPrice != nil {
		_, err = s.treasuryService.Update(treasury.CUSPID, currentValue, exitPrice)
		if err != nil {
			log.Printf("[TREASURIES_IMPORT] Warning: Failed to update treasury with optional fields: %v", err)
		}
	}

	return treasury, true, nil
}

// ensureSymbolExists creates a symbol if it doesn't exist
func (s *Server) ensureSymbolExists(symbol string) error {
	_, err := s.symbolService.GetBySymbol(symbol)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[IMPORT] Creating new symbol: %s", symbol)
			_, createErr := s.symbolService.Create(symbol)
			if createErr != nil {
				return fmt.Errorf("failed to create symbol %s: %w", symbol, createErr)
			}
		} else {
			return fmt.Errorf("error checking symbol %s: %w", symbol, err)
		}
	}
	return nil
}

// getAvailableDbFiles returns a list of .db files in the ./data directory
func (s *Server) getAvailableDbFiles() ([]string, error) {
	dataDir := "./data"

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var dbFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(file.Name()), ".db") {
			dbFiles = append(dbFiles, file.Name())
		}
	}

	// Sort database files alphabetically
	sort.Strings(dbFiles)

	return dbFiles, nil
}

// getBackupFiles returns a list of .db files in the ./backups directory
func (s *Server) getBackupFiles() ([]string, error) {
	backupDir := "./data/backups"

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backupFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(file.Name()), ".db") {
			backupFiles = append(backupFiles, file.Name())
		}
	}

	// Sort backup files by name (which includes timestamp)
	sort.Sort(sort.Reverse(sort.StringSlice(backupFiles)))

	return backupFiles, nil
}

// handleCreateBackup creates a backup of the specified database file
func (s *Server) handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse form data (handles both application/x-www-form-urlencoded and multipart/form-data)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		// Try regular form parsing as fallback
		if err := r.ParseForm(); err != nil {
			log.Printf("[BACKUP] Error parsing form: %v", err)
			http.Error(w, `{"success": false, "error": "Invalid form data"}`, http.StatusBadRequest)
			return
		}
	}

	dbFileName := r.FormValue("filename")
	log.Printf("[BACKUP] Received filename: '%s'", dbFileName)
	log.Printf("[BACKUP] All form values: %v", r.Form)

	if dbFileName == "" {
		log.Printf("[BACKUP] No filename provided")
		http.Error(w, `{"success": false, "error": "No filename provided"}`, http.StatusBadRequest)
		return
	}

	// Validate filename (security check)
	if strings.Contains(dbFileName, "..") || strings.Contains(dbFileName, "/") || strings.Contains(dbFileName, "\\") {
		log.Printf("[BACKUP] Invalid filename: %s", dbFileName)
		http.Error(w, `{"success": false, "error": "Invalid filename"}`, http.StatusBadRequest)
		return
	}

	// Construct full path to database file in data directory
	sourceFilePath := filepath.Join("./data", dbFileName)
	
	// Check if source file exists
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		log.Printf("[BACKUP] Source file does not exist: %s", sourceFilePath)
		http.Error(w, `{"success": false, "error": "Source file not found"}`, http.StatusNotFound)
		return
	}

	log.Printf("[BACKUP] Checkpointing WAL to ensure all data is committed")
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		log.Printf("[BACKUP] Warning: WAL checkpoint failed: %v", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	baseName := strings.TrimSuffix(dbFileName, ".db")
	backupFileName := fmt.Sprintf("%s.%s.db", baseName, timestamp)
	backupPath := filepath.Join("./data/backups", backupFileName)

	// Create backup by copying the file
	if err := s.copyFile(sourceFilePath, backupPath); err != nil {
		log.Printf("[BACKUP] Error creating backup: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create backup"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[BACKUP] Successfully created backup: %s -> %s", sourceFilePath, backupPath)

	// Return success response
	response := map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Backup created: %s", backupFileName),
		"filename": backupFileName,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// copyFile copies a file from src to dst
func (s *Server) copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync to ensure data is written
	return destFile.Sync()
}

// HandleBackupFile handles operations on individual backup files (currently only DELETE)
func (s *Server) HandleBackupFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		s.handleDeleteBackup(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleDeleteBackup deletes a specified backup file
func (s *Server) handleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract filename from URL path (e.g., /backup/filename.db)
	path := strings.TrimPrefix(r.URL.Path, "/backup/")
	if path == "" || path == r.URL.Path {
		log.Printf("[DELETE_BACKUP] No filename in path: %s", r.URL.Path)
		http.Error(w, `{"success": false, "error": "No filename provided"}`, http.StatusBadRequest)
		return
	}

	// URL decode the filename
	filename := path
	log.Printf("[DELETE_BACKUP] Request to delete: '%s'", filename)

	// Validate filename (security check)
	if filename == "" {
		log.Printf("[DELETE_BACKUP] Empty filename")
		http.Error(w, `{"success": false, "error": "No filename provided"}`, http.StatusBadRequest)
		return
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		log.Printf("[DELETE_BACKUP] Invalid filename: %s", filename)
		http.Error(w, `{"success": false, "error": "Invalid filename"}`, http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(filename), ".db") {
		log.Printf("[DELETE_BACKUP] Invalid file extension: %s", filename)
		http.Error(w, `{"success": false, "error": "Invalid file type"}`, http.StatusBadRequest)
		return
	}

	// Construct full backup file path
	backupPath := filepath.Join("./data/backups", filename)

	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		log.Printf("[DELETE_BACKUP] Backup file does not exist: %s", backupPath)
		http.Error(w, `{"success": false, "error": "Backup file not found"}`, http.StatusNotFound)
		return
	}

	// Delete the backup file
	if err := os.Remove(backupPath); err != nil {
		log.Printf("[DELETE_BACKUP] Error deleting backup file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to delete backup file"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[DELETE_BACKUP] Successfully deleted backup: %s", backupPath)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Backup deleted: %s", filename),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleSetCurrentDatabase sets the current active database and reconnects all services
func (s *Server) handleSetCurrentDatabase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"success": false, "error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse form data (handles both application/x-www-form-urlencoded and multipart/form-data)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		// Try regular form parsing as fallback
		if err := r.ParseForm(); err != nil {
			log.Printf("[SET_DATABASE] Error parsing form: %v", err)
			http.Error(w, `{"success": false, "error": "Invalid form data"}`, http.StatusBadRequest)
			return
		}
	}

	dbName := r.FormValue("database")
	if dbName == "" {
		log.Printf("[SET_DATABASE] Database name is required")
		http.Error(w, `{"success": false, "error": "Database name is required"}`, http.StatusBadRequest)
		return
	}

	// Validate that the database exists
	dbPath := filepath.Join("./data", dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("[SET_DATABASE] Database does not exist: %s", dbPath)
		http.Error(w, `{"success": false, "error": "Database does not exist"}`, http.StatusNotFound)
		return
	}

	// Close existing database connection
	log.Printf("[SET_DATABASE] Closing existing database connection")
	if err := s.db.Close(); err != nil {
		log.Printf("[SET_DATABASE] Warning: Error closing existing database: %v", err)
	}

	// Set the current database in the filesystem
	if err := database.SetCurrentDatabase(dbName); err != nil {
		log.Printf("[SET_DATABASE] Error setting current database: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to set current database"}`, http.StatusInternalServerError)
		return
	}

	// Connect to the new database
	log.Printf("[SET_DATABASE] Connecting to new database: %s", dbPath)
	dbWrapper, err := database.NewDB(dbPath)
	if err != nil {
		log.Printf("[SET_DATABASE] Error connecting to new database: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to connect to new database"}`, http.StatusInternalServerError)
		return
	}

	// Update server's database connection and reinitialize all services
	log.Printf("[SET_DATABASE] Reinitializing services with new database connection")
	s.db = dbWrapper.DB
	s.optionService = models.NewOptionService(dbWrapper.DB)
	s.symbolService = models.NewSymbolService(dbWrapper.DB)
	s.treasuryService = models.NewTreasuryService(dbWrapper.DB)
	s.longPositionService = models.NewLongPositionService(dbWrapper.DB)
	s.dividendService = models.NewDividendService(dbWrapper.DB)
	s.settingService = models.NewSettingService(dbWrapper.DB)
	s.metricService = models.NewMetricService(dbWrapper.DB)

	log.Printf("[SET_DATABASE] Successfully switched to database: %s", dbName)

	// Return success response
	response := map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Current database set to: %s", dbName),
		"database": dbName,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleCreateDatabase creates a new database
func (s *Server) handleCreateDatabase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"success": false, "error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse form data (handles both application/x-www-form-urlencoded and multipart/form-data)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		// Try regular form parsing as fallback
		if err := r.ParseForm(); err != nil {
			log.Printf("[CREATE_DATABASE] Error parsing form: %v", err)
			http.Error(w, `{"success": false, "error": "Invalid form data"}`, http.StatusBadRequest)
			return
		}
	}

	dbName := strings.TrimSpace(r.FormValue("name"))
	if dbName == "" {
		log.Printf("[CREATE_DATABASE] Database name is required")
		http.Error(w, `{"success": false, "error": "Database name is required"}`, http.StatusBadRequest)
		return
	}

	// Validate database name (simple validation)
	if strings.ContainsAny(dbName, "/\\:*?\"<>|") {
		log.Printf("[CREATE_DATABASE] Invalid database name: %s", dbName)
		http.Error(w, `{"success": false, "error": "Invalid database name"}`, http.StatusBadRequest)
		return
	}

	// Create the database
	if err := database.CreateNewDatabase(dbName); err != nil {
		log.Printf("[CREATE_DATABASE] Error creating database: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, `{"success": false, "error": "Database already exists"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"success": false, "error": "Failed to create database"}`, http.StatusInternalServerError)
		}
		return
	}

	// Add .db extension if not present for response
	if !strings.HasSuffix(dbName, ".db") {
		dbName = dbName + ".db"
	}

	log.Printf("[CREATE_DATABASE] Successfully created database: %s", dbName)

	// Return success response
	response := map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Database created: %s", dbName),
		"database": dbName,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGenerateTestData handles the POST request to generate wheel strategy test data
func (s *Server) HandleGenerateTestData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[GENERATE_TEST_DATA] Starting test data generation")

	// Test database connection
	var testCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&testCount)
	if err != nil {
		log.Printf("[GENERATE_TEST_DATA] ERROR: Database connection test failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Database connection failed: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	log.Printf("[GENERATE_TEST_DATA] Database connection OK, current symbols count: %d", testCount)

	// Read the SQL file content from internal/database
	sqlContent, err := os.ReadFile("internal/database/wheel_strategy_example_clean.sql")
	if err != nil {
		log.Printf("[GENERATE_TEST_DATA] ERROR: Failed to read SQL file: %v", err)
		http.Error(w, "Failed to read test data file", http.StatusInternalServerError)
		return
	}

	// Parse SQL content into individual statements, handling comments properly
	statements := s.parseSQLStatements(string(sqlContent))

	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("[GENERATE_TEST_DATA] ERROR: Failed to begin transaction: %v", err)
		http.Error(w, "Failed to begin database transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Execute each SQL statement
	successCount := 0
	for i, statement := range statements {
		if statement == "" {
			continue // Skip empty statements
		}

		log.Printf("[GENERATE_TEST_DATA] Executing statement %d: %s", i+1, statement[:min(len(statement), 200)])
		log.Printf("[SQL]  %s\n", statement)
		result, err := tx.Exec(statement)
		if err != nil {
			log.Printf("[GENERATE_TEST_DATA] ERROR: Failed to execute statement %d: %v", i+1, err)
			log.Printf("[GENERATE_TEST_DATA] Statement was: %s", statement)

			// Return JSON error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]interface{}{
				"success":   false,
				"error":     fmt.Sprintf("Failed to execute SQL statement %d: %v", i+1, err),
				"statement": statement,
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Log the result for debugging
		rowsAffected, _ := result.RowsAffected()
		log.Printf("[GENERATE_TEST_DATA] Statement %d executed successfully, rows affected: %d", i+1, rowsAffected)
		successCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("[GENERATE_TEST_DATA] ERROR: Failed to commit transaction: %v", err)
		http.Error(w, "Failed to commit database changes", http.StatusInternalServerError)
		return
	}

	log.Printf("[GENERATE_TEST_DATA] Successfully executed %d SQL statements", successCount)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully generated test data with %d operations", successCount),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleDeleteDatabase deletes a database file from the data directory
func (s *Server) handleDeleteDatabase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "DELETE" {
		http.Error(w, `{"success": false, "error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract database filename from URL path (e.g., /database/delete/test.db)
	path := strings.TrimPrefix(r.URL.Path, "/database/delete/")
	if path == "" || path == r.URL.Path {
		log.Printf("[DELETE_DATABASE] No filename in path: %s", r.URL.Path)
		http.Error(w, `{"success": false, "error": "No database name provided"}`, http.StatusBadRequest)
		return
	}

	// URL decode the filename
	dbName := path
	log.Printf("[DELETE_DATABASE] Request to delete: '%s'", dbName)

	// Validate database name (security check)
	if dbName == "" {
		log.Printf("[DELETE_DATABASE] Empty database name")
		http.Error(w, `{"success": false, "error": "No database name provided"}`, http.StatusBadRequest)
		return
	}

	if strings.Contains(dbName, "..") || strings.Contains(dbName, "/") || strings.Contains(dbName, "\\") {
		log.Printf("[DELETE_DATABASE] Invalid database name: %s", dbName)
		http.Error(w, `{"success": false, "error": "Invalid database name"}`, http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(dbName), ".db") {
		log.Printf("[DELETE_DATABASE] Invalid database extension: %s", dbName)
		http.Error(w, `{"success": false, "error": "Invalid database file type"}`, http.StatusBadRequest)
		return
	}

	// Check if this is the current database (prevent deletion of current database)
	currentDB, err := database.GetCurrentDatabase()
	if err == nil && currentDB == dbName {
		log.Printf("[DELETE_DATABASE] Cannot delete current database: %s", dbName)
		http.Error(w, `{"success": false, "error": "Cannot delete the currently active database"}`, http.StatusConflict)
		return
	}

	// Construct full database file path
	dbPath := filepath.Join("./data", dbName)

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("[DELETE_DATABASE] Database file does not exist: %s", dbPath)
		http.Error(w, `{"success": false, "error": "Database file not found"}`, http.StatusNotFound)
		return
	}

	// Delete the database file
	if err := os.Remove(dbPath); err != nil {
		log.Printf("[DELETE_DATABASE] Error deleting database file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to delete database file"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[DELETE_DATABASE] Successfully deleted database: %s", dbPath)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Database deleted: %s", dbName),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseSQLStatements parses SQL content into individual executable statements,
// properly handling comments and empty lines
func (s *Server) parseSQLStatements(sqlContent string) []string {
	lines := strings.Split(sqlContent, "\n")
	var statements []string
	var currentStatement strings.Builder
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines and comment lines
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}
		
		// Remove inline comments (everything after -- on the same line)
		if commentPos := strings.Index(trimmedLine, "--"); commentPos != -1 {
			trimmedLine = strings.TrimSpace(trimmedLine[:commentPos])
			if trimmedLine == "" {
				continue // Skip if line becomes empty after removing comment
			}
		}
		
		// Add line to current statement
		if currentStatement.Len() > 0 {
			currentStatement.WriteString(" ")
		}
		currentStatement.WriteString(trimmedLine)
		
		// If line ends with semicolon, we have a complete statement
		if strings.HasSuffix(trimmedLine, ";") {
			statement := strings.TrimSpace(currentStatement.String())
			// Remove trailing semicolon and add to statements if not empty
			statement = strings.TrimSuffix(statement, ";")
			statement = strings.TrimSpace(statement)
			if statement != "" {
				statements = append(statements, statement)
			}
			currentStatement.Reset()
		}
	}
	
	// Handle any remaining statement that doesn't end with semicolon
	if currentStatement.Len() > 0 {
		statement := strings.TrimSpace(currentStatement.String())
		if statement != "" && !strings.HasPrefix(statement, "--") {
			statements = append(statements, statement)
		}
	}
	
	return statements
}
