import { DateUtils } from './date-utils.js';

describe('DateUtils', () => {
    describe('parseMonthKey', () => {
        test('extracts YYYY-MM from ISO date string', () => {
            expect(DateUtils.parseMonthKey('2025-01-15T00:00:00Z')).toBe('2025-01');
            expect(DateUtils.parseMonthKey('2026-01-01')).toBe('2026-01');
            expect(DateUtils.parseMonthKey('2025-12-31T23:59:59Z')).toBe('2025-12');
        });
        
        test('handles null or invalid input', () => {
            expect(DateUtils.parseMonthKey(null)).toBeNull();
            expect(DateUtils.parseMonthKey('')).toBeNull();
            expect(DateUtils.parseMonthKey('-')).toBeNull();
        });
    });
    
    describe('formatMonthLabel', () => {
        test('shows year for disambiguation', () => {
            expect(DateUtils.formatMonthLabel('2025-01')).toBe('Jan 2025');
            expect(DateUtils.formatMonthLabel('2026-01')).toBe('Jan 2026');
            expect(DateUtils.formatMonthLabel('2025-12')).toBe('Dec 2025');
        });
        
        test('handles empty input', () => {
            expect(DateUtils.formatMonthLabel('')).toBe('');
            expect(DateUtils.formatMonthLabel(null)).toBe('');
        });
    });
    
    describe('sortYearMonthKeys', () => {
        test('handles year transitions correctly', () => {
            const keys = ['2026-01', '2025-12', '2025-11', '2026-02', '2025-01'];
            const sorted = keys.sort(DateUtils.sortYearMonthKeys);
            expect(sorted).toEqual([
                '2025-01', '2025-11', '2025-12', '2026-01', '2026-02'
            ]);
        });
        
        test('maintains chronological order across multiple years', () => {
            const keys = ['2027-01', '2025-12', '2026-06', '2025-01'];
            const sorted = keys.sort(DateUtils.sortYearMonthKeys);
            expect(sorted).toEqual([
                '2025-01', '2025-12', '2026-06', '2027-01'
            ]);
        });
        
        test('handles same month in different years', () => {
            const keys = ['2026-01', '2025-01', '2027-01'];
            const sorted = keys.sort(DateUtils.sortYearMonthKeys);
            expect(sorted).toEqual([
                '2025-01', '2026-01', '2027-01'
            ]);
        });
    });
    
    describe('parseDate', () => {
        test('parses ISO format (YYYY-MM-DD)', () => {
            const date = DateUtils.parseDate('2025-01-15');
            expect(date).toBeInstanceOf(Date);
            expect(date.getFullYear()).toBe(2025);
            expect(date.getMonth()).toBe(0);
            expect(date.getDate()).toBe(15);
        });
        
        test('parses US format (M/D/YYYY)', () => {
            const date = DateUtils.parseDate('1/15/2025');
            expect(date).toBeInstanceOf(Date);
            expect(date.getFullYear()).toBe(2025);
            expect(date.getMonth()).toBe(0);
            expect(date.getDate()).toBe(15);
        });
        
        test('parses dash-separated format (M-D-YYYY)', () => {
            const date = DateUtils.parseDate('1-15-2025');
            expect(date).toBeInstanceOf(Date);
            expect(date.getFullYear()).toBe(2025);
        });
        
        test('returns null for invalid dates', () => {
            expect(DateUtils.parseDate('invalid')).toBeNull();
            expect(DateUtils.parseDate('2025-13-01')).toBeNull();
            expect(DateUtils.parseDate('')).toBeNull();
        });
    });
});

describe('Year-End Transition Scenarios', () => {
    test('December 2025 to January 2026 sorts correctly', () => {
        const monthlyData = {
            '2025-12': { maxProfit: 1000, actualProfit: 800, openValue: 200 },
            '2026-01': { maxProfit: 1200, actualProfit: 900, openValue: 300 }
        };
        
        const sortedKeys = Object.keys(monthlyData).sort();
        expect(sortedKeys).toEqual(['2025-12', '2026-01']);
        
        const labels = sortedKeys.map(ym => DateUtils.formatMonthLabel(ym));
        expect(labels).toEqual(['Dec 2025', 'Jan 2026']);
    });
    
    test('Mixed months across year boundary', () => {
        const months = [
            '2025-11', '2025-12', '2026-01', '2026-02', '2025-10'
        ];
        
        const sorted = months.sort(DateUtils.sortYearMonthKeys);
        expect(sorted).toEqual([
            '2025-10', '2025-11', '2025-12', '2026-01', '2026-02'
        ]);
        
        const labels = sorted.map(ym => DateUtils.formatMonthLabel(ym));
        expect(labels).toEqual([
            'Oct 2025', 'Nov 2025', 'Dec 2025', 'Jan 2026', 'Feb 2026'
        ]);
    });
});
