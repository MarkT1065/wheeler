// Row actions dropdown functionality
document.addEventListener('DOMContentLoaded', function() {
    // Close all dropdowns when clicking outside
    document.addEventListener('click', function(event) {
        if (!event.target.closest('.row-actions')) {
            document.querySelectorAll('.actions-menu').forEach(menu => {
                menu.classList.remove('show');
            });
        }
    });
    
    // Toggle dropdown on button click
    document.addEventListener('click', function(event) {
        if (event.target.closest('.actions-toggle')) {
            event.stopPropagation();
            const toggle = event.target.closest('.actions-toggle');
            const menu = toggle.nextElementSibling;
            
            // Close all other menus
            document.querySelectorAll('.actions-menu').forEach(m => {
                if (m !== menu) {
                    m.classList.remove('show');
                }
            });
            
            // Toggle this menu
            menu.classList.toggle('show');
        }
    });
});
