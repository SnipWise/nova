# Nova Crew Server - Web UI

Interface web moderne pour interagir avec le Nova Crew Server Agent.

## 🎯 Fonctionnalités

### Chat & Streaming
- ✅ **SSE Streaming** - Réponses en temps réel
- ✅ **Markdown** - Rendu formaté avec highlight.js
- ✅ **Code Highlighting** - Coloration syntaxique automatique
- ✅ **Multi-Agent** - Support de plusieurs agents (coder, thinker, cook, generic)

### Human-in-the-Loop
- ✅ **Tool Validation** - Approbation des appels de fonction
- ✅ **Operation Cancel** - Annulation des opérations en attente
- ✅ **Real-time Notifications** - Alertes visuelles pour les tools

### Gestion de la Mémoire
- ✅ **Context Size** - Suivi de la taille du contexte
- ✅ **Clear Memory** - Réinitialisation de la conversation
- ✅ **View Messages** - Historique complet des messages
- ✅ **Export** - Export JSON de la conversation

### Interface
- ✅ **Design Moderne** - Dark theme, responsive
- ✅ **Modal System** - Confirmations élégantes
- ✅ **Auto-scroll** - Suit automatiquement la conversation
- ✅ **Loading States** - États visuels clairs

## 🏗️ Architecture v4 (Actuelle)

### Direct Connection

```
┌─────────────────────────────────────┐
│  Browser (localhost:3000)           │
│  - Vue.js 3 (CDN)                  │
│  - SSE Client                       │
└─────────────────┬───────────────────┘
                  │
                  │ HTTP/SSE
                  │ Port 8080
                  ↓
┌─────────────────────────────────────┐
│  Nova Crew Server                   │
│  - CORS Middleware ✅               │
│  - Multiple Chat Agents             │
│  - Tools Agent (optional)           │
│  - RAG Agent (optional)             │
│  - Compressor Agent (optional)      │
└─────────────────────────────────────┘
```

**Nouveautés v4:**
- ❌ Proxy CORS supprimé
- ✅ CORS intégré au SDK
- ✅ Connexion directe au serveur
- ✅ Routes personnalisées via `agent.Mux`

## 🚀 Quick Start

### Prérequis

- Go 1.22+
- Docker Desktop (avec Agentic Compose)
- Navigateur moderne (Chrome, Firefox, Safari)

### 1. Démarrer le Serveur Nova

```bash
cd samples/56-crew-server-agent
go run main.go
```

Vérifier les logs:
```
🚀 Server started on http://localhost:8080
```

### 2. Ouvrir le Navigateur

**Option A: Serveur Web Local (Recommandé)**
```bash
# Python
cd web
python3 -m http.server 3000

# OU Node.js
npx serve -p 3000
```

Puis ouvrir: http://localhost:3000

**Option B: Direct (File Protocol)**
```bash
open web/index.html
```

⚠️ Note: Certaines fonctionnalités peuvent être limitées en file://

### 3. Tester la Connexion

1. Ouvrir DevTools (F12)
2. Onglet Console
3. Vérifier: Pas d'erreurs CORS
4. Envoyer un message de test

## 📁 Structure des Fichiers

```
web/
├── index.html                          # Page principale
├── js/
│   ├── api.js                         # Client API (SSE, fetch)
│   ├── markdown.js                    # Rendering markdown
│   ├── app.js                         # Vue.js app principale
│   └── components/
│       ├── ChatMessage.js             # Composant message
│       ├── StatusBar.js               # Barre de statut
│       ├── InputBar.js                # Zone de saisie
│       ├── OperationControls.js       # Notifications tools
│       └── Modal.js                   # Système de modals
├── docs/
│   ├── MIGRATION-TO-DIRECT-CONNECTION.md
│   ├── CUSTOM-ROUTES-EXAMPLES.md
│   ├── REMOVING-CORS-PROXY.md
│   ├── CACHE-BUSTING.md
│   └── CHANGELOG-v4.md
├── testing/
│   └── test-full-validation-cycle.sh   # Tests curl
└── README.md                           # Ce fichier
```

## 💬 Utilisation

### Envoyer des Messages

1. Taper votre message dans la zone de saisie
2. Appuyer sur **Enter** pour envoyer (ou cliquer "Send")
3. Utiliser **Shift+Enter** pour nouvelle ligne
4. Regarder les réponses streamer en temps réel

### Routing d'Agents

Le système route automatiquement vers des agents spécialisés:

- **Coder Agent**: Programmation, code, debugging
- **Thinker Agent**: Philosophie, math, science, psychologie
- **Cook Agent**: Cuisine, recettes, nourriture
- **Generic Agent**: Tout le reste

### Validation de Tools

Quand l'agent veut appeler un tool:

1. Une notification apparaît avec les détails
2. Cliquer **Validate** pour approuver
3. Cliquer **Cancel** pour rejeter
4. L'agent procède selon votre choix

### Boutons d'Action

- **📤 Send**: Envoyer le message
- **⏹ Stop**: Arrêter le streaming
- **🗑 Clear Memory**: Réinitialiser la conversation
- **💬 View Messages**: Afficher l'historique
- **🤖 View Models**: Informations sur les modèles
- **🔄 Reset Operations**: Vider les opérations en attente

### Barre de Statut

Informations en temps réel:

- **Agent**: Agent actuellement actif
- **Context Size**: Taille du contexte
- **Chat Model**: Modèle utilisé pour le chat
- **Tools**: Modèle pour function calling
- **RAG**: Modèle d'embeddings

## 🔌 API Endpoints

### Endpoints Standards (Nova SDK)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/completion` | POST | Streaming chat completion (SSE) |
| `/completion/stop` | POST | Arrêter le streaming |
| `/models` | GET | Info sur les modèles |
| `/memory/reset` | POST | Réinitialiser la mémoire |
| `/memory/messages/list` | GET | Liste des messages |
| `/memory/messages/context-size` | GET | Taille du contexte |
| `/operation/validate` | POST | Valider un tool call |
| `/operation/cancel` | POST | Annuler un tool call |
| `/operation/reset` | POST | Reset toutes les operations |
| `/health` | GET | Health check |

### Endpoints Personnalisés

Vous pouvez ajouter vos propres routes! Voir [CUSTOM-ROUTES-EXAMPLES.md](docs/CUSTOM-ROUTES-EXAMPLES.md).

## 🔧 Configuration

### URL de l'API

Par défaut: `http://localhost:8080`

Pour modifier:

```javascript
// web/js/api.js (ligne 7)
const API_BASE_URL = 'http://your-server:8080';
```

### Cache Busting

Après modification des fichiers JS:

```html
<!-- web/index.html -->
<script src="js/api.js?v=5"></script>  <!-- Incrémenter la version -->
```

Voir [CACHE-BUSTING.md](docs/CACHE-BUSTING.md) pour plus d'infos.

### Personnalisation CSS

Toutes les styles sont dans `index.html` dans le tag `<style>`.

**Couleurs principales:**
- Background: `#1a1a1a`
- Cards: `#2d2d2d`
- Primary: `#4fc3f7` (blue)
- Success: `#43a047` (green)
- Danger: `#e53935` (red)

## 🧪 Tests

### Tests Manuels

1. **Chat basique**
   - Envoyer: "Hello"
   - Vérifier: Réponse streaming

2. **Tool Validation**
   - Envoyer: "Calculate 5 + 3"
   - Vérifier: Popup apparaît
   - Cliquer: Validate
   - Vérifier: Résultat "8"

3. **Modals**
   - Cliquer: "View Models"
   - Vérifier: Modal s'ouvre

### Tests Automatisés

```bash
cd web/testing
./test-full-validation-cycle.sh
```

### Tests CORS

```bash
# Test preflight
curl -I -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  http://localhost:8080/models

# Test GET
curl -H "Origin: http://localhost:3000" \
  http://localhost:8080/health
```

## 📚 Documentation

- [**MIGRATION-TO-DIRECT-CONNECTION.md**](docs/MIGRATION-TO-DIRECT-CONNECTION.md) - Guide migration v3 → v4
- [**CUSTOM-ROUTES-EXAMPLES.md**](docs/CUSTOM-ROUTES-EXAMPLES.md) - Exemples routes personnalisées
- [**REMOVING-CORS-PROXY.md**](docs/REMOVING-CORS-PROXY.md) - Suppression du proxy
- [**CACHE-BUSTING.md**](docs/CACHE-BUSTING.md) - Gestion cache navigateur
- [**CHANGELOG-v4.md**](docs/CHANGELOG-v4.md) - Détails changements v4

## 🐛 Dépannage

### Erreur CORS

**Symptômes:**
```
Access to fetch at 'http://localhost:8080/completion' blocked by CORS policy
```

**Solutions:**
1. Vérifier serveur démarre sur port 8080
2. Vérifier `API_BASE_URL` pointe vers 8080
3. Hard refresh (Cmd+Shift+R)
4. Vérifier headers CORS avec curl

### UI Ne Se Met Pas à Jour

**Symptômes:**
- Changements JS non visibles

**Solutions:**
1. Incrémenter version (?v=4 → ?v=5)
2. Hard refresh
3. Vider cache navigateur
4. DevTools → Network → Disable cache

### Tool Validation Ne Fonctionne Pas

**Symptômes:**
- Pas de popup de validation

**Solutions:**
1. Vérifier logs backend
2. Console JS pour erreurs
3. Tester avec curl (voir testing/)

### Stream Bloqué

**Symptômes:**
- Loading infini

**Solutions:**
1. Cliquer Stop
2. Refresh page
3. Vérifier logs backend

## 🎨 Composants Vue.js

### ChatMessage

Affiche un message avec markdown et code highlighting.

### StatusBar

Barre de statut avec infos agent/context/modèles.

### InputBar

Zone de saisie avec tous les boutons d'action.

### OperationControls

Notifications pour validation des tools.

### Modal

Système de modals réutilisable (info/confirm/list).

## 🔐 Sécurité

### CORS Production

Restreindre les origines:

```go
// Modifier crew.server.agent.go
allowedOrigins := map[string]bool{
    "https://app.example.com": true,
}
```

### HTTPS

En production:

```go
http.ListenAndServeTLS(":443", "cert.pem", "key.pem", handler)
```

## 📈 Performance

- **Bundle Size**: ~250KB (CDN)
- **Initial Load**: < 1s
- **Streaming**: Real-time (SSE)
- **Memory**: Efficient Vue.js 3

## 🌐 Compatibilité Navigateur

- Chrome/Edge: ✅
- Firefox: ✅
- Safari: ✅
- Mobile: ✅ Responsive

## 📝 Changelog

### v4.0.0 (2026-01-07)
- ✅ CORS middleware SDK
- ✅ Suppression proxy
- ✅ Routes personnalisées (Mux)
- ✅ Documentation complète

### v3.0.0
- ✅ Système modals
- ✅ Cache busting
- ✅ UI improvements

### v2.0.0
- ✅ SSE streaming fixes
- ✅ Tool validation
- ✅ CORS proxy

### v1.0.0
- ✅ Interface Vue.js 3
- ✅ Chat streaming
- ✅ Markdown rendering

## 🚀 Fonctionnalités Futures

- [ ] Dark/Light theme toggle
- [ ] Message export (JSON, MD)
- [ ] Multi-session support
- [ ] Voice input
- [ ] Copy code blocks
- [ ] Message search
- [ ] File upload
- [ ] Custom system instructions

## 🚢 Déploiement

Voir le guide de déploiement Docker dans la documentation du SDK Nova.

## 🤝 Contribution

1. Créer nouveau composant dans `js/components/`
2. Importer dans `index.html`
3. Utiliser dans `app.js`
4. Incrémenter cache version
5. Documenter

## 📞 Support

- **Issues**: GitHub Issues
- **Docs**: [/docs](./docs/)
- **Examples**: [CUSTOM-ROUTES-EXAMPLES.md](./docs/CUSTOM-ROUTES-EXAMPLES.md)

## 📄 License

Voir LICENSE dans le répertoire racine du projet Nova.

---

**Made with ❤️ for Nova SDK**
