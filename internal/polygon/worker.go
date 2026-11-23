package polygon

import (
	"context"
	"fmt"
	"log"
	"math"
	"stonks/internal/models"
	"strings"
	"sync"
	"time"
)

type APICallLog struct {
	Endpoint     string        `json:"endpoint"`
	Symbol       string        `json:"symbol"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	Error        string        `json:"error,omitempty"`
}

type WorkerStatus struct {
	Running       bool         `json:"running"`
	StartedAt     time.Time    `json:"started_at"`
	Uptime        string       `json:"uptime"`
	TotalAPICalls int64        `json:"total_api_calls"`
	AvgResponseMs float64      `json:"avg_response_ms"`
	LastError     string       `json:"last_error,omitempty"`
	LastErrorAt   *time.Time   `json:"last_error_at,omitempty"`
	CurrentTask   string       `json:"current_task"`
	RecentCalls   []APICallLog `json:"recent_calls"`
}

type RateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	lastCall time.Time
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCall)
	if elapsed < rl.interval {
		remaining := rl.interval - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(remaining):
		}
	}
	rl.lastCall = time.Now()
	return nil
}

type Worker struct {
	client               *Client
	symbolService        *models.SymbolService
	optionService        *models.OptionService
	longPositionService  *models.LongPositionService
	dividendService      *models.DividendService
	priceHistoryService  *models.PriceHistoryService
	optionHistoryService *models.OptionHistoryService
	indexHistoryService  *models.IndexHistoryService
	settingService       *models.SettingService
	rateLimiter          *RateLimiter
	cycleInterval        time.Duration
	waitGroup            sync.WaitGroup

	mu              sync.RWMutex
	running         bool
	startedAt       time.Time
	totalAPICalls   int64
	totalResponseMs float64
	lastError       string
	lastErrorAt     *time.Time
	currentTask     string
	recentCalls     []APICallLog
}

func NewWorker(
	symbolService *models.SymbolService,
	optionService *models.OptionService,
	longPositionService *models.LongPositionService,
	dividendService *models.DividendService,
	priceHistoryService *models.PriceHistoryService,
	optionHistoryService *models.OptionHistoryService,
	indexHistoryService *models.IndexHistoryService,
	settingService *models.SettingService,
) *Worker {
	return &Worker{
		symbolService:        symbolService,
		optionService:        optionService,
		longPositionService:  longPositionService,
		dividendService:      dividendService,
		priceHistoryService:  priceHistoryService,
		optionHistoryService: optionHistoryService,
		indexHistoryService:  indexHistoryService,
		settingService:       settingService,
		rateLimiter:          NewRateLimiter(12 * time.Second),
		cycleInterval:        time.Hour,
		recentCalls:          make([]APICallLog, 0, 20),
	}
}

func (w *Worker) GetStatus() WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	status := WorkerStatus{
		Running:       w.running,
		StartedAt:     w.startedAt,
		TotalAPICalls: w.totalAPICalls,
		LastError:     w.lastError,
		LastErrorAt:   w.lastErrorAt,
		CurrentTask:   w.currentTask,
		RecentCalls:   make([]APICallLog, len(w.recentCalls)),
	}

	copy(status.RecentCalls, w.recentCalls)

	if w.running {
		status.Uptime = time.Since(w.startedAt).Truncate(time.Second).String()
	}

	if w.totalAPICalls > 0 {
		status.AvgResponseMs = w.totalResponseMs / float64(w.totalAPICalls)
	}

	return status
}

func (w *Worker) setCurrentTask(task string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.currentTask = task
}

func (w *Worker) logAPICall(endpoint, symbol string, statusCode int, responseTime time.Duration, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := APICallLog{
		Endpoint:     endpoint,
		Symbol:       symbol,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
	}
	if err != nil {
		entry.Error = err.Error()
		errMsg := err.Error()
		now := time.Now()
		w.lastError = errMsg
		w.lastErrorAt = &now
	}

	w.totalAPICalls++
	w.totalResponseMs += float64(responseTime.Milliseconds())

	w.recentCalls = append(w.recentCalls, entry)
	if len(w.recentCalls) > 20 {
		w.recentCalls = w.recentCalls[len(w.recentCalls)-20:]
	}
}

func (w *Worker) getClient() (*Client, error) {
	apiKey := w.settingService.GetValue("POLYGON_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}
	return NewClient(apiKey), nil
}

func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.startedAt = time.Now()
	w.mu.Unlock()

	log.Printf("[WORKER] Background worker started")

	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		defer func() {
			w.mu.Lock()
			w.running = false
			w.currentTask = "stopped"
			w.mu.Unlock()
			log.Printf("[WORKER] Background worker stopped")
		}()

		for ctx.Err() == nil {
			w.runCycle(ctx)
			if !waitForWorkerCycle(ctx, w.cycleInterval) {
				return
			}
		}
	}()
}

func (w *Worker) Wait() {
	w.waitGroup.Wait()
}

func waitForWorkerCycle(ctx context.Context, interval time.Duration) bool {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (w *Worker) runCycle(ctx context.Context) {
	client, err := w.getClient()
	if err != nil {
		w.setCurrentTask("waiting for API key")
		log.Printf("[WORKER] No API key configured, skipping cycle")
		return
	}
	w.client = client

	activeSymbols, err := w.symbolService.GetActivePositionSymbols()
	if err != nil {
		log.Printf("[WORKER] Failed to get active symbols: %v", err)
	} else {
		w.processActiveSymbols(ctx, activeSymbols)
		if ctx.Err() != nil {
			return
		}
	}

	inactiveSymbols, err := w.symbolService.GetInactivePositionSymbols()
	if err != nil {
		log.Printf("[WORKER] Failed to get inactive symbols: %v", err)
	} else {
		w.processInactiveSymbols(ctx, inactiveSymbols)
		if ctx.Err() != nil {
			return
		}
	}

	w.indexHistoryFetch(ctx)
}

func (w *Worker) processActiveSymbols(ctx context.Context, symbols []string) {
	log.Printf("[WORKER] Processing %d active symbols (open positions)", len(symbols))

	for _, symbol := range symbols {
		if ctx.Err() != nil {
			return
		}

		w.setCurrentTask(fmt.Sprintf("active: %s price", symbol))
		w.updateSymbolPrice(ctx, symbol)
		if ctx.Err() != nil {
			return
		}

		w.setCurrentTask(fmt.Sprintf("active: %s history", symbol))
		w.backfillSymbolHistory(ctx, symbol)
		if ctx.Err() != nil {
			return
		}

		w.setCurrentTask(fmt.Sprintf("active: %s dividends", symbol))
		w.fetchSymbolDividends(ctx, symbol)
		if ctx.Err() != nil {
			return
		}
	}

	w.setCurrentTask("active: option history")
	w.optionHistoryFetch(ctx)
}

func (w *Worker) processInactiveSymbols(ctx context.Context, symbols []string) {
	log.Printf("[WORKER] Processing %d inactive symbols (closed positions only)", len(symbols))

	for _, symbol := range symbols {
		if ctx.Err() != nil {
			return
		}

		w.setCurrentTask(fmt.Sprintf("inactive: %s price", symbol))
		w.updateSymbolPrice(ctx, symbol)
		if ctx.Err() != nil {
			return
		}

		w.setCurrentTask(fmt.Sprintf("inactive: %s history", symbol))
		w.backfillSymbolHistory(ctx, symbol)
		if ctx.Err() != nil {
			return
		}
	}
}

func (w *Worker) updateSymbolPrice(ctx context.Context, symbol string) {
	if err := w.rateLimiter.Wait(ctx); err != nil {
		return
	}

	start := time.Now()
	quote, err := w.client.GetPreviousClose(ctx, symbol)
	elapsed := time.Since(start)

	if err != nil {
		w.logAPICall("GET /v2/aggs/ticker/"+symbol+"/prev", symbol, 0, elapsed, err)
		log.Printf("[WORKER] Failed to update price for %s: %v", symbol, err)
		return
	}
	w.logAPICall("GET /v2/aggs/ticker/"+symbol+"/prev", symbol, 200, elapsed, nil)

	currentSymbol, err := w.symbolService.GetBySymbol(symbol)
	if err != nil {
		log.Printf("[WORKER] Failed to get symbol data for %s: %v", symbol, err)
		return
	}

	_, err = w.symbolService.Update(
		symbol,
		quote.Results.Price,
		currentSymbol.Dividend,
		currentSymbol.ExDividendDate,
		currentSymbol.PERatio,
	)
	if err != nil {
		log.Printf("[WORKER] Failed to save price for %s: %v", symbol, err)
	}
}

func (w *Worker) backfillSymbolHistory(ctx context.Context, symbol string) {
	latestDate, err := w.priceHistoryService.GetLatestDate(symbol)
	if err != nil {
		log.Printf("[WORKER] Failed to get latest date for %s: %v", symbol, err)
		return
	}

	today := time.Now().Format("2006-01-02")
	from := latestDate.AddDate(0, 0, 1).Format("2006-01-02")
	if from >= today {
		return
	}

	if err := w.rateLimiter.Wait(ctx); err != nil {
		return
	}

	start := time.Now()
	aggs, err := w.client.GetAggregates(ctx, symbol, "day", from, today, 5000)
	elapsed := time.Since(start)

	if err != nil {
		w.logAPICall("GET /v2/aggs/ticker/"+symbol+"/range", symbol, 0, elapsed, err)
		log.Printf("[WORKER] Failed to fetch history for %s: %v", symbol, err)
		return
	}
	w.logAPICall("GET /v2/aggs/ticker/"+symbol+"/range", symbol, 200, elapsed, nil)

	for _, bar := range aggs.Results {
		barDate := time.UnixMilli(bar.Timestamp)
		priceBar := &models.PriceBar{
			Symbol:       symbol,
			BarDate:      barDate,
			Open:         bar.Open,
			High:         bar.High,
			Low:          bar.Low,
			Close:        bar.Close,
			Volume:       bar.Volume,
			VWAP:         bar.VWAP,
			Transactions: bar.Transactions,
		}
		if err := w.priceHistoryService.Upsert(priceBar); err != nil {
			log.Printf("[WORKER] Failed to upsert price bar for %s on %s: %v", symbol, barDate.Format("2006-01-02"), err)
		}
	}

	log.Printf("[WORKER] Backfilled %d price bars for %s", len(aggs.Results), symbol)
}

func (w *Worker) fetchSymbolDividends(ctx context.Context, symbol string) {
	if err := w.rateLimiter.Wait(ctx); err != nil {
		return
	}

	start := time.Now()
	divData, err := w.client.GetDividends(ctx, symbol, 4)
	elapsed := time.Since(start)

	if err != nil {
		w.logAPICall("GET /v3/reference/dividends?ticker="+symbol, symbol, 0, elapsed, err)
		log.Printf("[WORKER] Failed to fetch dividends for %s: %v", symbol, err)
		return
	}
	w.logAPICall("GET /v3/reference/dividends?ticker="+symbol, symbol, 200, elapsed, nil)

	if len(divData.Results) == 0 {
		return
	}

	latestDiv := divData.Results[0]
	if latestDiv.ExDividendDate == "" {
		return
	}

	exDate, err := time.Parse("2006-01-02", latestDiv.ExDividendDate)
	if err != nil {
		log.Printf("[WORKER] Failed to parse ex-dividend date for %s: %v", symbol, err)
		return
	}

	currentSymbol, err := w.symbolService.GetBySymbol(symbol)
	if err != nil {
		log.Printf("[WORKER] Failed to get symbol data for %s: %v", symbol, err)
		return
	}

	_, err = w.symbolService.Update(
		symbol,
		currentSymbol.Price,
		latestDiv.CashAmount,
		&exDate,
		currentSymbol.PERatio,
	)
	if err != nil {
		log.Printf("[WORKER] Failed to update dividend info for %s: %v", symbol, err)
	}
}

func (w *Worker) optionHistoryFetch(ctx context.Context) {
	log.Printf("[WORKER] Starting option history fetch")

	openOptions, err := w.optionService.GetOpen()
	if err != nil {
		log.Printf("[WORKER] Failed to get open options: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")

	for _, opt := range openOptions {
		if ctx.Err() != nil {
			return
		}

		optionTicker := buildPolygonOptionTicker(opt)
		if optionTicker == "" {
			continue
		}

		w.setCurrentTask(fmt.Sprintf("active: %s option %s", opt.Symbol, optionTicker))

		latestDate, err := w.optionHistoryService.GetLatestDate(optionTicker)
		if err != nil {
			log.Printf("[WORKER] Failed to get latest date for %s: %v", optionTicker, err)
			continue
		}

		from := latestDate.AddDate(0, 0, 1).Format("2006-01-02")
		if from >= today {
			continue
		}

		if err := w.rateLimiter.Wait(ctx); err != nil {
			return
		}

		start := time.Now()
		aggs, err := w.client.GetOptionAggregates(ctx, optionTicker, "day", from, today, 5000)
		elapsed := time.Since(start)

		if err != nil {
			w.logAPICall("GET /v2/aggs/ticker/"+optionTicker+"/range", opt.Symbol, 0, elapsed, err)
			log.Printf("[WORKER] Failed to fetch option history for %s: %v", optionTicker, err)
			continue
		}
		w.logAPICall("GET /v2/aggs/ticker/"+optionTicker+"/range", opt.Symbol, 200, elapsed, nil)

		contractType := strings.ToLower(opt.Type)
		for _, bar := range aggs.Results {
			barDate := time.UnixMilli(bar.Timestamp)
			optBar := &models.OptionBar{
				OptionTicker:     optionTicker,
				UnderlyingSymbol: opt.Symbol,
				ContractType:     contractType,
				StrikePrice:      opt.Strike,
				ExpirationDate:   opt.Expiration,
				BarDate:          barDate,
				Open:             bar.Open,
				High:             bar.High,
				Low:              bar.Low,
				Close:            bar.Close,
				Volume:           bar.Volume,
				VWAP:             bar.VWAP,
				Transactions:     bar.Transactions,
			}
			if err := w.optionHistoryService.Upsert(optBar); err != nil {
				log.Printf("[WORKER] Failed to upsert option bar for %s on %s: %v", optionTicker, barDate.Format("2006-01-02"), err)
			}
		}

		log.Printf("[WORKER] Fetched %d option bars for %s", len(aggs.Results), optionTicker)
	}

	log.Printf("[WORKER] Option history fetch complete")
}

var trackedIndices = []string{"I:SPX", "I:NDX", "I:DJI", "I:VIX"}

func (w *Worker) indexHistoryFetch(ctx context.Context) {
	w.setCurrentTask("fetching index history")
	log.Printf("[WORKER] Starting index history fetch")

	today := time.Now().Format("2006-01-02")

	for _, indexTicker := range trackedIndices {
		if ctx.Err() != nil {
			return
		}

		latestDate, err := w.indexHistoryService.GetLatestDate(indexTicker)
		if err != nil {
			log.Printf("[WORKER] Failed to get latest date for %s: %v", indexTicker, err)
			continue
		}

		from := latestDate.AddDate(0, 0, 1).Format("2006-01-02")
		if from >= today {
			continue
		}

		if err := w.rateLimiter.Wait(ctx); err != nil {
			return
		}

		start := time.Now()
		aggs, err := w.client.GetIndexAggregates(ctx, indexTicker, "day", from, today, 5000)
		elapsed := time.Since(start)

		if err != nil {
			w.logAPICall("GET /v2/aggs/ticker/"+indexTicker+"/range", indexTicker, 0, elapsed, err)
			log.Printf("[WORKER] Failed to fetch index history for %s: %v", indexTicker, err)
			continue
		}
		w.logAPICall("GET /v2/aggs/ticker/"+indexTicker+"/range", indexTicker, 200, elapsed, nil)

		for _, bar := range aggs.Results {
			barDate := time.UnixMilli(bar.Timestamp)
			indexBar := &models.IndexBar{
				IndexTicker: indexTicker,
				BarDate:     barDate,
				Open:        bar.Open,
				High:        bar.High,
				Low:         bar.Low,
				Close:       bar.Close,
			}
			if err := w.indexHistoryService.Upsert(indexBar); err != nil {
				log.Printf("[WORKER] Failed to upsert index bar for %s on %s: %v", indexTicker, barDate.Format("2006-01-02"), err)
			}
		}

		log.Printf("[WORKER] Fetched %d index bars for %s", len(aggs.Results), indexTicker)
	}

	log.Printf("[WORKER] Index history fetch complete")
}

func buildPolygonOptionTicker(opt *models.Option) string {
	symbol := strings.ToUpper(opt.Symbol)
	dateStr := opt.Expiration.Format("060102")

	var typeChar string
	switch strings.ToUpper(opt.Type) {
	case "CALL":
		typeChar = "C"
	case "PUT":
		typeChar = "P"
	default:
		return ""
	}

	strikePadded := fmt.Sprintf("%08d", int(math.Round(opt.Strike*1000)))

	return fmt.Sprintf("O:%s%s%s%s", symbol, dateStr, typeChar, strikePadded)
}
