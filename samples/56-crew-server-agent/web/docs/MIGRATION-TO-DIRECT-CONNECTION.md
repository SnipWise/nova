# Migration: Suppression du Proxy CORS

## Résumé des Modifications

Le SDK Nova a été modifié pour intégrer le support CORS directement. Le proxy CORS intermédiaire n'est plus nécessaire.

## Architecture

### Avant (avec proxy)
```
Browser (localhost:3000)
    ↓
CORS Proxy (localhost:8081)
    ↓
Nova Crew Server (localhost:8080)
```

### Après (direct)
```
Browser (localhost:3000)
    ↓
Nova Crew Server (localhost:8080) ✅ CORS intégré
```

## Modifications Apportées

### 1. SDK - crew.server.agent.go

#### Ajout du Middleware CORS
Un nouveau middleware `corsMiddleware()` a été ajouté qui:
- Ajoute les headers CORS à toutes les réponses
- Gère les requêtes preflight OPTIONS
- Autorise toutes les origines (configurable pour production)

```go
// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

#### Modification de StartServer()
Le middleware est maintenant appliqué automatiquement:

```go
func (agent *CrewServerAgent) StartServer() error {
	mux := http.NewServeMux()

	// Expose mux for custom routes
	agent.Mux = mux

	// Routes...
	mux.HandleFunc("POST /completion", agent.handleCompletion)
	// ... autres routes

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	agent.Log.Info("🚀 Server started on http://localhost%s", agent.Port)
	return http.ListenAndServe(agent.Port, handler)
}
```

#### Exposition du Mux
Le champ `Mux *http.ServeMux` est maintenant public et assigné dans `StartServer()`:

```go
type CrewServerAgent struct {
	*serverbase.BaseServerAgent

	// ... autres champs

	// HTTP server multiplexer for custom routes
	Mux *http.ServeMux
}
```

Cela permet d'ajouter des routes personnalisées:

```go
crewAgent.Mux.HandleFunc("GET /custom/endpoint", myHandler)
crewAgent.StartServer()
```

### 2. Frontend - api.js

#### Changement de Port
```javascript
// Avant
const API_BASE_URL = 'http://localhost:8081';

// Après
const API_BASE_URL = 'http://localhost:8080';
```

### 3. Frontend - index.html

#### Cache Busting
Version incrémentée de v=3 à v=4 pour forcer le refresh:

```html
<script src="js/api.js?v=4"></script>
<script src="js/markdown.js?v=4"></script>
<!-- ... etc -->
```

## Instructions de Test

### 1. Compiler le SDK Modifié

```bash
cd /Users/k33g/Library/CloudStorage/Dropbox/SnipWise/nova
go mod tidy
```

### 2. Arrêter le Proxy CORS

```bash
# Si le proxy tourne encore, l'arrêter
# Plus besoin de:
# cd samples/56-crew-server-agent/web/proxy
# go run main.go
```

### 3. Démarrer le Serveur Nova

```bash
cd samples/56-crew-server-agent
go run main.go
```

Vérifier dans les logs:
```
🚀 Server started on http://localhost:8080
```

### 4. Ouvrir le Navigateur

```bash
# Si vous utilisez un serveur web pour le frontend
open http://localhost:3000

# OU ouvrir directement index.html
open web/index.html
```

### 5. Vérifier la Connexion Directe

#### Dans DevTools (F12)

**Network Tab:**
- Les requêtes doivent maintenant aller directement vers `localhost:8080`
- Plus aucune requête vers `localhost:8081`

**Console:**
- Pas d'erreur CORS
- Les messages SSE doivent s'afficher normalement

#### Headers CORS Vérifiés

Vous pouvez tester avec curl:

```bash
# Requête preflight (OPTIONS)
curl -I -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:8080/models

# Devrait retourner:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# Access-Control-Allow-Headers: Content-Type, Authorization, Accept
```

```bash
# Requête GET normale
curl -H "Origin: http://localhost:3000" \
  http://localhost:8080/models

# Devrait retourner les models + headers CORS
```

### 6. Tester Toutes les Fonctionnalités

- ✅ Envoi de message
- ✅ Streaming SSE
- ✅ Validation de tools
- ✅ Annulation de tools
- ✅ Clear Memory
- ✅ View Messages
- ✅ View Models
- ✅ Reset Operations
- ✅ Context size update

## Nettoyage (Optionnel)

Une fois que tout fonctionne, vous pouvez supprimer le répertoire proxy:

```bash
rm -rf samples/56-crew-server-agent/web/proxy/
```

## Rollback (si problème)

Si vous rencontrez des problèmes, vous pouvez revenir en arrière:

### 1. Restaurer api.js
```javascript
const API_BASE_URL = 'http://localhost:8081';
```

### 2. Restaurer index.html
```html
<script src="js/api.js?v=3"></script>
<!-- ... -->
```

### 3. Redémarrer le proxy
```bash
cd samples/56-crew-server-agent/web/proxy
go run main.go
```

## Configuration Production

Pour la production, vous voudrez peut-être restreindre les origines autorisées.

### Option 1: Modifier le Middleware (crew.server.agent.go)

```go
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Whitelist d'origines autorisées
		allowedOrigins := map[string]bool{
			"https://app.example.com": true,
			"https://example.com":     true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

### Option 2: Ajouter une Option de Configuration

Vous pourriez créer une nouvelle option pour `NewAgent()`:

```go
// WithAllowedOrigins sets the allowed CORS origins
func WithAllowedOrigins(origins []string) CrewServerAgentOption {
	return func(agent *CrewServerAgent) error {
		agent.allowedOrigins = origins
		return nil
	}
}
```

## Bénéfices de la Migration

✅ **Architecture simplifiée** - Un seul serveur au lieu de deux
✅ **Moins de latence** - Pas de saut réseau supplémentaire
✅ **Moins de points de défaillance** - Un processus en moins
✅ **Production-ready** - CORS géré nativement
✅ **Debugging facile** - Un seul point d'entrée
✅ **Extensibilité** - Possibilité d'ajouter des routes personnalisées via `agent.Mux`

## Support

Si vous rencontrez des problèmes:

1. Vérifier que le SDK a bien été recompilé
2. Vérifier les headers CORS avec curl
3. Vérifier la console du navigateur (erreurs CORS?)
4. Vérifier les logs du serveur Go
5. Hard refresh du navigateur (Cmd+Shift+R / Ctrl+Shift+F5)

## Prochaines Étapes

Maintenant que le Mux est exposé, vous pouvez ajouter vos routes personnalisées:

```go
// Exemple: Route pour obtenir des statistiques
crewAgent.Mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := map[string]interface{}{
		"uptime": time.Since(startTime).String(),
		"requests": requestCount,
	}
	json.NewEncoder(w).Encode(stats)
})

// Exemple: Route pour changer d'agent
crewAgent.Mux.HandleFunc("POST /agent/switch", func(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := crewAgent.SetSelectedAgentId(req.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
})

crewAgent.StartServer()
```

Toutes vos routes personnalisées bénéficieront automatiquement du middleware CORS! 🎉
