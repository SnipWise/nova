# Theme Switching - Light/Dark Mode

## Overview

The chat interface now supports both light and dark themes with a convenient toggle button in the header.

## Features

- **Dynamic Theme Switching**: Switch between light and dark themes with a single click
- **Persistent Preference**: Your theme choice is saved in localStorage and restored on page reload
- **Smooth Transitions**: Theme changes are applied instantly with smooth CSS transitions
- **Complete Coverage**: Both the app UI and code syntax highlighting adapt to the selected theme

## Files

### CSS Files

| File | Description |
|------|-------------|
| [css/styles.css](../css/styles.css) | Dark theme styles (default) |
| [css/styles.light.css](../css/styles.light.css) | Light theme styles |

### Highlight.js Themes

| File | Description |
|------|-------------|
| [lib/github-dark.min.css](../lib/github-dark.min.css) | Dark syntax highlighting theme |
| [lib/github-light.min.css](../lib/github-light.min.css) | Light syntax highlighting theme |

### JavaScript

| File | Description |
|------|-------------|
| [js/theme.js](../js/theme.js) | Theme switcher Vue component |

## Usage

### Toggle Theme

Click the theme toggle button in the top-right corner of the header:
- 🌙 icon = Currently in dark mode (click to switch to light)
- ☀️ icon = Currently in light mode (click to switch to dark)

### Theme Persistence

Your theme preference is automatically saved to `localStorage` and will be restored when you:
- Refresh the page
- Close and reopen the browser
- Return to the application later

## Implementation Details

### Theme Switcher Component

Located in [js/theme.js](../js/theme.js), the `ThemeSwitcher` component:

```javascript
const ThemeSwitcher = {
    template: `
        <div class="theme-switcher">
            <button @click="toggleTheme" class="theme-toggle-btn">
                {{ isDark ? '☀️' : '🌙' }}
            </button>
        </div>
    `,
    // ...
}
```

### Dynamic CSS Loading

The component dynamically updates two `<link>` elements in the HTML:

```javascript
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
```

### Integration in Main App

The theme switcher is integrated into the main application in [js/app.js](../js/app.js):

```javascript
const App = {
    components: {
        ChatMessage,
        StatusBar,
        InputBar,
        ThemeSwitcher,  // ← Added
        // ...
    },
    // ...
}
```

And placed in the header template:

```html
<header class="header">
    <h1>🚀 Nova Crew Server Agent</h1>
    <theme-switcher />  <!-- ← Theme toggle button -->
    <status-bar ... />
</header>
```

## Color Schemes

### Dark Theme (Default)

- **Background**: `#1a1a1a` (Dark gray)
- **Foreground**: `#e0e0e0` (Light gray)
- **Primary Accent**: `#4fc3f7` (Light blue)
- **User Messages**: `#1e3a5f` (Dark blue)
- **Assistant Messages**: `#2d2d2d` (Medium gray)
- **Code Blocks**: `#1a1a1a` with GitHub Dark syntax

### Light Theme

- **Background**: `#f5f5f5` (Light gray)
- **Foreground**: `#212121` (Dark gray)
- **Primary Accent**: `#1976d2` (Blue)
- **User Messages**: `#e3f2fd` (Light blue)
- **Assistant Messages**: `#ffffff` (White)
- **Code Blocks**: `#f5f5f5` with GitHub Light syntax

## CSS Variables Approach (Alternative)

For future improvements, consider using CSS custom properties (variables) for even smoother theme switching:

```css
:root {
    --bg-primary: #1a1a1a;
    --text-primary: #e0e0e0;
    /* ... */
}

[data-theme="light"] {
    --bg-primary: #f5f5f5;
    --text-primary: #212121;
    /* ... */
}
```

This would allow switching themes by simply changing a `data-theme` attribute on the root element.

## Browser Compatibility

The theme switcher uses:
- **localStorage**: Supported in all modern browsers (IE8+)
- **Dynamic CSS loading**: Supported in all browsers
- **Vue 3 Composition API**: Requires modern browsers (ES6+)

## Testing

### Manual Testing

1. **Toggle Functionality**
   - Click the theme button
   - Verify the icon changes (🌙 ↔ ☀️)
   - Verify colors change instantly

2. **Persistence**
   - Change theme
   - Refresh the page
   - Verify theme is preserved

3. **Visual Inspection**
   - Check all UI elements (header, messages, buttons, modals)
   - Check code syntax highlighting
   - Verify contrast and readability

### Browser DevTools

Open DevTools (F12) and check:
- **Application → Local Storage** → Verify `theme` key is set to `"light"` or `"dark"`
- **Network** → Verify CSS files are loaded correctly
- **Console** → No errors related to theme switching

## Customization

### Adding a New Theme

1. Create a new CSS file (e.g., `styles.blue.css`)
2. Update [js/theme.js](../js/theme.js) to add the new theme option
3. Update the `applyTheme()` method to handle the new theme

Example:

```javascript
data() {
    return {
        currentTheme: 'dark'  // 'dark', 'light', or 'blue'
    };
},

methods: {
    applyTheme() {
        const themes = {
            dark: {
                highlight: 'lib/github-dark.min.css',
                app: 'css/styles.css?v=18'
            },
            light: {
                highlight: 'lib/github-light.min.css',
                app: 'css/styles.light.css?v=18'
            },
            blue: {
                highlight: 'lib/github-dark.min.css',
                app: 'css/styles.blue.css?v=18'
            }
        };

        const theme = themes[this.currentTheme];
        document.getElementById('highlight-theme').href = theme.highlight;
        document.getElementById('app-theme').href = theme.app;
    }
}
```

### System Theme Detection

To automatically detect the user's system theme preference:

```javascript
mounted() {
    // Check if user prefers dark mode
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const savedTheme = localStorage.getItem('theme');

    this.isDark = savedTheme ? savedTheme === 'dark' : prefersDark;
    this.applyTheme();
}
```

## Troubleshooting

### Theme Not Switching

1. **Check Browser Console**: Look for JavaScript errors
2. **Verify Files Exist**: Ensure all CSS files are in the correct location
3. **Check localStorage**: Clear localStorage and try again
4. **Hard Refresh**: Press Cmd+Shift+R (Mac) or Ctrl+Shift+F5 (Windows/Linux)

### Styles Not Applied

1. **Check CSS Version**: Ensure cache busting version (`?v=18`) is correct
2. **Inspect Element**: Use browser DevTools to verify which CSS file is loaded
3. **Check File Paths**: Verify relative paths in `applyTheme()` method

### Theme Resets After Reload

- **localStorage Disabled**: Check if localStorage is enabled in browser settings
- **Private/Incognito Mode**: localStorage may not persist in private browsing

## Future Enhancements

Potential improvements for the theme system:

1. **Multiple Themes**: Add more theme options (blue, purple, high-contrast, etc.)
2. **CSS Variables**: Migrate to CSS custom properties for smoother transitions
3. **Theme Gallery**: Show preview of themes before selecting
4. **Auto Theme**: Automatically switch based on time of day
5. **Per-User Themes**: Store theme preference in user profile (backend)
6. **Accessibility**: Add high-contrast theme for better accessibility

## Related Documentation

- [Cache Busting](./CACHE-BUSTING.md) - JavaScript version management
- [Local Dependencies](./LOCAL-DEPENDENCIES.md) - Local JavaScript dependencies
- [Custom Routes Examples](./CUSTOM-ROUTES-EXAMPLES.md) - Custom API routes

---

**Status**: ✅ Complete
**Version**: v18
**Date**: 2026-01-07
