# Cache Busting - JavaScript Version Management

## What is Cache Busting?

**Cache busting** is a technique that forces the browser to download the new version of a JavaScript/CSS file instead of using the cached version.

## Problem

When you modify a JavaScript file, the browser may continue using the old cached version, preventing your changes from appearing immediately.

### Example of the Problem

```javascript
// You modify InputBar.js to remove confirm()
const handleResetMemory = () => {
    emit('reset-memory');  // New version
};

// But the browser still uses the old cached version:
const handleResetMemory = () => {
    if (confirm('...')) {  // Old version
        emit('reset-memory');
    }
};
```

## Solution: Version Query Parameters

### Concept

Add a version number as a query parameter to the script URL:

```html
<!-- Without version (may use cache) -->
<script src="js/app.js"></script>

<!-- With version (forces download) -->
<script src="js/app.js?v=1"></script>
```

For the browser, `app.js?v=1` and `app.js?v=2` are **different** URLs, so it downloads the new file.

## Implementation in Our Application

### index.html

```html
<!-- App Scripts -->
<script src="js/api.js?v=3"></script>
<script src="js/markdown.js?v=3"></script>
<script src="js/components/ChatMessage.js?v=3"></script>
<script src="js/components/OperationControls.js?v=3"></script>
<script src="js/components/StatusBar.js?v=3"></script>
<script src="js/components/InputBar.js?v=3"></script>
<script src="js/components/Modal.js?v=3"></script>
<script src="js/app.js?v=3"></script>
```

### When to Increment the Version?

#### ✅ Increment When:

1. **You modify JavaScript code**
   ```javascript
   // Before modification: ?v=3
   // After modification: ?v=4
   ```

2. **Users don't see your changes**
   - Symptom: "I fixed the bug but it's still there"
   - Action: Increment the version

3. **You deploy to production**
   - Always increment for deployments

#### ❌ DON'T Increment When:

1. You only modify HTML (index.html)
2. You only modify inline CSS
3. You only touch backend files (Go)

### Modification Process

```bash
# 1. Modify JavaScript code
vim js/components/InputBar.js

# 2. Increment version in index.html
# Change ?v=3 to ?v=4

# 3. Reload the browser
# Simple F5 is now sufficient
```

## Version History

### v1 (Initial Version)
- First version of the application
- No query parameters

### v2 (Add Modals)
- Removal of native `confirm()`
- Addition of custom modal system
- Modal for Clear Memory
- Modal for View Messages
- Modal for View Models
- Modal for Reset Operations

### v3 (Fix Unknown Label)
- Removal of "UNKNOWN" label in message list
- Addition of `v-if="msg.role"` to hide empty labels

## Alternatives to Manual Versioning

### 1. Content Hash (Build Tools)

With build tools like Webpack/Vite:

```html
<!-- Hash changes automatically when file changes -->
<script src="js/app.a3d8f9b2.js"></script>
```

**Advantages**: Automatic, precise per file
**Disadvantages**: Requires a build tool

### 2. Timestamp

```html
<script src="js/app.js?t=1704654321"></script>
```

**Advantages**: Unique per build
**Disadvantages**: Cache invalidated even without change

### 3. Git Commit Hash

```html
<script src="js/app.js?v=a3d8f9b"></script>
```

**Advantages**: Traceable in Git
**Disadvantages**: Requires automation

## Our Choice: Simple Manual Versioning

We use manual versioning (`?v=1`, `?v=2`, etc.) because:

✅ **Simplicity**: No build tool required
✅ **Control**: You decide when to invalidate cache
✅ **CDN-friendly**: Works with Vue.js and CDN libraries
✅ **Production-ready**: Sufficient for most cases

## Useful Commands

### Find Current Version

```bash
grep "js/app.js?v=" web/index.html
```

### Replace All Versions

```bash
# Change from v=3 to v=4 everywhere
sed -i '' 's/\?v=3/\?v=4/g' web/index.html
```

### Verify All Scripts Have Same Version

```bash
grep "\.js?v=" web/index.html | grep -o "v=[0-9]*" | sort -u
# Should display a single line: v=3
```

## Best Practices

### ✅ DO

- Systematically increment after JS modification
- Use the same number for all scripts
- Document important changes
- Test after each increment

### ❌ DON'T

- Forget to increment after modification
- Use different versions for each file
- Skip numbers (v=3 → v=5)
- Reuse an old number

## Troubleshooting

### Changes Still Not Visible

1. **Verify version in index.html**
   ```bash
   grep "js/app.js?v=" web/index.html
   ```

2. **Hard Refresh in Browser**
   - Mac: `Cmd + Shift + R`
   - Windows/Linux: `Ctrl + Shift + F5`

3. **Inspect in DevTools**
   - F12 → Network → Filter "js" → Verify loaded URLs
   - Look for `app.js?v=X` in the list

4. **Clear Cache Manually**
   - Chrome: Settings → Privacy → Clear browsing data → Cached images and files

5. **Incognito Mode**
   - Test in an incognito window (no cache)

### File is Loaded But Changes Don't Work

- Verify you modified the correct file
- Check there are no JavaScript errors in the console (F12)
- Verify the web server is serving the latest files

## Complete Example

### Scenario: Add a New Feature

```bash
# 1. Modify code
echo "console.log('New feature');" >> js/app.js

# 2. Open index.html
vim web/index.html

# 3. Change manually
# Before:
# <script src="js/app.js?v=3"></script>
# After:
# <script src="js/app.js?v=4"></script>

# 4. Apply to all scripts
# Replace ?v=3 with ?v=4 for all scripts

# 5. Reload browser
# F5 or Cmd+R

# 6. Verify in DevTools
# Network → See app.js?v=4 loaded
```

## Conclusion

Cache busting with query parameters is a simple and effective solution for our CDN-based Vue.js application. By systematically incrementing the version number after each JavaScript modification, we ensure users always receive the latest version of the code.

**Current version**: `v=3` (as of 2026-01-07)

**Next version**: `v=4` (upon next JS modification)
