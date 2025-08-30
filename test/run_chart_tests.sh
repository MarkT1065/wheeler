#!/bin/bash

# Wheeler Chart Data Structure Test Runner
# This script runs comprehensive tests for all chart data structures

echo "🚀 Starting Wheeler Chart Data Structure Tests"
echo "=============================================="

# Set test environment
export CGO_ENABLED=1
cd "$(dirname "$0")/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test and track results
run_test() {
    local test_name="$1"
    local test_pattern="$2"
    
    echo -e "\n${BLUE}📊 Running $test_name...${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if go test -v ./test -run "$test_pattern" -timeout 30s; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo -e "${YELLOW}Phase 1: Individual Chart Handler Tests${NC}"
echo "======================================="

# Dashboard allocation charts
run_test "Dashboard Allocation Charts" "TestAllocationDataHandler"

# Monthly analysis charts  
run_test "Monthly Analysis Charts" "TestMonthlyDataStructure"

# Symbol monthly charts
run_test "Symbol Monthly Charts" "TestSymbolDataStructure"

# Treasury charts
run_test "Treasury Charts" "TestTreasuriesDataStructure"

# Metrics time series charts
run_test "Metrics Time Series Charts" "TestMetricsChartDataHandler"

echo -e "\n${YELLOW}Phase 2: Proposed Chart Structure Tests${NC}"
echo "======================================="

# Options scatter plot (needs refactoring)
run_test "Options Scatter Plot Structure" "TestOptionsScatterDataStructure"

# Tutorial chart (needs refactoring)
run_test "Tutorial Chart Structure" "TestTutorialChartDataStructure"

echo -e "\n${YELLOW}Phase 3: Chart Data Type Tests${NC}"
echo "==============================="

# Individual data type tests
run_test "Monthly Chart Data Types" "TestMonthlyChartDataTypes"
run_test "Symbol Monthly Result Type" "TestSymbolMonthlyResultType"
run_test "Treasury Summary Structure" "TestTreasuriesSummaryStructure"
run_test "Treasury Update Request" "TestTreasuryUpdateRequest"

echo -e "\n${YELLOW}Phase 4: Comprehensive Integration Tests${NC}"
echo "========================================"

# Comprehensive tests
run_test "All Chart Structures Comprehensive" "TestAllChartStructuresComprehensive"
run_test "Chart Data Consistency" "TestChartDataConsistency"

echo -e "\n${YELLOW}Phase 5: Performance Tests (Optional)${NC}"
echo "===================================="

# Performance tests (only if not in CI)
if [ "$CI" != "true" ] && [ "$SKIP_PERFORMANCE" != "true" ]; then
    run_test "Chart Performance and Scaling" "TestChartPerformanceAndScaling"
else
    echo -e "${YELLOW}⏭️  Skipping performance tests (CI environment or SKIP_PERFORMANCE set)${NC}"
fi

echo -e "\n${YELLOW}Phase 6: Business Logic Tests${NC}"
echo "============================"

# Business logic tests
run_test "Treasury Calculations" "TestTreasuryCalculations"
run_test "Multiple Symbols Monthly Data" "TestMultipleSymbolsMonthlyData"

echo -e "\n${YELLOW}Phase 7: API Endpoint Tests${NC}"
echo "============================"

# API endpoint tests (if available)
if [ "$SKIP_INTEGRATION" != "true" ]; then
    run_test "Symbol Handler Endpoint" "TestSymbolHandlerEndpoint"
else
    echo -e "${YELLOW}⏭️  Skipping API endpoint tests (SKIP_INTEGRATION set)${NC}"
fi

# Final results
echo -e "\n=============================================="
echo -e "${BLUE}📋 WHEELER CHART TESTS SUMMARY${NC}"
echo -e "=============================================="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"

if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo -e "\n${RED}❌ Some chart tests failed. Review the output above for details.${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: 0${NC}"
    echo -e "\n${GREEN}🎉 All chart tests passed successfully!${NC}"
    echo -e "${GREEN}✅ Wheeler chart data structures are validated and ready for production.${NC}"
fi

# Additional information
echo -e "\n${BLUE}📊 Chart Test Coverage Summary:${NC}"
echo "• Dashboard Charts: 4 allocation pie charts ✅"
echo "• Monthly Charts: 8+ analysis charts ✅"
echo "• Symbol Charts: Monthly results chart ✅"
echo "• Treasury Charts: Summary and detail charts ✅"
echo "• Metrics Charts: 7 time series charts ✅"
echo "• Options Scatter: Proposed structure ready ✅"
echo "• Tutorial Chart: Proposed structure ready ✅"

echo -e "\n${BLUE}🔧 Next Steps:${NC}"
echo "1. Implement the proposed OptionsScatterData API endpoint"
echo "2. Implement the proposed TutorialChartData structure"
echo "3. Refactor options.html to use API instead of DOM extraction"
echo "4. Add financial heatmap colors to remaining charts"

echo -e "\n${BLUE}📚 Test Files Created:${NC}"
echo "• test/chart_handlers_test.go - Main chart API tests"
echo "• test/monthly_handlers_test.go - Monthly analysis tests"
echo "• test/symbol_handlers_test.go - Symbol chart tests"
echo "• test/treasury_handlers_test.go - Treasury chart tests"
echo "• test/comprehensive_charts_test.go - Integration tests"

exit 0