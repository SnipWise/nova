# 📚 Nova Crew Server - Web Interface Documentation Index

Bienvenue dans la documentation de l'interface web pour Nova Crew Server Agent !

## 🚀 Démarrage Rapide

**Nouveau ici ?** Commencez par ces fichiers dans l'ordre :

1. **[QUICKSTART.md](QUICKSTART.md)** - Démarrage en 3 étapes (5 minutes)
2. **[demo-questions.md](demo-questions.md)** - Questions d'exemple pour tester
3. **[CHECKLIST.md](CHECKLIST.md)** - Vérifier que tout fonctionne

## 📖 Documentation Complète

### Guide Utilisateur

- **[README.md](README.md)** - Documentation complète de l'interface
  - Fonctionnalités
  - Installation
  - Utilisation
  - API
  - Dépannage

### Guide Développeur

- **[PROJECT-STRUCTURE.md](PROJECT-STRUCTURE.md)** - Architecture technique détaillée
  - Structure des fichiers
  - Architecture des composants
  - Flux de données
  - Gestion d'état
  - Personnalisation
  - Performance

### Guides Pratiques

- **[demo-questions.md](demo-questions.md)** - Exemples de questions pour chaque agent
  - Coder Agent
  - Thinker Agent
  - Cook Agent
  - Function Calling
  - RAG
  - Tests markdown

- **[CHECKLIST.md](CHECKLIST.md)** - Liste de vérification complète
  - Fichiers requis
  - Tests fonctionnels
  - Dépannage
  - Métriques de succès

- **[TESTING.md](TESTING.md)** - Guide de test complet
  - Tests API (scripts curl)
  - Tests interface web (navigateur)
  - Checklist de tests
  - Guide de dépannage

- **[FIX-CORS.md](FIX-CORS.md)** - Solution au problème CORS
  - Explication du problème
  - Utilisation du proxy CORS
  - Solutions alternatives

## 🛠 Fichiers Techniques

### Code Source

- **[index.html](index.html)** - Point d'entrée HTML + CSS
- **[js/api.js](js/api.js)** - Service API (SSE streaming)
- **[js/markdown.js](js/markdown.js)** - Rendu markdown + highlighting
- **[js/app.js](js/app.js)** - Application Vue.js principale

### Composants Vue.js

- **[js/components/ChatMessage.js](js/components/ChatMessage.js)** - Affichage des messages
- **[js/components/InputBar.js](js/components/InputBar.js)** - Zone de saisie et boutons
- **[js/components/StatusBar.js](js/components/StatusBar.js)** - Infos contexte/modèles
- **[js/components/OperationControls.js](js/components/OperationControls.js)** - Validation d'outils

### Scripts de Lancement

- **[start.sh](start.sh)** - Lancement macOS/Linux
- **[start.bat](start.bat)** - Lancement Windows

## 📊 Vue d'Ensemble du Système

```
┌─────────────────────────────────────────────────┐
│          Web Browser (Vue.js 3 SPA)             │
│  ┌───────────────────────────────────────────┐  │
│  │  index.html (UI + CSS)                    │  │
│  │  ├── js/app.js (Main App)                 │  │
│  │  │   ├── ChatMessage Component            │  │
│  │  │   ├── InputBar Component               │  │
│  │  │   ├── StatusBar Component              │  │
│  │  │   └── OperationControls Component      │  │
│  │  ├── js/api.js (API Layer)                │  │
│  │  └── js/markdown.js (Rendering)           │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────┬───────────────────────────┘
                      │ HTTP + SSE
                      ↓
┌─────────────────────────────────────────────────┐
│     Nova Crew Server (Go) - Port 8080           │
│  ┌───────────────────────────────────────────┐  │
│  │  REST API + SSE Streaming                 │  │
│  │  ├── POST /completion (SSE)               │  │
│  │  ├── POST /memory/reset                   │  │
│  │  ├── GET /memory/messages/list            │  │
│  │  ├── POST /operation/validate             │  │
│  │  └── GET /models                          │  │
│  └───────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────┐  │
│  │  Agent Crew                               │  │
│  │  ├── Coder Agent (programming)            │  │
│  │  ├── Thinker Agent (philosophy/science)   │  │
│  │  ├── Cook Agent (culinary)                │  │
│  │  └── Generic Agent (default)              │  │
│  └───────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────┐  │
│  │  Specialized Agents                       │  │
│  │  ├── Tools Agent (function calling)       │  │
│  │  ├── RAG Agent (document retrieval)       │  │
│  │  ├── Compressor Agent (context compress)  │  │
│  │  └── Orchestrator Agent (topic routing)   │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## 🎯 Cas d'Usage Principaux

### 1. Chat Simple
```
User → InputBar → API → /completion → Agent → SSE Stream → ChatMessage
```

### 2. Function Calling
```
User → Question → Tools Agent → Notification → OperationControls → Validate → Execute
```

### 3. Agent Routing
```
User → Question → Orchestrator → Topic Detection → Match Function → Switch Agent
```

### 4. Context Management
```
Messages → Context Size Poll → StatusBar Display
User → Clear Memory → API → Reset → Empty Messages
```

## 🔑 Concepts Clés

### SSE (Server-Sent Events)
Streaming unidirectionnel serveur → client pour les réponses en temps réel.

### Markdown + Syntax Highlighting
Conversion markdown → HTML avec coloration de code via Highlight.js.

### Vue.js Composition API
Gestion d'état réactive moderne sans Vuex/Pinia.

### Multi-Agent Orchestration
Routage automatique vers l'agent spécialisé approprié.

### Human-in-the-Loop
Validation manuelle des appels d'outils critiques.

## 🚦 Workflow de Développement

```bash
# 1. Démarrer le serveur Go
cd samples/56-crew-server-agent
go run main.go

# 2. Démarrer le serveur web (nouveau terminal)
cd web
./start.sh

# 3. Ouvrir navigateur
open http://localhost:3000

# 4. Développer
# Éditez les fichiers .js ou .html
# Rafraîchissez le navigateur (pas de build)

# 5. Déboguer
# Ouvrez DevTools (F12)
# Console pour logs
# Network pour API calls
```

## 📈 Ordre de Lecture Recommandé

### Pour les Utilisateurs
1. QUICKSTART.md
2. demo-questions.md
3. CHECKLIST.md
4. README.md (référence)

### Pour les Développeurs
1. PROJECT-STRUCTURE.md
2. README.md (section API)
3. Code source (js/*.js)
4. Composants (js/components/*.js)

### Pour le Déploiement
1. CHECKLIST.md
2. README.md (section Security)
3. Code source (modifications CORS/auth)

## 🆘 Besoin d'Aide ?

| Problème | Voir |
|---|---|
| Installation | QUICKSTART.md |
| Bugs/Erreurs | CHECKLIST.md → Dépannage |
| Fonctionnalités | README.md |
| Architecture | PROJECT-STRUCTURE.md |
| Exemples | demo-questions.md |

## 📦 Fichiers par Catégorie

### 📖 Documentation (Vous êtes ici)
- INDEX.md (ce fichier)
- README.md
- QUICKSTART.md
- PROJECT-STRUCTURE.md
- CHECKLIST.md
- demo-questions.md
- specs.txt

### 💻 Code Source
- index.html
- js/api.js
- js/markdown.js
- js/app.js
- js/components/*.js

### 🔧 Utilitaires
- start.sh
- start.bat

## 🎓 Ressources Externes

### Technologies Utilisées
- [Vue.js 3 Documentation](https://vuejs.org/)
- [Marked.js (Markdown)](https://marked.js.org/)
- [Highlight.js (Syntax)](https://highlightjs.org/)
- [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)

### Nova SDK
- [Nova SDK Repository](../../README.md)
- [Crew Server Documentation](../../nova-sdk/agents/crewserver/README.fr.md)

## 🏁 Prêt à Commencer ?

### Option 1 : Démarrage Rapide (Recommandé)
```bash
# Lire QUICKSTART.md
cat QUICKSTART.md

# Lancer
./start.sh
```

### Option 2 : Lecture Approfondie
```bash
# Lire toute la documentation
cat README.md
cat PROJECT-STRUCTURE.md
```

### Option 3 : Plonger dans le Code
```bash
# Explorer les composants
ls -la js/components/
cat js/app.js
```

---

**Bon développement ! 🚀**

Si vous avez des questions, consultez la documentation appropriée ci-dessus.
