export const DateUtils = {
    parseMonthKey: (dateStr) => {
        if (!dateStr || dateStr === '-') return null;
        return dateStr.substring(0, 7);
    },
    
    formatMonthLabel: (yearMonth) => {
        if (!yearMonth) return '';
        const [year, month] = yearMonth.split('-');
        const date = new Date(year, parseInt(month) - 1);
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            year: 'numeric' 
        });
    },
    
    sortYearMonthKeys: (a, b) => {
        return a.localeCompare(b);
    },
    
    parseDate: (value) => {
        const formats = [
            /^\d{4}-\d{2}-\d{2}$/,
            /^\d{1,2}\/\d{1,2}\/\d{4}$/,
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
};
