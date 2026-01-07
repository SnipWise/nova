/**
 * Theme Switcher Component
 * Manages light/dark theme switching
 */

const ThemeSwitcher = {
    template: `
        <div class="theme-switcher">
            <button
                @click="toggleTheme"
                class="theme-toggle-btn"
                :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
            >
                {{ isDark ? '☀️' : '🌙' }}
            </button>
        </div>
    `,

    data() {
        return {
            isDark: true
        };
    },

    mounted() {
        // Load theme preference from localStorage
        const savedTheme = localStorage.getItem('theme') || 'dark';
        this.isDark = savedTheme === 'dark';
        this.applyTheme();
    },

    methods: {
        toggleTheme() {
            this.isDark = !this.isDark;
            this.applyTheme();
            localStorage.setItem('theme', this.isDark ? 'dark' : 'light');
        },

        applyTheme() {
            const highlightTheme = document.getElementById('highlight-theme');
            const appTheme = document.getElementById('app-theme');

            if (this.isDark) {
                highlightTheme.href = 'lib/github-dark.min.css';
                appTheme.href = 'css/styles.css?v=18';
            } else {
                highlightTheme.href = 'lib/github-light.min.css';
                appTheme.href = 'css/styles.light.css?v=18';
            }
        }
    }
};

// Make it globally available
window.ThemeSwitcher = ThemeSwitcher;
