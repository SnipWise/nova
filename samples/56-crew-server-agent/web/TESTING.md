# Testing Guide

Guide complet pour tester l'API et l'interface web.

## 🎯 Deux Types de Tests

### 1. Tests API (curl)
Scripts bash pour tester l'API directement avec curl.
📁 Dossier : [`testing/`](testing/)

### 2. Tests Interface Web (navigateur)
Tests manuels de l'interface Vue.js dans le navigateur.

---

## 🧪 Tests API (Scripts curl)

### Démarrage Rapide

```bash
# 1. Démarrer tous les services
cd samples/56-crew-server-agent/web
./start-all.sh

# 2. Dans un autre terminal, lancer les tests
cd samples/56-crew-server-agent/web/testing
./run-all-tests.sh
```

### Tests Individuels

```bash
cd samples/56-crew-server-agent/web/testing

# Test de santé
./test-health.sh

# Test des modèles
./test-models.sh

# Test de streaming
./test-stream.sh

# Test avec question personnalisée
./test-stream.sh "Explain what is React in one sentence"
```

### Voir Documentation Complète
📖 [testing/README.md](testing/README.md)

---

## 🌐 Tests Interface Web (Navigateur)

### Prérequis

1. **Démarrer tous les services** :
   ```bash
   cd samples/56-crew-server-agent/web
   ./start-all.sh
   ```

2. **Ouvrir le navigateur** :
   ```
   http://localhost:3000
   ```

3. **Ouvrir DevTools** (F12) :
   - Console : Voir les logs
   - Network : Voir les requêtes API

### Checklist de Tests

#### ✅ Connexion Backend

1. La page charge sans erreur
2. La barre de statut affiche :
   - Agent actif
   - Context size
   - Modèles

**Erreur commune** : "Failed to connect to server"
- Vérifier que `start-all.sh` est lancé
- Vérifier que le proxy CORS écoute sur port 8081

#### ✅ Chat Simple

1. Taper : "Hello, how are you?"
2. Appuyer sur Enter
3. Vérifier :
   - Message utilisateur apparaît (droite, bleu)
   - Message assistant apparaît (gauche, gris)
   - Le texte s'affiche progressivement (streaming)

#### ✅ Markdown & Code

1. Taper : "Write a Python hello world"
2. Vérifier :
   - Code est dans un bloc coloré
   - Syntaxe Python est highlightée
   - Les couleurs sont visibles

#### ✅ Routage Multi-Agents

Tester chaque agent :

**Coder Agent** :
```
Write a Go function that reverses a string
```
→ Vérifier que "Agent: coder" s'affiche dans la barre de statut

**Thinker Agent** :
```
What is the nature of consciousness?
```
→ Vérifier que "Agent: thinker" s'affiche

**Cook Agent** :
```
Give me a recipe for chocolate cake
```
→ Vérifier que "Agent: cook" s'affiche

**Generic Agent** :
```
What is the capital of France?
```
→ Vérifier que "Agent: generic" s'affiche

#### ✅ Function Calling

1. Taper : "Say hello to Alice"
2. Vérifier :
   - Une notification d'opération apparaît
   - Elle contient "Validate" et "Cancel"
   - L'operation_id est affiché
3. Cliquer sur "Validate"
4. Vérifier :
   - La réponse contient "Hello, Alice!"

**Test 2** :
```
Calculate the sum of 42 and 58
```
→ Résultat devrait être 100

#### ✅ Contrôles de Mémoire

**Clear Memory** :
1. Avoir quelques messages dans l'historique
2. Cliquer "Clear Memory"
3. Confirmer
4. Vérifier :
   - Tous les messages disparaissent
   - Context size retombe à une petite valeur

**View Messages** :
1. Avoir quelques messages
2. Cliquer "View Messages"
3. Vérifier :
   - Console DevTools affiche tous les messages
   - Format JSON correct

**View Models** :
1. Cliquer "View Models"
2. Vérifier :
   - Alert affiche les modèles
   - Chat model, Tools model, RAG model présents

#### ✅ Stop Streaming

1. Taper une longue question :
   ```
   Explain the entire history of computer science from the 1800s to today
   ```
2. Pendant que ça stream, cliquer "Stop"
3. Vérifier :
   - Le streaming s'arrête immédiatement
   - Le message partiel reste affiché

#### ✅ Context Size

1. Envoyer plusieurs messages
2. Observer la barre de statut
3. Vérifier :
   - "Context Size" augmente
   - Le nombre change toutes les ~2 secondes

#### ✅ Responsive Design

Tester sur différentes tailles :

**Desktop** (> 1024px) :
- Messages prennent ~85% de largeur
- Boutons alignés en ligne

**Tablet** (768px - 1024px) :
- Messages plus larges
- Boutons réorganisés

**Mobile** (< 768px) :
- Messages prennent 95% de largeur
- Boutons empilés verticalement
- Zone de saisie adaptée

#### ✅ Auto-Scroll

1. Avoir plusieurs messages (remplir l'écran)
2. Envoyer un nouveau message
3. Vérifier :
   - La page scrolle automatiquement vers le bas
   - Le nouveau message est visible

### Tests DevTools

#### Console Tab

Vérifier qu'il n'y a **PAS** :
- ❌ Erreurs JavaScript (rouge)
- ❌ Erreurs de chargement de ressources
- ❌ Erreurs CORS

Devrait avoir :
- ✅ Logs d'API calls
- ✅ Messages SSE loggés

#### Network Tab

1. Filtrer par "Fetch/XHR"
2. Envoyer un message
3. Vérifier :
   - Request à `/completion` (status 200)
   - Type "EventStream"
   - Headers CORS présents

**Exemple de headers attendus** :
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Content-Type: text/event-stream
```

#### Application Tab (Storage)

L'application **NE stocke PAS** de données localement :
- Pas de localStorage
- Pas de sessionStorage
- Pas de cookies

→ Refresh de la page = perte de l'historique (comportement attendu)

---

## 🐛 Guide de Dépannage

### Problème : Page Blanche

**Diagnostic** :
1. Ouvrir Console (F12)
2. Chercher erreurs JavaScript

**Solutions** :
- Vérifier que tous les fichiers JS sont chargés
- Vérifier la connexion Internet (CDN)
- Vider le cache (Ctrl+Shift+R)

### Problème : "Failed to connect to server"

**Diagnostic** :
```bash
curl http://localhost:8081/health
```

**Solutions** :
1. Proxy CORS non démarré :
   ```bash
   cd samples/56-crew-server-agent/web/proxy
   go run main.go
   ```

2. Backend non démarré :
   ```bash
   cd samples/56-crew-server-agent
   go run main.go
   ```

### Problème : Pas de Streaming

**Diagnostic** :
- Network tab → Vérifier EventStream
- Console → Chercher erreurs

**Solutions** :
- Vérifier que le backend répond
- Tester avec curl :
  ```bash
  cd web/testing
  ./test-stream.sh
  ```

### Problème : Code Non Coloré

**Diagnostic** :
- Console → Chercher "highlight.js"

**Solutions** :
- Vérifier connexion Internet
- Vérifier que Highlight.js est chargé
- Spécifier le langage : \`\`\`python

### Problème : Markdown Non Rendu

**Diagnostic** :
- Console → Chercher "marked.js"

**Solutions** :
- Vérifier connexion Internet
- Vérifier que Marked.js est chargé

---

## 📊 Métriques de Succès

L'application fonctionne correctement si :

- ✅ Tous les tests API passent (`run-all-tests.sh`)
- ✅ La page charge en < 2 secondes
- ✅ Premier token reçu en < 1 seconde
- ✅ Streaming fluide sans saccades
- ✅ Markdown s'affiche correctement
- ✅ Code est coloré automatiquement
- ✅ Tous les boutons fonctionnent
- ✅ Aucune erreur dans la console
- ✅ Responsive fonctionne sur mobile

---

## 🎓 Tests Avancés

### Test de Charge

Envoyer plusieurs messages rapidement pour tester :
- Gestion de la file d'attente
- Stabilité du streaming
- Gestion de la mémoire

### Test de Contexte Long

Envoyer beaucoup de messages pour tester :
- Compression du contexte (à ~8500 tokens)
- Performance avec historique long
- Gestion de la mémoire browser

### Test de Validation Multiple

Envoyer plusieurs commandes d'outils sans valider pour tester :
- File d'attente d'opérations
- Affichage de plusieurs notifications
- Reset des opérations

---

## 📚 Ressources

- [API Documentation](../nova-sdk/agents/crewserver/README.fr.md)
- [Scripts de Test curl](testing/README.md)
- [Guide CORS](FIX-CORS.md)
- [Questions de Démo](demo-questions.md)

---

**Bon test ! 🧪**
