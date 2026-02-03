# 🚀 Démarrage rapide - Gateway Server avec qwen-code

Guide rapide pour utiliser le gateway server avec qwen-code et les outils.

## 📦 Prérequis

1. **Serveur LLM** : Un moteur llama.cpp en cours d'exécution
   ```bash
   # Le serveur doit être accessible sur http://localhost:12434
   ```

2. **qwen-code** : Installer qwen-code
   ```bash
   npm install -g @qwen-code/qwen-code
   ```

## 🎯 Démarrage en 3 étapes

### Étape 1 : Démarrer le gateway server

```bash
cd samples/85-gateway-server-agent-crew
go run main.go
```

Vous devriez voir :
```
🚀 Gateway crew server starting on http://localhost:8080
📡 OpenAI-compatible endpoint: POST /v1/chat/completions
👥 Crew agents: coder, thinker, generic
🔧 Tools mode: passthrough (client-side)
```

### Étape 2 : Configurer les variables d'environnement

```bash
export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=none
export OPENAI_MODEL=crew
```

### Étape 3 : Lancer qwen-code

```bash
qwen-code
```

C'est tout ! 🎉

## ✅ Test rapide

Une fois qwen-code lancé, testez avec :

```
You: Write a hello world in Go
```

Le gateway devrait automatiquement router vers l'agent **coder** et générer le code.

## 🛠️ Utilisation des outils

Qwen-code gère automatiquement les outils. Par exemple :

```
You: Read the file package.json and tell me the version
```

Qwen-code va :
1. Déclarer l'outil `read_file` au gateway
2. Le LLM décide d'utiliser l'outil
3. Qwen-code exécute la lecture du fichier
4. Le LLM génère la réponse avec le contenu

**Tout cela est automatique !** 🎯

## 🔍 Modes de fonctionnement

### Mode actuel : **Passthrough** (défaut)

- ✅ Qwen-code gère les outils
- ✅ Exécution côté client
- ✅ Sécurité maximale
- ✅ Flexibilité totale

### Mode alternatif : **Auto-Execute**

Pour activer le mode Auto-Execute (outils côté serveur), voir [README-tools.md](README-tools.md#configuration-avancée).

## 📊 Agents disponibles

Le gateway route automatiquement vers le bon agent :

| Agent | Déclencheurs | Use case |
|-------|-------------|----------|
| **coder** | coding, programming, development, code, software | Code, debug, refactoring |
| **thinker** | philosophy, thinking, ideas, psychology, math, science | Réflexion, analyse, problèmes complexes |
| **generic** | tout le reste | Questions générales |

## 🧪 Tester avec curl

Si vous voulez tester sans qwen-code :

```bash
# Test simple
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }' | jq .
```

Pour des exemples plus avancés avec outils :

```bash
./examples-tools.sh
```

## 🐛 Dépannage

### Erreur : "connection refused"

**Cause :** Le serveur LLM n'est pas démarré

**Solution :** Vérifiez que llama.cpp tourne sur `localhost:12434`

### Erreur : "model not found"

**Cause :** Les modèles ne sont pas téléchargés

**Solution :** Vérifiez que les modèles dans `main.go` sont disponibles :
- `hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m`
- `hf.co/menlo/lucy-gguf:q4_k_m`
- `hf.co/menlo/jan-nano-gguf:q4_k_m`

### Qwen-code ne trouve pas le modèle

**Cause :** Variable `OPENAI_MODEL` non définie

**Solution :**
```bash
export OPENAI_MODEL=crew
```

### Les outils ne fonctionnent pas

**Cause :** Qwen-code doit être configuré pour utiliser les outils

**Solution :** Vérifiez la configuration de qwen-code pour les outils disponibles

## 📚 Documentation complète

- [README-tools.md](README-tools.md) - Guide complet sur les outils
- [examples-tools.sh](examples-tools.sh) - Exemples pratiques avec curl
- [test.sh](test.sh) - Suite de tests du gateway

## 🎨 Personnalisation

### Changer le port

Modifiez dans `main.go` :

```go
gatewayserver.WithPort(8080), // Changez 8080 par votre port
```

### Modifier les agents

Ajoutez, supprimez ou modifiez les agents dans la fonction `main()` :

```go
agentCrew := map[string]*chat.Agent{
    "coder":   coderAgent,
    "thinker": thinkerAgent,
    "generic": genericAgent,
    // Ajoutez vos agents ici
}
```

### Personnaliser le routage

Modifiez la fonction `matchAgentFunction` pour changer les règles de routage :

```go
matchAgentFunction := func(currentAgentId, topic string) string {
    switch strings.ToLower(topic) {
    case "coding":
        return "coder"
    case "philosophy":
        return "thinker"
    // Ajoutez vos règles ici
    default:
        return "generic"
    }
}
```

## 💡 Conseils d'utilisation

### 1. Toujours préciser le contexte

❌ Mauvais : "Fix this"
✅ Bon : "Fix the syntax error in the Go function reverseString"

### 2. Utiliser les bons mots-clés pour le routage

- Pour du code : "write", "debug", "fix", "code", "function"
- Pour de la réflexion : "explain", "why", "philosophy", "analyze"
- Pour des questions générales : tout le reste

### 3. Profiter des outils de qwen-code

Qwen-code a accès à votre système de fichiers local, utilisez-le !

```
You: Read all .go files in the current directory and find potential bugs
```

## 🌟 Fonctionnalités avancées

### Compression automatique

Le gateway compresse automatiquement l'historique quand il dépasse 7000 caractères :

```go
gatewayserver.WithCompressorAgentAndContextSize(compressorAgent, 7000)
```

### Orchestration multi-agents

L'orchestrateur analyse automatiquement le sujet et route vers le bon agent :

```go
gatewayserver.WithOrchestratorAgent(orchestratorAgent)
```

### Hooks de cycle de vie

Vous pouvez ajouter des hooks avant/après chaque requête :

```go
gatewayserver.BeforeCompletion(func(agent *gatewayserver.GatewayServerAgent) {
    fmt.Printf("📥 Request received\n")
})
```

## 🔗 Liens utiles

- [Nova SDK Documentation](../../README.md)
- [Qwen Code GitHub](https://github.com/QwenLM/qwen-code)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)

---

**Besoin d'aide ?** Consultez [README-tools.md](README-tools.md) pour plus de détails.
