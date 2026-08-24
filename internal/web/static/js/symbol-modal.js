/**
 * Shared Symbol Modal Module
 * Consolidates Symbol modal behavior across all Wheeler application pages
 */

class SymbolModal {
    constructor() {
        this.modal = null;
        this.newSymbolBtn = null;
        this.closeModal = null;
        this.cancelModal = null;
        this.symbolForm = null;
        this.modalTitle = null;
        this.marketInput = null;
        this.isEditMode = false;
        this.editingSymbol = null;

        this.init();
    }
    
    init() {
        // Get modal elements
        this.modal = document.getElementById('symbolModal');
        this.newSymbolBtn = document.getElementById('newSymbolBtn');
        this.closeModal = document.getElementById('closeModal');
        this.cancelModal = document.getElementById('cancelModal');
        this.symbolForm = document.getElementById('symbolForm');
        this.modalTitle = document.getElementById('modalTitle');
        this.marketInput = document.getElementById('marketInput');
        
        if (!this.modal) {
            console.warn('Symbol modal not found on this page');
            return;
        }

        if (this.marketInput) {
            fetch('/api/markets')
                .then(response => response.json())
                .then(markets => {
                    this.marketInput.innerHTML = '<option value="">Select a market</option>';
                    markets.forEach(market => {
                        const option = document.createElement('option');
                        option.value = market.id;
                        option.textContent = market.display_name;
                        this.marketInput.appendChild(option);
                    });
                })
                .catch(error => {
                    console.error('Failed to load markets:', error);
                    this.marketInput.innerHTML = '<option value="">Markets unavailable</option>';
                });
        }
        
        this.bindEvents();
    }
    
    bindEvents() {
        // Open modal for new symbol
        if (this.newSymbolBtn) {
            this.newSymbolBtn.addEventListener('click', (e) => {
                e.preventDefault();
                this.open(false);
            });
        }
        
        // Close modal events
        if (this.closeModal) {
            this.closeModal.addEventListener('click', () => this.close());
        }
        if (this.cancelModal) {
            this.cancelModal.addEventListener('click', () => this.close());
        }
        
        // Close modal when clicking outside
        window.addEventListener('click', (event) => {
            if (event.target === this.modal) {
                this.close();
            }
        });
        
        // Handle form submission
        if (this.symbolForm) {
            this.symbolForm.addEventListener('submit', (e) => this.handleSubmit(e));
        }
    }
    
    open(editMode = false, symbolData = null) {
        if (!this.modal) return;
        
        this.isEditMode = editMode;
        if (this.modalTitle) {
            this.modalTitle.textContent = editMode ? 'Edit Symbol' : 'New Symbol';
        }
        
        if (editMode && symbolData) {
            this.editingSymbol = symbolData.symbol;
            const symbolInput = document.getElementById('symbolInput');
            const priceInput = document.getElementById('priceInput');
            const dividendInput = document.getElementById('dividendInput');
            const exDividendDateInput = document.getElementById('exDividendDateInput');
            const peRatioInput = document.getElementById('peRatioInput');
            
            if (symbolInput) {
                const parts = symbolData.symbol.split(':');
                symbolInput.value = parts.length === 2 ? parts[1] : symbolData.symbol;
                symbolInput.disabled = true;
            }
            if (this.marketInput) {
                const parts = symbolData.symbol.split(':');
                this.marketInput.value = parts.length === 2 ? parts[0] : '';
                this.marketInput.disabled = parts.length === 2;
            }
            if (priceInput) priceInput.value = symbolData.price || '';
            if (dividendInput) dividendInput.value = symbolData.dividend || '';
            if (exDividendDateInput) exDividendDateInput.value = symbolData.ex_dividend_date || '';
            if (peRatioInput) peRatioInput.value = symbolData.pe_ratio || '';
        } else {
            if (this.symbolForm) {
                this.symbolForm.reset();
            }
            const symbolInput = document.getElementById('symbolInput');
            if (symbolInput) {
                symbolInput.disabled = false;
            }
			if (this.marketInput) this.marketInput.disabled = false;
            this.editingSymbol = null;
        }
        
        this.modal.style.display = 'block';
    }
    
    close() {
        if (!this.modal) return;
        
        this.modal.style.display = 'none';
        if (this.symbolForm) {
            this.symbolForm.reset();
        }
        this.isEditMode = false;
        this.editingSymbol = null;
    }
    
    handleSubmit(e) {
        e.preventDefault();
        
        const symbolInput = document.getElementById('symbolInput');
        const priceInput = document.getElementById('priceInput');
        const dividendInput = document.getElementById('dividendInput');
        const exDividendDateInput = document.getElementById('exDividendDateInput');
        const peRatioInput = document.getElementById('peRatioInput');
		const market = this.marketInput?.value;
        
        if (!symbolInput) {
            console.error('Symbol input not found');
            return;
        }
        
        if (!market) {
            alert('Select a market');
            return;
        }

        const symbolData = {
            symbol: `${market}:${symbolInput.value.toUpperCase().trim()}`,
            price: parseFloat(priceInput?.value) || 0,
            dividend: parseFloat(dividendInput?.value) || 0,
            ex_dividend_date: exDividendDateInput?.value || null,
            pe_ratio: parseFloat(peRatioInput?.value) || null
        };
        
        const isLegacyEdit = this.isEditMode && this.editingSymbol && !this.editingSymbol.includes(':');
        const url = isLegacyEdit
            ? `/api/symbols/${encodeURIComponent(this.editingSymbol)}/qualify`
            : `/api/symbols/${encodeURIComponent(symbolData.symbol)}`;
        const method = isLegacyEdit ? 'POST' : 'PUT';
        const body = isLegacyEdit
            ? { market: market }
            : {
                price: symbolData.price,
                dividend: symbolData.dividend,
                ex_dividend_date: symbolData.ex_dividend_date,
                pe_ratio: symbolData.pe_ratio
            };

        fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        })
        .then(response => {
            if (response.ok) {
                return response.json();
            }
            return response.text().then(text => {
                throw new Error(text.trim() || 'Failed to save symbol');
            });
        })
        .then(data => {
            console.log('Symbol saved successfully:', data);
            this.close();
            // Refresh the page to show updated symbols
            window.location.reload();
        })
        .catch(error => {
            console.error('Error saving symbol:', error);
            const message = error.message || 'Failed to save symbol. Please try again.';
            // Show error modal if available, otherwise alert
            if (window.showErrorModal) {
                window.showErrorModal(message);
            } else {
                alert(message);
            }
        });
    }
}

// Initialize the Symbol Modal when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Only initialize if there's no custom implementation already present
    if (!window.openSymbolModal && !window.symbolModalInitialized) {
        window.symbolModal = new SymbolModal();
        
        // Make the open method globally available for backward compatibility
        window.openSymbolModal = function(editMode = false, symbolData = null) {
            if (window.symbolModal) {
                window.symbolModal.open(editMode, symbolData);
            }
        };
    }
});

// Export for module systems (if needed in the future)
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SymbolModal;
}
