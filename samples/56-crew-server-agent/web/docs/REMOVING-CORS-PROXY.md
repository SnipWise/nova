# Se Passer du Proxy CORS

## Situation Actuelle

Actuellement, l'architecture est:

```
Browser (localhost:3000)
    ↓
CORS Proxy (localhost:8081)
    ↓
Nova Crew Server (localhost:8080)
```

Le proxy est nécessaire car le serveur Nova ne retourne pas les headers CORS requis pour tous les endpoints.

## Pourquoi Avons-Nous Besoin du Proxy?

### Le Problème CORS

**CORS (Cross-Origin Resource Sharing)** est une sécurité du navigateur qui bloque les requêtes HTTP entre différentes origines.

**Origine** = Protocole + Domaine + Port

```
http://localhost:3000  ← Front-end
http://localhost:8080  ← Backend (origine différente!)
```

Le navigateur bloque car:
- Ports différents (3000 ≠ 8080)
- Même si c'est localhost

### Headers CORS Manquants

Le serveur Nova Crew Server ne retourne pas ces headers pour certains endpoints:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

**Endpoints affectés**:
- `/models`
- `/memory/messages/list`
- `/memory/messages/context-size`
- `/memory/reset`
- `/operation/validate`
- `/operation/cancel`
- `/operation/reset`

**Endpoint OK**:
- `/completion` (a déjà les headers CORS)

## Solution 1: Modifier le Backend Go (Recommandé)

### Ajouter un Middleware CORS Global

**Fichier**: `samples/56-crew-server-agent/main.go`

```go
// Middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Autoriser toutes les origines (ou spécifier l'origine du front)
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Méthodes HTTP autorisées
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        // Headers autorisés
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

        // Credentials (si nécessaire pour cookies/auth)
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        // Preflight request (OPTIONS)
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Continuer vers le handler suivant
        next.ServeHTTP(w, r)
    })
}

func main() {
    // ... code existant ...

    // Créer le mux
    mux := http.NewServeMux()

    // Enregistrer les routes
    mux.HandleFunc("/completion", handleCompletion)
    mux.HandleFunc("/models", handleModels)
    mux.HandleFunc("/memory/messages/list", handleMessagesList)
    mux.HandleFunc("/memory/messages/context-size", handleContextSize)
    mux.HandleFunc("/memory/reset", handleMemoryReset)
    mux.HandleFunc("/operation/validate", handleOperationValidate)
    mux.HandleFunc("/operation/cancel", handleOperationCancel)
    mux.HandleFunc("/operation/reset", handleOperationReset)
    mux.HandleFunc("/health", handleHealth)

    // Appliquer le middleware CORS
    handler := corsMiddleware(mux)

    // Démarrer le serveur
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### Avantages de Cette Approche

✅ **Une seule source de vérité**: Un seul serveur
✅ **Plus simple**: Pas besoin de proxy
✅ **Moins de latence**: Pas de saut réseau supplémentaire
✅ **Production-ready**: Fonctionne directement en production
✅ **Debugging facile**: Un seul point d'entrée

### Migration

#### 1. Modifier main.go

Ajouter le middleware CORS comme montré ci-dessus.

#### 2. Modifier api.js

Changer l'URL de base:

```javascript
// Avant (avec proxy)
constructor(baseURL = 'http://localhost:8081') {
    this.baseURL = baseURL;
}

// Après (direct)
constructor(baseURL = 'http://localhost:8080') {
    this.baseURL = baseURL;
}
```

#### 3. Arrêter le Proxy

```bash
# Plus besoin de ça
cd web/proxy
go run main.go  # ← Supprimer
```

#### 4. Tester

```bash
# 1. Démarrer le serveur Nova (avec CORS)
cd samples/56-crew-server-agent
go run main.go

# 2. Ouvrir le navigateur
open http://localhost:3000

# 3. Vérifier dans DevTools
# Network → Voir les requêtes vers localhost:8080
# Pas d'erreur CORS
```

## Solution 2: Utiliser une Library CORS Go

### Option A: github.com/rs/cors

```go
import "github.com/rs/cors"

func main() {
    mux := http.NewServeMux()

    // Routes...
    mux.HandleFunc("/completion", handleCompletion)
    // etc.

    // CORS avec library
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    handler := c.Handler(mux)
    http.ListenAndServe(":8080", handler)
}
```

**Installation**:
```bash
go get github.com/rs/cors
```

### Option B: github.com/gorilla/handlers

```go
import "github.com/gorilla/handlers"

func main() {
    mux := http.NewServeMux()

    // Routes...

    // CORS headers
    headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
    originsOk := handlers.AllowedOrigins([]string{"*"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

    handler := handlers.CORS(headersOk, originsOk, methodsOk)(mux)
    http.ListenAndServe(":8080", handler)
}
```

## Solution 3: Configuration Nginx (Production)

En production, vous pouvez utiliser Nginx comme reverse proxy:

```nginx
server {
    listen 80;
    server_name example.com;

    # Servir le front-end statique
    location / {
        root /var/www/nova-chat;
        try_files $uri $uri/ /index.html;
    }

    # Proxy vers le backend
    location /api/ {
        proxy_pass http://localhost:8080/;

        # Headers CORS
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization";

        # WebSocket/SSE support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_buffering off;
        proxy_cache off;
    }
}
```

**Avantages**:
- Même origine (pas de CORS)
- Caching, load balancing, SSL
- Standard en production

## Solution 4: Servir le Front depuis Go

Intégrer le front-end dans le serveur Go:

```go
func main() {
    mux := http.NewServeMux()

    // API endpoints
    mux.HandleFunc("/api/completion", handleCompletion)
    mux.HandleFunc("/api/models", handleModels)
    // etc.

    // Servir les fichiers statiques
    fs := http.FileServer(http.Dir("./web"))
    mux.Handle("/", fs)

    // Pas besoin de CORS car même origine!
    http.ListenAndServe(":8080", mux)
}
```

**Modification dans api.js**:
```javascript
constructor(baseURL = '/api') {  // Chemin relatif
    this.baseURL = baseURL;
}
```

**Avantages**:
- ✅ Pas de CORS (même origine)
- ✅ Un seul serveur
- ✅ Un seul port
- ✅ Simple à déployer

**Inconvénients**:
- ❌ Mélange front et back
- ❌ Moins flexible pour le développement

## Comparaison des Solutions

| Solution | Complexité | Dev | Prod | Recommandé |
|----------|-----------|-----|------|------------|
| **Middleware CORS Go** | ⭐ Facile | ✅ | ✅ | **OUI** |
| Library CORS | ⭐⭐ Moyen | ✅ | ✅ | Oui |
| Nginx | ⭐⭐⭐ Avancé | ❌ | ✅ | Production |
| Servir Front depuis Go | ⭐⭐ Moyen | ⚠️ | ✅ | Si simple |
| **Garder le Proxy** | ⭐ Facile | ✅ | ❌ | Dev only |

## Recommandation

### Pour le Développement

**Option recommandée**: Modifier `main.go` pour ajouter le middleware CORS manuel.

**Raisons**:
- Pas de dépendance externe
- Code simple et compréhensible
- Fonctionne immédiatement
- Prêt pour la production

### Pour la Production

**Option 1**: Middleware CORS + serveur séparé pour le front
- Front sur CDN/S3/Nginx
- Backend sur serveur dédié

**Option 2**: Nginx reverse proxy
- Front et back derrière Nginx
- SSL/TLS, caching, load balancing

**Option 3**: Servir le front depuis Go
- Simple, un seul déploiement
- Bon pour applications internes

## Code Complet du Middleware

### Version Basique

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Version Production

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        // Whitelist d'origines autorisées
        allowedOrigins := map[string]bool{
            "http://localhost:3000":  true,
            "https://example.com":     true,
            "https://app.example.com": true,
        }

        // Vérifier si l'origine est autorisée
        if allowedOrigins[origin] {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        } else if origin == "" {
            // Requête same-origin (pas de header Origin)
            w.Header().Set("Access-Control-Allow-Origin", "*")
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Max-Age", "86400") // 24h cache

        // Preflight
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Tester CORS

### Vérifier les Headers

```bash
# Tester avec curl
curl -I -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -X OPTIONS http://localhost:8080/models

# Devrait retourner:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, OPTIONS
# Access-Control-Allow-Headers: Content-Type
```

### Tester depuis le Navigateur

```javascript
// Console DevTools (F12)
fetch('http://localhost:8080/models')
  .then(res => res.json())
  .then(data => console.log('Success:', data))
  .catch(err => console.error('CORS Error:', err));
```

## Checklist de Migration

- [ ] Ajouter le middleware CORS à `main.go`
- [ ] Modifier `baseURL` dans `api.js` (8081 → 8080)
- [ ] Incrémenter la version des scripts (`?v=4`)
- [ ] Redémarrer le serveur Nova
- [ ] Tester tous les endpoints dans le navigateur
- [ ] Vérifier qu'il n'y a plus d'erreurs CORS dans la console
- [ ] Tester le streaming SSE
- [ ] Tester la validation des tools
- [ ] Supprimer le répertoire `web/proxy/` (optionnel)
- [ ] Mettre à jour la documentation

## Conclusion

Le proxy CORS est une solution temporaire pour le développement. Pour une application production-ready, il est préférable d'ajouter le support CORS directement dans le serveur Go backend.

**Action recommandée**: Ajouter le middleware CORS à `main.go` et supprimer le proxy.

**Bénéfices**:
- Architecture simplifiée
- Meilleure performance
- Code plus maintenable
- Prêt pour la production
