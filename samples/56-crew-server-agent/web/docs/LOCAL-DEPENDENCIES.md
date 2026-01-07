# Local JavaScript Dependencies

## 📋 Summary

JavaScript dependencies (Vue.js, Marked.js, Highlight.js) have been downloaded locally to eliminate reliance on external CDNs.

## 🎯 Objective

- **Independence**: No need for Internet connection to develop
- **Performance**: Faster loading (no external requests)
- **Reliability**: No risk of CDN unavailability
- **Security**: Full control over executed code
- **Deployment**: Self-contained application, easy to deploy

## 📁 File Structure

### Before (CDN)
```
web/
├── index.html (CDN links)
└── js/
    └── ...
```

### After (Local)
```
web/
├── index.html (local links)
├── lib/                              ← NEW
│   ├── vue.global.prod.js           (144 KB)
│   ├── marked.min.js                (34 KB)
│   ├── highlight.min.js             (119 KB)
│   └── github-dark.min.css          (1.3 KB)
└── js/
    └── ...
```

## 📦 Downloaded Dependencies

### 1. Vue.js 3.4.15
- **File**: `lib/vue.global.prod.js`
- **Size**: 144 KB
- **Source**: https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js
- **Usage**: Vue.js 3 framework (Composition API)

### 2. Marked.js 11.1.1
- **File**: `lib/marked.min.js`
- **Size**: 34 KB
- **Source**: https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js
- **Usage**: Markdown parsing and rendering

### 3. Highlight.js 11.9.0
- **File**: `lib/highlight.min.js`
- **Size**: 119 KB
- **Source**: https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js
- **Usage**: Code syntax highlighting

### 4. Highlight.js Theme (GitHub Dark)
- **File**: `lib/github-dark.min.css`
- **Size**: 1.3 KB
- **Source**: https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css
- **Usage**: Dark highlighting theme

## 📝 HTML Modifications

### index.html

**Before**:
```html
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">
<link rel="stylesheet" href="css/styles.css?v=5">
<!-- ... -->
<script src="https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js"></script>
<script src="https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
```

**After**:
```html
<link rel="stylesheet" href="lib/github-dark.min.css">
<link rel="stylesheet" href="css/styles.css?v=5">
<!-- ... -->
<!-- Dependencies (Local) -->
<script src="lib/vue.global.prod.js"></script>
<script src="lib/marked.min.js"></script>
<script src="lib/highlight.min.js"></script>
```

## 🚀 Benefits

### 1. Offline Development
- ✅ No need for Internet connection
- ✅ Local development without external dependencies
- ✅ Works on isolated networks

### 2. Performance
```
Before (CDN):
- DNS query to CDN
- Variable network latency
- Depends on Internet speed

After (Local):
- Files served locally
- Minimal latency
- Always fast
```

### 3. Reliability
- ✅ No risk of CDN unavailability
- ✅ No unexpected changes (frozen versions)
- ✅ Full control over versions

### 4. Security
- ✅ No requests to third-party domains
- ✅ Full control over executed code
- ✅ No risk of CDN compromise
- ✅ Compliant with strict security policies

### 5. Deployment
- ✅ Self-contained application
- ✅ Single directory to deploy
- ✅ Works without Internet access
- ✅ Easy to package (Docker, etc.)

## 📊 Comparison

| Aspect | CDN | Local | Winner |
|--------|-----|-------|---------|
| **First visit** | Fast (CDN cache) | Fast (local) | Tie |
| **Subsequent visits** | Very fast (cache) | Very fast (cache) | Tie |
| **Offline** | ❌ Doesn't work | ✅ Works | **Local** |
| **Reliability** | Depends on CDN | Always available | **Local** |
| **Security** | External dependency | Full control | **Local** |
| **Bundle size** | 0 KB initial | +298 KB | CDN |
| **Network requests** | +3 requests | 0 external requests | **Local** |

## 🔧 Updating Dependencies

### Update Vue.js

```bash
cd web/lib
curl -o vue.global.prod.js https://cdn.jsdelivr.net/npm/vue@3.5.0/dist/vue.global.prod.js
```

### Update Marked.js

```bash
cd web/lib
curl -o marked.min.js https://cdn.jsdelivr.net/npm/marked@12.0.0/marked.min.js
```

### Update Highlight.js

```bash
cd web/lib
curl -o highlight.min.js https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/highlight.min.js
curl -o github-dark.min.css https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/styles/github-dark.min.css
```

## 🧪 Verification

### 1. Verify files exist

```bash
ls -lh web/lib/
```

**Expected result**:
```
-rw-r--r--  github-dark.min.css    (1.3K)
-rw-r--r--  highlight.min.js       (119K)
-rw-r--r--  marked.min.js          (34K)
-rw-r--r--  vue.global.prod.js     (144K)
```

### 2. Test loading

```bash
# Start the server
cd samples/56-crew-server-agent
go run main.go

# Open http://localhost:3000
```

### 3. Verify in DevTools

**Network Tab**:
- ✅ `vue.global.prod.js` loaded from `localhost:3000`
- ✅ `marked.min.js` loaded from `localhost:3000`
- ✅ `highlight.min.js` loaded from `localhost:3000`
- ✅ No requests to external CDNs

**Console**:
- ✅ `Vue` is defined
- ✅ `marked` is defined
- ✅ `hljs` is defined

### 4. Test Offline

1. Start the application
2. Disconnect Internet
3. Refresh the page
4. ✅ Application still works

## 📦 Total Size

| Dependency | Size | Percentage |
|------------|------|------------|
| Vue.js | 144 KB | 48% |
| Highlight.js | 119 KB | 40% |
| Marked.js | 34 KB | 11% |
| GitHub Dark CSS | 1.3 KB | 1% |
| **Total** | **298 KB** | **100%** |

**Note**: All dependencies are minified and production-ready.

## 🔒 File Integrity

To verify file integrity (optional):

```bash
# Generate checksums
cd web/lib
shasum -a 256 *.js *.css > checksums.txt

# Verify checksums
shasum -a 256 -c checksums.txt
```

## 📚 Versions Used

| Library | Version | Release date |
|---------|---------|--------------|
| Vue.js | 3.4.15 | Jan 2024 |
| Marked.js | 11.1.1 | Dec 2023 |
| Highlight.js | 11.9.0 | Nov 2023 |

## 🎯 Best Practices

### 1. Version Dependencies

Files in `lib/` should be committed to Git:

```bash
git add web/lib/
git commit -m "Add local JavaScript dependencies"
```

### 2. Document Versions

Keep track of versions in a `lib/VERSIONS.md` file:

```markdown
# Dependency Versions

- Vue.js: 3.4.15
- Marked.js: 11.1.1
- Highlight.js: 11.9.0
```

### 3. Test After Update

Always test the application after updating a dependency:

```bash
# Update
curl -o lib/vue.global.prod.js https://...

# Test
go run main.go
# Open http://localhost:3000
# Verify everything works
```

## 🚫 Reverting to CDNs (if needed)

If you want to revert to CDNs:

```html
<!-- In index.html, replace -->
<link rel="stylesheet" href="lib/github-dark.min.css">
<!-- with -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">

<!-- And same for scripts -->
<script src="https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js"></script>
<script src="https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
```

## 🌐 Deployment

### Docker

Local dependencies facilitate Docker deployment:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./samples/56-crew-server-agent

FROM alpine:latest
COPY --from=builder /app/server /server
COPY ./samples/56-crew-server-agent/web /web
EXPOSE 3000 8080
CMD ["/server"]
```

**Advantage**: Everything is included, no Internet needed at runtime.

### Production

In production, local dependencies offer:
- ✅ Reproducible deployment
- ✅ No external network dependency
- ✅ Full control over versions
- ✅ Better security

## 📈 Performance Impact

### First Visit

| Metric | CDN | Local |
|--------|-----|-------|
| DNS queries | 3 | 0 |
| HTTP requests | 3 external | 3 local |
| Latency | Variable | Minimal |
| Total time | ~500ms | ~50ms |

### Subsequent Visits

| Metric | CDN | Local |
|--------|-----|-------|
| Cache hit | ✅ (if same CDN) | ✅ (always) |
| Total time | ~10ms | ~10ms |

## ✅ Migration Checklist

- [x] Create `web/lib/` folder
- [x] Download Vue.js
- [x] Download Marked.js
- [x] Download Highlight.js
- [x] Download CSS theme
- [x] Modify `index.html` to use local files
- [x] Test the application
- [x] Verify in DevTools (no CDN requests)
- [x] Test offline
- [x] Document versions

## 📝 Conclusion

Using local JavaScript dependencies makes the application:
- More **reliable** (no CDN dependency)
- More **secure** (full control)
- More **performant** (no network latency)
- More **simple to deploy** (self-contained)

**Cost**: +298 KB of static files (negligible)

**Benefit**: Completely autonomous application 🎉

---

**Status**: ✅ Complete
**Date**: 2026-01-07
**Total size**: 298 KB (minified)
