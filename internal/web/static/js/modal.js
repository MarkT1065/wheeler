/**
 * Modal Utility Module
 * Provides reusable modal functions for confirm dialogs, error messages, etc.
 */

export function showConfirmModal(title, message, onConfirm) {
    const confirmModal = document.getElementById('confirmModal');
    if (!confirmModal) {
        console.error('confirmModal element not found');
        return;
    }
    
    const titleElement = document.getElementById('confirmModalTitle');
    const messageElement = document.getElementById('confirmModalMessage');
    const confirmButton = document.getElementById('confirmModalConfirm');
    
    if (titleElement) titleElement.textContent = title;
    if (messageElement) messageElement.innerHTML = message;
    
    confirmModal.style.display = 'block';
    
    if (confirmButton) {
        confirmButton.onclick = function() {
            closeConfirmModal();
            if (onConfirm) onConfirm();
        };
    }
}

export function closeConfirmModal() {
    const confirmModal = document.getElementById('confirmModal');
    if (confirmModal) {
        confirmModal.style.display = 'none';
    }
}

export function showErrorModal(message) {
    const errorModal = document.getElementById('errorModal');
    if (!errorModal) {
        alert(message);
        return;
    }
    
    const errorMessage = document.getElementById('errorMessage');
    if (errorMessage) {
        errorMessage.textContent = message;
    }
    
    errorModal.style.display = 'block';
}

export function closeErrorModal() {
    const errorModal = document.getElementById('errorModal');
    if (errorModal) {
        errorModal.style.display = 'none';
    }
}

export function showModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'block';
    }
}

export function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'none';
    }
}

// Close modals when clicking outside
window.addEventListener('click', function(event) {
    if (event.target.classList.contains('modal')) {
        event.target.style.display = 'none';
    }
});

// Close modals with Escape key
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        const modals = document.querySelectorAll('.modal');
        modals.forEach(modal => {
            if (modal.style.display === 'block') {
                modal.style.display = 'none';
            }
        });
    }
});
