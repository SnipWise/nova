# Pre-Launch Checklist

Vérifiez ces points avant de lancer l'interface web.

## ✅ Fichiers du Projet

Vérifiez que tous les fichiers sont présents :

```bash
cd samples/56-crew-server-agent/web
ls -la
```

Vous devriez voir :
- [ ] `index.html` - Point d'entrée principal
- [ ] `js/api.js` - Service API
- [ ] `js/markdown.js` - Utilitaires markdown
- [ ] `js/app.js` - Application Vue
- [ ] `js/components/ChatMessage.js`
- [ ] `js/components/InputBar.js`
- [ ] `js/components/StatusBar.js`
- [ ] `js/components/OperationControls.js`
- [ ] `README.md` - Documentation
- [ ] `QUICKSTART.md` - Guide rapide
- [ ] `start.sh` (exécutable)
- [ ] `start.bat` (Windows)

## 🚀 Serveur Go

1. **Vérifier que le serveur Go démarre** :
   ```bash
   cd samples/56-crew-server-agent
   go run main.go
   ```

2. **Attendre le message** :
   ```
   🚀 Server starting on http://localhost:8080
   ```

3. **Tester l'API** :
   ```bash
   curl http://localhost:8080/health
   ```

   Devrait retourner :
   ```json
   {"status":"ok","message":"Server is healthy"}
   ```

## 🌐 Serveur Web

1. **Lancer le serveur web** :
   ```bash
   cd web
   ./start.sh    # macOS/Linux
   # OU
   start.bat     # Windows
   ```

2. **Vérifier le message** :
   ```
   Starting web server on http://localhost:3000
   ```

3. **Ouvrir le navigateur** :
   ```
   http://localhost:3000
   ```

## 🧪 Tests Fonctionnels

### Test 1 : Interface Charge
- [ ] La page se charge sans erreurs
- [ ] Le titre "Nova Crew Server Agent" est visible
- [ ] La barre de statut affiche des informations
- [ ] La zone de saisie est visible
- [ ] Les boutons sont présents

### Test 2 : Envoi de Message
- [ ] Tapez "Hello, world!" dans la zone de saisie
- [ ] Appuyez sur Enter ou cliquez "Send"
- [ ] Le message utilisateur apparaît
- [ ] Une réponse de l'assistant commence à streamer
- [ ] Le texte apparaît progressivement (streaming)

### Test 3 : Markdown
- [ ] Tapez : "Write a Python hello world with explanation"
- [ ] Vérifiez que le code est dans un bloc coloré
- [ ] Vérifiez que la syntaxe Python est highlightée

### Test 4 : Routage d'Agents
- [ ] Question de code : "Write a Go function" → Agent "coder"
- [ ] Question philo : "What is consciousness?" → Agent "thinker"
- [ ] Question cuisine : "Recipe for pasta" → Agent "cook"
- [ ] Vérifiez la barre de statut pour voir l'agent actif

### Test 5 : Function Calling
- [ ] Tapez : "Say hello to Alice"
- [ ] Une notification d'opération apparaît
- [ ] Cliquez "Validate"
- [ ] La réponse contient le résultat de la fonction

### Test 6 : Contrôles
- [ ] **Stop** : Démarrez un message long, cliquez Stop
- [ ] **Clear Memory** : Conversation se réinitialise
- [ ] **View Messages** : Console affiche tous les messages
- [ ] **View Models** : Alert affiche les modèles

### Test 7 : Contexte
- [ ] Envoyez plusieurs messages
- [ ] Observez la taille du contexte augmenter
- [ ] Le nombre devrait être > 0

## 🐛 Console DevTools

Ouvrez les DevTools (F12) et vérifiez :

### Console Tab
- [ ] Aucune erreur JavaScript rouge
- [ ] Les logs montrent les appels API
- [ ] Les messages SSE sont loggés

### Network Tab
- [ ] Request à `/completion` avec status 200
- [ ] Type "EventStream" pour le streaming
- [ ] Pas d'erreurs CORS

### Sources Tab
- [ ] Tous les fichiers JS chargés
- [ ] Vue.js, Marked.js, Highlight.js présents

## 🎨 Interface Visuelle

Vérifiez que l'interface s'affiche correctement :

- [ ] **Thème sombre** : Fond noir/gris foncé
- [ ] **Messages utilisateur** : Alignés à droite, fond bleu
- [ ] **Messages assistant** : Alignés à gauche, fond gris
- [ ] **Boutons** : Couleurs appropriées (bleu, vert, rouge, orange)
- [ ] **Code blocks** : Fond noir avec coloration syntaxique
- [ ] **Scrolling** : Auto-scroll vers le bas lors de nouveaux messages

## 📱 Responsive Design

Si possible, testez sur différentes tailles d'écran :

- [ ] **Desktop** : Layout correct sur grand écran
- [ ] **Tablet** : Messages plus larges, boutons réorganisés
- [ ] **Mobile** : Layout en colonne, boutons empilés

## 🔍 Dépannage

Si quelque chose ne fonctionne pas :

### Problème : Page blanche
**Solution** :
1. Ouvrir DevTools Console
2. Vérifier les erreurs JavaScript
3. Vérifier que tous les CDN sont chargés

### Problème : "Failed to connect to server"
**Solution** :
1. Vérifier que le serveur Go est lancé
2. Tester `curl http://localhost:8080/health`
3. Vérifier l'URL dans `js/api.js`

### Problème : Pas de streaming
**Solution** :
1. Vérifier Network tab pour `/completion`
2. S'assurer que EventStream est supporté
3. Vérifier les logs du serveur Go

### Problème : Code non coloré
**Solution** :
1. Vérifier que Highlight.js est chargé (Network tab)
2. Vérifier la connexion Internet (CDN)
3. Spécifier le langage dans le code fence (\`\`\`python)

## ✨ Tests Avancés

Pour les utilisateurs avancés :

### Test de Performance
```javascript
// Dans la console DevTools
console.time('render')
// Envoyez un long message
// Quand terminé :
console.timeEnd('render')
```

### Test de Mémoire
```javascript
// Dans la console DevTools
console.log(performance.memory)
// Envoyez plusieurs messages
console.log(performance.memory)
// La mémoire ne devrait pas augmenter de façon excessive
```

### Test CORS
```bash
# Depuis un autre domaine
curl -X POST http://localhost:8080/completion \
  -H "Origin: http://example.com" \
  -H "Content-Type: application/json" \
  -d '{"data":{"message":"test"}}'
```

## 📊 Métriques de Succès

L'interface est fonctionnelle si :

- ✅ Temps de chargement < 2 secondes
- ✅ Premier token reçu < 1 seconde
- ✅ Streaming fluide sans saccades
- ✅ Pas d'erreurs dans la console
- ✅ Markdown s'affiche correctement
- ✅ Code est coloré automatiquement
- ✅ Tous les boutons fonctionnent
- ✅ Responsive sur mobile

## 🎉 Prêt à l'Emploi !

Si tous les tests passent, l'interface est prête à être utilisée !

Consultez `demo-questions.md` pour des exemples de questions à poser.

---

**Besoin d'aide ?** Consultez :
- `README.md` - Documentation complète
- `QUICKSTART.md` - Guide de démarrage
- `PROJECT-STRUCTURE.md` - Architecture technique
