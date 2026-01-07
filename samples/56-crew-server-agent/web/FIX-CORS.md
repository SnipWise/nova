# Fix CORS Connection Issue

## Problème

L'interface web ne peut pas se connecter au backend car les endpoints API (sauf `/completion`) n'ont pas les headers CORS nécessaires.

## Solution 1 : Proxy CORS (Recommandé - Simple)

Utilisez le proxy CORS fourni qui ajoute les headers manquants.

### Étapes

**Terminal 1 - Backend** :
```bash
cd samples/56-crew-server-agent
go run main.go
```

**Terminal 2 - Proxy CORS** :
```bash
cd samples/56-crew-server-agent/web/proxy
go run main.go
```

Le proxy écoute sur **port 8081** et redirige vers le backend (port 8080) en ajoutant les headers CORS.

**Terminal 3 - Interface Web** :
```bash
cd samples/56-crew-server-agent/web
./start.sh
```

**Navigateur** :
```
http://localhost:3000
```

✅ **Ça marche ! Le proxy ajoute automatiquement les CORS à tous les endpoints.**

## Solution 2 : Modifier l'URL de l'API

Si vous voulez tester directement sans proxy, modifiez l'URL de l'API dans le code :

### Option A : Utiliser 127.0.0.1 au lieu de localhost

Éditez `js/api.js` :

```javascript
// Avant
const API_BASE_URL = 'http://localhost:8080';

// Après
const API_BASE_URL = 'http://127.0.0.1:8080';
```

Parfois, `127.0.0.1` vs `localhost` peut changer le comportement CORS.

### Option B : Désactiver CORS dans le Navigateur (DEV SEULEMENT)

⚠️ **Pour développement uniquement - NE PAS utiliser en production**

#### Chrome/Edge

```bash
# macOS
open -na "Google Chrome" --args --user-data-dir=/tmp/chrome-dev --disable-web-security

# Linux
google-chrome --user-data-dir=/tmp/chrome-dev --disable-web-security

# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --user-data-dir=%TEMP%\chrome-dev --disable-web-security
```

#### Firefox

1. Ouvrir `about:config`
2. Chercher `security.fileuri.strict_origin_policy`
3. Mettre à `false`

## Solution 3 : Ajouter CORS au SDK (Permanent)

Pour une solution permanente, il faut modifier le SDK Nova pour ajouter les headers CORS à tous les handlers.

### Fichiers à modifier

1. **nova-sdk/agents/serverbase/base.server.go**

Ajouter dans chaque handler :

```go
func (agent *BaseServerAgent) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Ajouter ces lignes
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Code existant
	w.Header().Set("Content-Type", "application/json")
	// ...
}
```

2. **Ou créer un middleware CORS**

Dans `nova-sdk/agents/crewserver/crew.server.agent.go`, modifier `StartServer()` :

```go
func (agent *CrewServerAgent) StartServer() error {
	mux := http.NewServeMux()

	// Routes...
	mux.HandleFunc("POST /completion", agent.handleCompletion)
	// ...

	// Wrapper CORS
	corsHandler := corsMiddleware(mux)

	agent.Log.Info("🚀 Server started on http://localhost%s", agent.Port)
	return http.ListenAndServe(agent.Port, corsHandler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

## Solution 4 : Extension Navigateur (Temporaire)

Installez une extension CORS pour le navigateur :

### Chrome/Edge
- [CORS Unblock](https://chrome.google.com/webstore/detail/cors-unblock/)
- [Allow CORS](https://chrome.google.com/webstore/detail/allow-cors-access-control/)

### Firefox
- [CORS Everywhere](https://addons.mozilla.org/en-US/firefox/addon/cors-everywhere/)

⚠️ **Attention : N'activez ces extensions qu'en développement !**

## Vérification

Pour vérifier que CORS fonctionne :

```bash
# Test avec curl
curl -v -H "Origin: http://localhost:3000" http://localhost:8081/health

# Vous devriez voir :
# < Access-Control-Allow-Origin: *
# < Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# < Access-Control-Allow-Headers: Content-Type, Authorization
```

## Recommandation

✅ **Utilisez la Solution 1 (Proxy CORS)** pour le développement :
- Simple à mettre en place
- Pas de modification du code
- Fonctionne immédiatement
- Facile à désactiver

Pour **production** :
- Utilisez la Solution 3 (modifier le SDK)
- Ajoutez une authentification
- Restreignez les origines autorisées
- Utilisez HTTPS

## Architecture avec Proxy

```
┌──────────────┐
│   Browser    │
│ localhost:   │
│   3000       │
└──────┬───────┘
       │ HTTP
       ↓
┌──────────────┐
│  Static Web  │
│   Server     │
│ (Python)     │
└──────┬───────┘
       │ Fetch API
       ↓
┌──────────────┐
│ CORS Proxy   │  ← Ajoute headers CORS
│ localhost:   │
│   8081       │
└──────┬───────┘
       │ HTTP
       ↓
┌──────────────┐
│ Nova Crew    │
│   Server     │
│ localhost:   │
│   8080       │
└──────────────┘
```

## Dépannage

### "Connection refused" sur port 8081

**Problème** : Le proxy n'est pas démarré

**Solution** : Lancez `go run cors-proxy.go`

### "502 Bad Gateway"

**Problème** : Le backend (port 8080) n'est pas démarré

**Solution** : Lancez `go run main.go` dans l'autre terminal

### L'API fonctionne mais pas l'interface

**Problème** : L'interface utilise encore le port 8080

**Solution** : L'API dans `js/api.js` devrait pointer vers `http://localhost:8081` quand vous utilisez le proxy

**Vérifiez** :
```javascript
// js/api.js
const API_BASE_URL = 'http://localhost:8081'; // Avec proxy
// OU
const API_BASE_URL = 'http://localhost:8080'; // Sans proxy (si SDK modifié)
```

## Ports Utilisés

| Service | Port | Description |
|---|---|---|
| Backend (Go) | 8080 | API Nova Crew Server |
| Proxy CORS | 8081 | Proxy avec headers CORS |
| Frontend | 3000 | Interface web statique |

---

**Question ?** Le proxy est la solution la plus simple pour le développement ! 🚀
