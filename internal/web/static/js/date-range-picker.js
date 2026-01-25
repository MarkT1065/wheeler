/**
 * Reusable Month Range Picker
 * Creates a custom month-based date range picker with shortcuts
 * 
 * Usage:
 *   initMonthRangePicker('elementId', {
 *     initialFrom: '2025-01',
 *     initialTo: '2025-12',
 *     onApply: function(fromMonth, toMonth) {
 *       // Handle date range change
 *     }
 *   });
 */

function initMonthRangePicker(elementId, options) {
    const defaults = {
        initialFrom: null,
        initialTo: null,
        onApply: function(fromMonth, toMonth) {
            // Default: reload page with query params
            const url = new URL(window.location.href);
            url.searchParams.set('from', fromMonth);
            url.searchParams.set('to', toMonth);
            window.location.href = url.toString();
        }
    };
    
    const settings = Object.assign({}, defaults, options);
    
    const $input = $('#' + elementId);
    if ($input.length === 0) {
        console.error('Month range picker: Element not found: ' + elementId);
        return;
    }
    
    let startMonth = null;
    let endMonth = null;
    let pickerOpen = false;
    
    // Set initial display value
    if (settings.initialFrom && settings.initialTo) {
        const fromMoment = moment(settings.initialFrom + '-01', 'YYYY-MM-DD');
        const toMoment = moment(settings.initialTo + '-01', 'YYYY-MM-DD');
        $input.val(fromMoment.format('MMM YYYY') + ' - ' + toMoment.format('MMM YYYY'));
        startMonth = fromMoment;
        endMonth = toMoment;
    } else {
        $input.val('Last 12 months');
        startMonth = moment().subtract(11, 'months').startOf('month');
        endMonth = moment().startOf('month');
    }
    
    // Click handler to open picker
    $input.on('click', function() {
        if (pickerOpen) return;
        pickerOpen = true;
        
        const input = $(this);
        const offset = input.offset();
        
        // Create overlay
        const $overlay = $('<div class="month-range-overlay"></div>').css({
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            zIndex: 9998
        });
        
        // Create shortcuts panel
        const $shortcuts = $('<div class="shortcuts"></div>').css({
            display: 'flex',
            flexDirection: 'column',
            gap: '5px',
            borderRight: '1px solid #575757',
            paddingRight: '15px'
        });
        
        const shortcuts = [
            { label: 'This year', calc: () => [moment().startOf('year'), moment().startOf('month')] },
            { label: 'Last year', calc: () => [moment().subtract(1, 'year').startOf('year'), moment().subtract(1, 'year').endOf('year')] },
            { label: 'Last 6 months', calc: () => [moment().subtract(5, 'months').startOf('month'), moment().startOf('month')] },
            { label: 'Last 12 months', calc: () => [moment().subtract(11, 'months').startOf('month'), moment().startOf('month')] }
        ];
        
        shortcuts.forEach(sc => {
            const $btn = $('<button></button>').text(sc.label).css({
                padding: '8px 12px',
                background: '#1a1a1a',
                border: '1px solid #575757',
                borderRadius: '4px',
                color: '#e0e0e0',
                cursor: 'pointer',
                fontSize: '14px',
                textAlign: 'left',
                whiteSpace: 'nowrap'
            }).on('mouseenter', function() {
                $(this).css('background', '#404040');
            }).on('mouseleave', function() {
                $(this).css('background', '#1a1a1a');
            }).on('click', function() {
                const [start, end] = sc.calc();
                applyRange(start, end);
            });
            $shortcuts.append($btn);
        });
        
        // Create calendars container
        const $calendars = $('<div class="calendars"></div>').css({
            display: 'flex',
            gap: '15px'
        });
        
        // Create year-month picker
        function createYearMonthPicker(year, isStart) {
            const $container = $('<div class="year-month-picker"></div>');
            
            const $header = $('<div class="picker-header"></div>').css({
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '10px'
            });
            
            const $prevBtn = $('<button>&lt;</button>').css({
                background: '#1a1a1a',
                border: '1px solid #575757',
                color: '#e0e0e0',
                padding: '5px 10px',
                borderRadius: '4px',
                cursor: 'pointer'
            }).on('click', function() {
                updateYearMonthPicker(isStart, year - 1);
            });
            
            const $yearLabel = $('<span></span>').text(year).css({
                color: '#e0e0e0',
                fontSize: '16px',
                fontWeight: 'bold'
            });
            
            const $nextBtn = $('<button>&gt;</button>').css({
                background: '#1a1a1a',
                border: '1px solid #575757',
                color: '#e0e0e0',
                padding: '5px 10px',
                borderRadius: '4px',
                cursor: 'pointer'
            }).on('click', function() {
                updateYearMonthPicker(isStart, year + 1);
            });
            
            $header.append($prevBtn, $yearLabel, $nextBtn);
            
            const $months = $('<div class="months-grid"></div>').css({
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: '8px'
            });
            
            const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
            monthNames.forEach((name, idx) => {
                const $month = $('<div></div>').text(name).css({
                    padding: '12px',
                    textAlign: 'center',
                    background: '#1a1a1a',
                    border: '1px solid #575757',
                    borderRadius: '4px',
                    color: '#e0e0e0',
                    cursor: 'pointer'
                }).on('mouseenter', function() {
                    if (!$(this).hasClass('active')) {
                        $(this).css('background', '#404040');
                    }
                }).on('mouseleave', function() {
                    if (!$(this).hasClass('active')) {
                        $(this).css('background', '#1a1a1a');
                    }
                }).on('click', function() {
                    const selectedMoment = moment().year(year).month(idx).startOf('month');
                    if (isStart) {
                        startMonth = selectedMoment;
                        if (endMonth && startMonth.isAfter(endMonth)) {
                            endMonth = startMonth.clone();
                        }
                    } else {
                        endMonth = selectedMoment;
                        if (startMonth && endMonth.isBefore(startMonth)) {
                            startMonth = endMonth.clone();
                        }
                    }
                    updateSelection();
                });
                $months.append($month);
            });
            
            $container.append($header, $months);
            return $container;
        }
        
        function updateYearMonthPicker(isStart, newYear) {
            $calendars.empty();
            const startYear = isStart ? newYear : (startMonth ? startMonth.year() : moment().subtract(1, 'year').year());
            const endYear = !isStart ? newYear : (endMonth ? endMonth.year() : moment().year());
            $calendars.append(createYearMonthPicker(startYear, true));
            $calendars.append(createYearMonthPicker(endYear, false));
            updateSelection();
        }
        
        function updateSelection() {
            $('.months-grid div').removeClass('active').css('background', '#1a1a1a');
            if (startMonth) {
                const startIdx = startMonth.month();
                $('.months-grid').eq(0).find('div').eq(startIdx).addClass('active').css('background', '#27ae60');
            }
            if (endMonth) {
                const endIdx = endMonth.month();
                $('.months-grid').eq(1).find('div').eq(endIdx).addClass('active').css('background', '#27ae60');
            }
        }
        
        function applyRange(start, end) {
            const fromMonth = start.format('YYYY-MM');
            const toMonth = end.format('YYYY-MM');
            
            // Update input display
            $input.val(start.format('MMM YYYY') + ' - ' + end.format('MMM YYYY'));
            
            // Close picker
            closePicker();
            
            // Call callback
            settings.onApply(fromMonth, toMonth);
        }
        
        // Initialize calendars
        const startYear = startMonth ? startMonth.year() : moment().subtract(1, 'year').year();
        const endYear = endMonth ? endMonth.year() : moment().year();
        
        $calendars.append(createYearMonthPicker(startYear, true));
        $calendars.append(createYearMonthPicker(endYear, false));
        
        // Create buttons
        const $buttons = $('<div class="picker-buttons"></div>').css({
            display: 'flex',
            justifyContent: 'flex-end',
            gap: '10px',
            marginTop: '15px',
            paddingTop: '15px',
            borderTop: '1px solid #575757'
        });
        
        const $cancelBtn = $('<button>Cancel</button>').css({
            padding: '8px 16px',
            background: '#1a1a1a',
            border: '1px solid #575757',
            borderRadius: '4px',
            color: '#e0e0e0',
            cursor: 'pointer'
        }).on('click', function() {
            closePicker();
        });
        
        const $applyBtn = $('<button>Apply</button>').css({
            padding: '8px 16px',
            background: '#27ae60',
            border: 'none',
            borderRadius: '4px',
            color: '#fff',
            cursor: 'pointer'
        }).on('click', function() {
            if (startMonth && endMonth) {
                applyRange(startMonth, endMonth);
            }
        });
        
        $buttons.append($cancelBtn, $applyBtn);
        
        // Create main container
        const $mainContainer = $('<div class="calendars-container"></div>').css({
            display: 'flex',
            gap: '15px'
        });
        $mainContainer.append($shortcuts, $calendars);
        
        // Build picker wrapper
        const $pickerWrapper = $('<div></div>').css({
            position: 'absolute',
            top: offset.top + input.outerHeight() + 5,
            left: offset.left,
            background: '#2d2d2d',
            border: '1px solid #575757',
            borderRadius: '4px',
            padding: '15px',
            zIndex: 9999,
            minWidth: '600px'
        });
        
        $pickerWrapper.append($mainContainer, $buttons);
        
        function closePicker() {
            $overlay.remove();
            $pickerWrapper.remove();
            pickerOpen = false;
        }
        
        $overlay.on('click', closePicker);
        
        $('body').append($overlay, $pickerWrapper);
        updateSelection();
    });
}
