# API Configuration Guide

This document explains how to configure the API base URL for the Nova Crew Server web interface.

## Configuration Methods

The API base URL can be configured in multiple ways, with the following priority order:

1. **window.NOVA_API_URL** (highest priority)
2. **localStorage** `nova_api_url`
3. **Auto-detection** from current page URL
4. **Default** `http://localhost:8080` (lowest priority)

---

## Method 1: config.js File (Recommended)

Edit the `config.js` file in the web root directory:

```javascript
// config.js
window.NOVA_API_URL = 'http://localhost:3000';
```

**Pros:**
- Easy to configure
- Version controlled
- Works for all users

**Use cases:**
- Changing the default port
- Deploying to production with a fixed API URL

---

## Method 2: localStorage (Developer Console)

Set the URL via browser console or application code:

```javascript
// Set custom URL
localStorage.setItem('nova_api_url', 'http://localhost:3000');

// Clear custom URL (revert to default)
localStorage.removeItem('nova_api_url');

// Check current URL
localStorage.getItem('nova_api_url');
```

**Pros:**
- No file changes needed
- Per-browser configuration
- Persists across page reloads

**Use cases:**
- Testing different API endpoints
- Local development with multiple ports
- Per-developer configuration

---

## Method 3: Auto-detection

If no configuration is set, the interface automatically detects the current hostname and uses port 8080:

- Page URL: `http://192.168.1.100:5000/`
- API URL: `http://192.168.1.100:8080`

**Pros:**
- Zero configuration
- Works across different environments

**Use cases:**
- Deploying to remote servers
- When API and web UI are on the same host

---

## Method 4: Default Fallback

If all other methods fail, defaults to:
```
http://localhost:8080
```

---

## Examples

### Example 1: Change Port to 3000

**config.js:**
```javascript
window.NOVA_API_URL = 'http://localhost:3000';
```

### Example 2: Remote Server

**config.js:**
```javascript
window.NOVA_API_URL = 'http://192.168.1.100:8080';
```

### Example 3: HTTPS Production

**config.js:**
```javascript
window.NOVA_API_URL = 'https://api.example.com';
```

### Example 4: Testing Multiple Endpoints

**Browser Console:**
```javascript
// Test endpoint 1
localStorage.setItem('nova_api_url', 'http://localhost:8080');
location.reload();

// Test endpoint 2
localStorage.setItem('nova_api_url', 'http://localhost:3000');
location.reload();

// Back to default
localStorage.removeItem('nova_api_url');
location.reload();
```

---

## Verification

To verify which API URL is being used:

1. Open browser DevTools Console (F12)
2. Run:
   ```javascript
   console.log(window.CrewServerAPI ? new CrewServerAPI().baseURL : 'API not loaded');
   ```

Or check the Network tab for API requests to see the actual URLs being called.

---

## CORS Considerations

When using a different host/port for the API, ensure the Nova Crew Server is configured to allow CORS from the web interface origin.

The server should include these headers:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Accept
```

---

## Troubleshooting

### API calls failing with CORS errors
- Ensure the API server allows CORS from your web interface origin
- Check that the API URL is correct (no trailing slash)

### API URL not updating
1. Clear browser cache (Ctrl+Shift+R or Cmd+Shift+R)
2. Clear localStorage: `localStorage.clear()`
3. Check config.js is loaded before api.js in index.html

### Auto-detection not working
- Check that `window.location.hostname` is not empty
- Manually set URL via config.js or localStorage
