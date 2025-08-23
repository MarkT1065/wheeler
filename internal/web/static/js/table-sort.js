// Table Sorting Functionality for Wheeler Application
class TableSorter {
    constructor() {
        this.sortDirection = {};
        this.initializeSortableTables();
    }

    initializeSortableTables() {
        document.addEventListener('DOMContentLoaded', () => {
            this.makeTablesSortable();
        });
    }

    makeTablesSortable() {
        const tables = document.querySelectorAll('table');
        tables.forEach(table => {
            const headers = table.querySelectorAll('th');
            if (headers.length > 0) {
                table.classList.add('sortable-table');
                headers.forEach((header, index) => {
                    if (header.textContent.trim() !== '') {
                        header.classList.add('sortable-header');
                        header.setAttribute('data-column', index);
                        header.style.cursor = 'pointer';
                        header.addEventListener('click', (e) => this.sortTable(e, table));
                        
                        // Add sort indicator
                        const sortIndicator = document.createElement('span');
                        sortIndicator.className = 'sort-indicator';
                        sortIndicator.innerHTML = ' ⇅';
                        header.appendChild(sortIndicator);
                    }
                });
            }
        });
    }

    sortTable(event, table) {
        const header = event.target.closest('th');
        const columnIndex = parseInt(header.getAttribute('data-column'));
        const tableId = table.id || `table-${Math.random().toString(36).substr(2, 9)}`;
        
        if (!table.id) {
            table.id = tableId;
        }

        const currentDirection = this.sortDirection[`${tableId}-${columnIndex}`] || 'asc';
        const newDirection = currentDirection === 'asc' ? 'desc' : 'asc';
        this.sortDirection[`${tableId}-${columnIndex}`] = newDirection;

        // Clear all sort indicators in this table
        table.querySelectorAll('.sort-indicator').forEach(indicator => {
            indicator.innerHTML = ' ⇅';
            indicator.parentElement.classList.remove('sort-asc', 'sort-desc');
        });

        // Set the active sort indicator
        const indicator = header.querySelector('.sort-indicator');
        indicator.innerHTML = newDirection === 'asc' ? ' ↑' : ' ↓';
        header.classList.add(`sort-${newDirection}`);

        this.performSort(table, columnIndex, newDirection);
    }

    performSort(table, columnIndex, direction) {
        const tbody = table.querySelector('tbody');
        if (!tbody) return;

        const rows = Array.from(tbody.querySelectorAll('tr'));
        const headerRow = table.querySelector('thead tr');
        
        rows.sort((a, b) => {
            const aCell = a.cells[columnIndex];
            const bCell = b.cells[columnIndex];
            
            if (!aCell || !bCell) return 0;
            
            let aValue = this.getCellValue(aCell);
            let bValue = this.getCellValue(bCell);
            
            // Determine data type and sort accordingly
            const comparison = this.compareValues(aValue, bValue);
            
            return direction === 'asc' ? comparison : -comparison;
        });

        // Remove existing rows
        rows.forEach(row => row.remove());
        
        // Add sorted rows back
        rows.forEach(row => tbody.appendChild(row));
    }

    getCellValue(cell) {
        // Get text content and clean it
        let value = cell.textContent.trim();
        
        // Handle special cases for financial data
        if (value === '' || value === '-' || value === 'N/A') {
            return '';
        }
        
        // Remove currency symbols and commas for numerical sorting
        const cleanValue = value.replace(/[$,%]/g, '');
        
        return cleanValue;
    }

    compareValues(a, b) {
        // Handle empty values
        if (a === '' && b === '') return 0;
        if (a === '') return 1;
        if (b === '') return -1;
        
        // Try to parse as numbers (including percentages and currency)
        const numA = this.parseNumber(a);
        const numB = this.parseNumber(b);
        
        if (!isNaN(numA) && !isNaN(numB)) {
            return numA - numB;
        }
        
        // Try to parse as dates
        const dateA = this.parseDate(a);
        const dateB = this.parseDate(b);
        
        if (dateA && dateB) {
            return dateA.getTime() - dateB.getTime();
        }
        
        // Default to string comparison (case insensitive)
        return a.toLowerCase().localeCompare(b.toLowerCase());
    }

    parseNumber(value) {
        // Remove common non-numeric characters
        const cleaned = value.replace(/[%$,\s]/g, '');
        
        // Handle negative numbers in parentheses
        if (cleaned.startsWith('(') && cleaned.endsWith(')')) {
            return -parseFloat(cleaned.slice(1, -1));
        }
        
        return parseFloat(cleaned);
    }

    parseDate(value) {
        // Try various date formats
        const formats = [
            // ISO format
            /^\d{4}-\d{2}-\d{2}$/,
            // US format
            /^\d{1,2}\/\d{1,2}\/\d{4}$/,
            // Other common formats
            /^\d{1,2}-\d{1,2}-\d{4}$/
        ];
        
        for (let format of formats) {
            if (format.test(value)) {
                const date = new Date(value);
                if (!isNaN(date.getTime())) {
                    return date;
                }
            }
        }
        
        return null;
    }
}

// Initialize table sorting when page loads
const tableSorter = new TableSorter();