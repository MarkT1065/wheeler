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
        
        if (!this.modal) {
            console.warn('Symbol modal not found on this page');
            return;
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
        this.modalTitle.textContent = editMode ? 'Edit Symbol' : 'New Symbol';
        
        if (editMode && symbolData) {
            this.editingSymbol = symbolData.symbol;
            document.getElementById('symbolInput').value = symbolData.symbol;
            document.getElementById('priceInput').value = symbolData.price || '';
            document.getElementById('dividendInput').value = symbolData.dividend || '';
            document.getElementById('exDividendDateInput').value = symbolData.ex_dividend_date || '';
            document.getElementById('peRatioInput').value = symbolData.pe_ratio || '';
            document.getElementById('symbolInput').disabled = true;
        } else {
            this.symbolForm.reset();
            document.getElementById('symbolInput').disabled = false;
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
        
        const symbolData = {
            symbol: document.getElementById('symbolInput').value.toUpperCase(),
            price: parseFloat(document.getElementById('priceInput').value) || 0,
            dividend: parseFloat(document.getElementById('dividendInput').value) || 0,
            ex_dividend_date: document.getElementById('exDividendDateInput').value || null,
            pe_ratio: parseFloat(document.getElementById('peRatioInput').value) || null
        };
        
        const url = `/api/symbols/${symbolData.symbol}`;
        const method = 'PUT'; // Using PUT for both create and update
        
        fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                price: symbolData.price,
                dividend: symbolData.dividend,
                ex_dividend_date: symbolData.ex_dividend_date,
                pe_ratio: symbolData.pe_ratio
            })
        })
        .then(response => {
            if (response.ok) {
                return response.json();
            }
            throw new Error('Failed to save symbol');
        })
        .then(data => {
            console.log('Symbol saved successfully:', data);
            this.close();
            // Refresh the page to show updated symbols
            window.location.reload();
        })
        .catch(error => {
            console.error('Error saving symbol:', error);
            // Show error modal if available, otherwise alert
            if (window.showErrorModal) {
                window.showErrorModal('Failed to save symbol. Please try again.');
            } else {
                alert('Failed to save symbol. Please try again.');
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