/**
 * Navigation Toggle Functionality
 * Handles collapsible symbols section in the sidebar
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize symbols toggle functionality
    initializeSymbolsToggle();
    
    // Initialize admin toggle functionality
    initializeAdminToggle();
    
    // Check localStorage for symbols section state and restore it
    restoreSymbolsState();
    
    // Check localStorage for admin section state and restore it
    restoreAdminState();
});

function initializeSymbolsToggle() {
    const symbolsToggle = document.getElementById('symbolsToggle');
    const symbolsList = document.getElementById('symbolsList');
    
    if (!symbolsToggle || !symbolsList) {
        return; // Elements not found, skip initialization
    }
    
    symbolsToggle.addEventListener('click', function(e) {
        e.preventDefault();
        toggleSymbolsSection();
    });
}

function toggleSymbolsSection() {
    const symbolsToggle = document.getElementById('symbolsToggle');
    const symbolsList = document.getElementById('symbolsList');
    
    if (!symbolsToggle || !symbolsList) {
        return;
    }
    
    const isExpanded = symbolsList.classList.contains('expanded');
    
    if (isExpanded) {
        // Collapse
        symbolsList.classList.remove('expanded');
        symbolsToggle.classList.remove('expanded');
        localStorage.setItem('symbolsExpanded', 'false');
    } else {
        // Expand
        symbolsList.classList.add('expanded');
        symbolsToggle.classList.add('expanded');
        localStorage.setItem('symbolsExpanded', 'true');
    }
}

function restoreSymbolsState() {
    const symbolsToggle = document.getElementById('symbolsToggle');
    const symbolsList = document.getElementById('symbolsList');
    
    if (!symbolsToggle || !symbolsList) {
        return;
    }
    
    // Default to expanded if no preference is stored
    const isExpanded = localStorage.getItem('symbolsExpanded') !== 'false';
    
    if (isExpanded) {
        symbolsList.classList.add('expanded');
        symbolsToggle.classList.add('expanded');
    } else {
        symbolsList.classList.remove('expanded');
        symbolsToggle.classList.remove('expanded');
    }
}

function initializeAdminToggle() {
    const adminToggle = document.getElementById('adminToggle');
    const adminList = document.getElementById('adminList');
    
    if (!adminToggle || !adminList) {
        return; // Elements not found, skip initialization
    }
    
    adminToggle.addEventListener('click', function(e) {
        e.preventDefault();
        toggleAdminSection();
    });
}

function toggleAdminSection() {
    const adminToggle = document.getElementById('adminToggle');
    const adminList = document.getElementById('adminList');
    
    if (!adminToggle || !adminList) {
        return;
    }
    
    const isExpanded = adminList.classList.contains('expanded');
    
    if (isExpanded) {
        // Collapse
        adminList.classList.remove('expanded');
        adminToggle.classList.remove('expanded');
        localStorage.setItem('adminExpanded', 'false');
    } else {
        // Expand
        adminList.classList.add('expanded');
        adminToggle.classList.add('expanded');
        localStorage.setItem('adminExpanded', 'true');
    }
}

function restoreAdminState() {
    const adminToggle = document.getElementById('adminToggle');
    const adminList = document.getElementById('adminList');
    
    if (!adminToggle || !adminList) {
        return;
    }
    
    // Default to collapsed if no preference is stored
    const isExpanded = localStorage.getItem('adminExpanded') === 'true';
    
    if (isExpanded) {
        adminList.classList.add('expanded');
        adminToggle.classList.add('expanded');
    } else {
        adminList.classList.remove('expanded');
        adminToggle.classList.remove('expanded');
    }
}

// Export functions for testing or external use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        initializeSymbolsToggle,
        toggleSymbolsSection,
        restoreSymbolsState,
        initializeAdminToggle,
        toggleAdminSection,
        restoreAdminState
    };
}