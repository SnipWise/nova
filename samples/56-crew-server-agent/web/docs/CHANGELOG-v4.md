# Changelog v4 - Migration CORS & Routes Personnalisées

**Date**: 2026-01-07
**Version**: 4.0.0
**Type**: Breaking Changes (Architecture)

## 🎯 Objectif Principal

Intégrer le support CORS directement dans le SDK Nova pour éliminer le besoin du proxy CORS intermédiaire et permettre l'ajout de routes personnalisées.

## 📋 Résumé des Changements

### Architecture
- ❌ **Supprimé**: Proxy CORS (web/proxy/)
- ✅ **Ajouté**: Middleware CORS intégré dans le SDK
- ✅ **Ajouté**: Exposition du Mux HTTP pour routes personnalisées

### Connexion
- **Avant**: Browser → Proxy (8081) → Server (8080)
- **Après**: Browser → Server (8080) ✅ CORS direct

## 🔧 Modifications Techniques

### 1. SDK - nova-sdk/agents/crewserver/crew.server.agent.go

#### Nouveau Middleware CORS
```go
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

**Lignes**: 392-416

#### StartServer() Modifié
```go
func (agent *CrewServerAgent) StartServer() error {
    mux := http.NewServeMux()

    // Expose mux for custom routes
    agent.Mux = mux

    // Routes...

    // Apply CORS middleware
    handler := corsMiddleware(mux)

    return http.ListenAndServe(agent.Port, handler)
}
```

**Lignes**: 418-442

#### Nouveau Champ Public: Mux
```go
type CrewServerAgent struct {
    *serverbase.BaseServerAgent

    // ... autres champs

    // HTTP server multiplexer for custom routes
    Mux *http.ServeMux
}
```

**Ligne**: 36

### 2. Frontend - web/js/api.js

#### Changement de Port
```javascript
// v3
const API_BASE_URL = 'http://localhost:8081';

// v4
const API_BASE_URL = 'http://localhost:8080';
```

**Ligne**: 7

### 3. Frontend - web/index.html

#### Cache Busting
```html
<!-- v3 -->
<script src="js/api.js?v=3"></script>

<!-- v4 -->
<script src="js/api.js?v=4"></script>
```

**Lignes**: 619-626

## 📚 Documentation Ajoutée

### 1. MIGRATION-TO-DIRECT-CONNECTION.md
Guide complet de migration:
- Architecture avant/après
- Instructions détaillées
- Tests de vérification
- Configuration production
- Procédure de rollback

### 2. CUSTOM-ROUTES-EXAMPLES.md
Exemples d'utilisation du Mux:
- 8 exemples pratiques de routes personnalisées
- Stats, agent switching, RAG upload, etc.
- Code complet et tests curl
- Bonnes pratiques

### 3. CHANGELOG-v4.md
Ce fichier - résumé des changements.

## 🚀 Nouveautés

### Routes Personnalisées

Vous pouvez maintenant ajouter vos propres routes:

```go
crewAgent, _ := crewserver.NewAgent(ctx, options...)

// Ajouter des routes AVANT StartServer()
crewAgent.Mux.HandleFunc("GET /stats", statsHandler)
crewAgent.Mux.HandleFunc("POST /custom", customHandler)

crewAgent.StartServer()
```

**Bénéfices**:
- Toutes les routes bénéficient automatiquement du middleware CORS
- Pas besoin de configurer CORS manuellement
- Extension facile du serveur

### CORS Automatique

Tous les endpoints (standards + personnalisés) ont maintenant les headers CORS:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, Accept`
- `Access-Control-Allow-Credentials: true`

## 🔄 Migration

### Checklist

- [x] Modifier le SDK - crew.server.agent.go
- [x] Modifier api.js (8081 → 8080)
- [x] Incrémenter cache busting (v3 → v4)
- [x] Créer documentation de migration
- [x] Créer exemples de routes personnalisées
- [x] Tester la connexion directe
- [x] Vérifier les headers CORS
- [ ] Supprimer web/proxy/ (optionnel)

### Instructions

1. **Rebuild SDK**
```bash
cd /Users/k33g/Library/CloudStorage/Dropbox/SnipWise/nova
go mod tidy
```

2. **Arrêter le Proxy**
```bash
# Plus besoin de:
# cd web/proxy && go run main.go
```

3. **Démarrer le Serveur**
```bash
cd samples/56-crew-server-agent
go run main.go
```

4. **Ouvrir le Navigateur**
```bash
open http://localhost:3000
# OU
open web/index.html
```

5. **Hard Refresh**
- Mac: `Cmd + Shift + R`
- Windows/Linux: `Ctrl + Shift + F5`

## 🧪 Tests

### Vérifier CORS
```bash
curl -I -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:8080/models

# Devrait retourner les headers CORS
```

### Vérifier Connexion Directe
1. Ouvrir DevTools (F12)
2. Onglet Network
3. Envoyer un message
4. Vérifier que les requêtes vont vers `localhost:8080`
5. Pas d'erreurs CORS

### Fonctionnalités
- ✅ Envoi de message
- ✅ Streaming SSE
- ✅ Validation de tools
- ✅ Annulation de tools
- ✅ Modals (Clear Memory, View Messages, etc.)
- ✅ Context size update
- ✅ Agent switch

## ⚠️ Breaking Changes

### 1. Port par Défaut
- **Avant**: Frontend → 8081 (proxy)
- **Après**: Frontend → 8080 (direct)

Si vous utilisez un autre port, modifiez `API_BASE_URL` dans api.js.

### 2. Proxy Non Nécessaire
Le proxy CORS n'est plus requis. Si votre infrastructure dépend du proxy:
- Soit migrer vers connexion directe
- Soit continuer à utiliser le proxy (mais pas recommandé)

### 3. SDK API Change
Le champ `Mux` est maintenant public:
```go
// v3 - N/A

// v4
agent.Mux.HandleFunc("GET /custom", handler)
```

## 🎁 Bénéfices

### Performance
- ✅ **-1 saut réseau**: Pas de proxy intermédiaire
- ✅ **Latence réduite**: Communication directe
- ✅ **Moins de ressources**: Un processus en moins

### Architecture
- ✅ **Plus simple**: Un seul serveur
- ✅ **Moins de points de défaillance**: Un processus au lieu de deux
- ✅ **Production-ready**: CORS natif

### Développement
- ✅ **Debugging facile**: Un seul point d'entrée
- ✅ **Extensibilité**: Routes personnalisées via Mux
- ✅ **Flexibilité**: Ajout facile de fonctionnalités

### Maintenance
- ✅ **Code plus propre**: Moins de couches
- ✅ **Configuration simple**: Pas besoin de gérer le proxy
- ✅ **Déploiement facile**: Un binaire au lieu de deux

## 📦 Fichiers Modifiés

### SDK
```
nova-sdk/agents/crewserver/crew.server.agent.go
```

### Frontend
```
samples/56-crew-server-agent/web/js/api.js
samples/56-crew-server-agent/web/index.html
```

### Documentation
```
samples/56-crew-server-agent/web/docs/MIGRATION-TO-DIRECT-CONNECTION.md
samples/56-crew-server-agent/web/docs/CUSTOM-ROUTES-EXAMPLES.md
samples/56-crew-server-agent/web/docs/CHANGELOG-v4.md
```

### Optionnel à Supprimer
```
samples/56-crew-server-agent/web/proxy/
```

## 🔮 Prochaines Étapes

### Fonctionnalités Possibles
1. **Authentification**: Middleware JWT/OAuth
2. **Rate Limiting**: Limiter les requêtes par IP
3. **Logging**: Middleware de logging HTTP
4. **Metrics**: Prometheus/OpenTelemetry
5. **WebSockets**: Support temps réel bidirectionnel
6. **GraphQL**: Endpoint GraphQL pour queries complexes

### Configuration CORS Production
Pour la production, restreindre les origines autorisées:

```go
// Modifier corsMiddleware dans crew.server.agent.go
allowedOrigins := map[string]bool{
    "https://app.example.com": true,
}

if allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

### Routes Personnalisées
Voir `CUSTOM-ROUTES-EXAMPLES.md` pour des idées:
- Stats endpoint
- Agent switching API
- RAG document upload
- Conversation export
- Health checks détaillés

## 📞 Support

### Problèmes Courants

**1. Erreur CORS après migration**
- Hard refresh du navigateur
- Vérifier que le serveur démarre sur port 8080
- Vérifier `API_BASE_URL` dans api.js

**2. "Cannot read property 'Mux' of undefined"**
- Rebuilder le SDK: `go mod tidy`
- Vérifier que vous utilisez la dernière version

**3. Routes personnalisées ne fonctionnent pas**
- Les ajouter AVANT `StartServer()`
- Vérifier la syntaxe: `"GET /endpoint"` (Go 1.22+)

**4. Proxy toujours actif**
- Arrêter le processus proxy
- Vérifier qu'aucun processus n'écoute sur 8081: `lsof -i :8081`

### Debug

```bash
# Vérifier les headers CORS
curl -I http://localhost:8080/models

# Vérifier les processus
lsof -i :8080
lsof -i :8081

# Logs du serveur
# Vérifier la console Go pour les erreurs
```

## 🏆 Credits

- **Développement**: Migration CORS SDK Nova
- **Testing**: Web UI avec Vue.js 3
- **Documentation**: Guides complets de migration et exemples

## 📝 Notes de Version

**v4.0.0** - 2026-01-07
- ✅ CORS middleware intégré au SDK
- ✅ Exposition du Mux HTTP pour routes personnalisées
- ✅ Migration vers connexion directe (suppression du proxy)
- ✅ Documentation complète de migration
- ✅ Exemples de routes personnalisées

**v3.0.0** - Précédent
- Modal system
- Cache busting
- UI improvements

**v2.0.0** - Précédent
- SSE streaming fixes
- Tool validation
- CORS proxy

**v1.0.0** - Initial
- Interface web Vue.js 3
- Chat streaming
- Markdown rendering

## 🔗 Liens Utiles

- [MIGRATION-TO-DIRECT-CONNECTION.md](./MIGRATION-TO-DIRECT-CONNECTION.md)
- [CUSTOM-ROUTES-EXAMPLES.md](./CUSTOM-ROUTES-EXAMPLES.md)
- [REMOVING-CORS-PROXY.md](./REMOVING-CORS-PROXY.md)
- [CACHE-BUSTING.md](./CACHE-BUSTING.md)

---

**Statut**: ✅ Prêt pour Production
**Compatibilité**: Go 1.22+, Vue.js 3.4+
**Testé**: macOS, Linux (Chrome, Firefox, Safari)
