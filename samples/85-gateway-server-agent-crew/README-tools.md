# 🛠️ Guide d'utilisation des outils avec le Gateway Server

Ce document explique comment utiliser les outils (tools/functions) avec le gateway server en mode **Passthrough**.

## 📋 Table des matières

- [Qu'est-ce que le mode Passthrough ?](#quest-ce-que-le-mode-passthrough-)
- [Comment ça fonctionne ?](#comment-ça-fonctionne-)
- [Utilisation avec qwen-code](#utilisation-avec-qwen-code)
- [Format des outils](#format-des-outils)
- [Exemples pratiques](#exemples-pratiques)
- [Mode Auto-Execute vs Passthrough](#mode-auto-execute-vs-passthrough)

## Qu'est-ce que le mode Passthrough ?

Le **mode Passthrough** (transparence) est le mode par défaut du gateway server. Dans ce mode :

- 🔄 Le gateway **transmet** les appels d'outils du LLM vers le client
- 💻 Le **client** (qwen-code, aider, continue.dev, etc.) **exécute** les outils
- 📤 Le client renvoie les résultats au gateway
- 🔁 Le gateway transmet les résultats au LLM pour continuer la conversation

### Avantages du mode Passthrough

✅ **Flexibilité** : Le client contrôle quels outils sont disponibles
✅ **Sécurité** : Les outils s'exécutent dans l'environnement du client, pas sur le serveur
✅ **Simplicité** : Pas besoin de configurer les outils côté serveur
✅ **Compatible** : Fonctionne avec tous les clients OpenAI standard

## Comment ça fonctionne ?

### Flux de communication

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│             │         │             │         │             │
│  qwen-code  │◄───────►│   Gateway   │◄───────►│     LLM     │
│  (Client)   │         │   Server    │         │   Backend   │
│             │         │             │         │             │
└─────────────┘         └─────────────┘         └─────────────┘
      │                                                 │
      │ 1. Envoie la requête avec les outils           │
      │────────────────────────────────────────────────►│
      │                                                 │
      │ 2. LLM décide d'appeler un outil                │
      │◄────────────────────────────────────────────────│
      │                                                 │
      │ 3. Client exécute l'outil                       │
      │                                                 │
      │ 4. Renvoie le résultat                          │
      │────────────────────────────────────────────────►│
      │                                                 │
      │ 5. LLM génère la réponse finale                 │
      │◄────────────────────────────────────────────────│
```

### Étapes détaillées

1. **Le client envoie une requête** avec la liste des outils disponibles
2. **Le gateway route** vers l'agent approprié (coder, thinker, generic)
3. **Le LLM décide** s'il doit utiliser un outil
4. **Le gateway renvoie** l'appel d'outil au client (avec `finish_reason: "tool_calls"`)
5. **Le client exécute** l'outil localement
6. **Le client renvoie** le résultat avec `role: "tool"`
7. **Le LLM génère** la réponse finale basée sur le résultat

## Utilisation avec qwen-code

### Configuration

```bash
export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=none
export OPENAI_MODEL=crew
```

### Lancement

```bash
# Terminal 1 : Démarrer le gateway
cd samples/85-gateway-server-agent-crew
go run main.go

# Terminal 2 : Utiliser qwen-code
qwen-code
```

### Configuration des outils dans qwen-code

Qwen-code doit être configuré avec les outils disponibles. Voici un exemple de configuration :

```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "description": "Read the contents of a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file to read"
            }
          },
          "required": ["path"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "write_file",
        "description": "Write content to a file",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file to write"
            },
            "content": {
              "type": "string",
              "description": "Content to write to the file"
            }
          },
          "required": ["path", "content"]
        }
      }
    }
  ]
}
```

## Format des outils

### Définition d'un outil (envoyée par le client)

```json
{
  "type": "function",
  "function": {
    "name": "nom_de_la_fonction",
    "description": "Description de ce que fait la fonction",
    "parameters": {
      "type": "object",
      "properties": {
        "param1": {
          "type": "string",
          "description": "Description du paramètre"
        },
        "param2": {
          "type": "number",
          "description": "Description du paramètre"
        }
      },
      "required": ["param1"]
    }
  }
}
```

### Appel d'outil (renvoyé par le LLM)

```json
{
  "id": "call_abc123",
  "type": "function",
  "function": {
    "name": "nom_de_la_fonction",
    "arguments": "{\"param1\":\"valeur1\",\"param2\":42}"
  }
}
```

### Résultat d'outil (envoyé par le client)

```json
{
  "role": "tool",
  "content": "{\"result\": \"success\", \"data\": \"...\"}",
  "tool_call_id": "call_abc123"
}
```

## Exemples pratiques

### Exemple 1 : Requête simple avec outils

**Requête initiale du client :**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What time is it?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_current_time",
          "description": "Get the current time",
          "parameters": {"type": "object", "properties": {}}
        }
      }
    ]
  }'
```

**Réponse du gateway (appel d'outil) :**

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "model": "crew",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call_xyz",
            "type": "function",
            "function": {
              "name": "get_current_time",
              "arguments": "{}"
            }
          }
        ]
      },
      "finish_reason": "tool_calls"
    }
  ]
}
```

**Requête du client avec le résultat :**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What time is it?"},
      {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call_xyz",
            "type": "function",
            "function": {
              "name": "get_current_time",
              "arguments": "{}"
            }
          }
        ]
      },
      {
        "role": "tool",
        "content": "{\"time\": \"14:30:00\"}",
        "tool_call_id": "call_xyz"
      }
    ],
    "tools": [...]
  }'
```

**Réponse finale :**

```json
{
  "id": "chatcmpl-124",
  "object": "chat.completion",
  "model": "crew",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "It is currently 14:30:00 (2:30 PM)."
      },
      "finish_reason": "stop"
    }
  ]
}
```

### Exemple 2 : Avec qwen-code (automatique)

Qwen-code gère automatiquement ce flux :

```
👤 Utilisateur : "Lis le fichier package.json"

🤖 LLM : [appelle read_file avec path="package.json"]
         ↓
💻 qwen-code : [exécute la lecture du fichier]
         ↓
🤖 LLM : "Voici le contenu de package.json: ..."
```

## Mode Auto-Execute vs Passthrough

| Aspect | Passthrough (défaut) | Auto-Execute |
|--------|---------------------|--------------|
| **Exécution** | Client | Serveur |
| **Configuration** | Outils définis par le client | Outils définis côté serveur |
| **Sécurité** | Outils dans l'environnement client | Outils dans l'environnement serveur |
| **Flexibilité** | Client contrôle les outils | Serveur contrôle les outils |
| **Use case** | Applications avec accès local (IDE, CLI) | Services web, APIs |

### Quand utiliser Passthrough ?

✅ Applications de bureau (qwen-code, IDE extensions)
✅ CLI tools qui ont accès au système de fichiers local
✅ Quand le client doit contrôler les outils disponibles
✅ Pour des raisons de sécurité (isolation des outils)

### Quand utiliser Auto-Execute ?

✅ Services web sans client intelligent
✅ APIs publiques avec outils prédéfinis
✅ Quand tous les clients doivent avoir les mêmes outils
✅ Chatbots web simples

## Configuration avancée

### Activer le mode Auto-Execute

Si vous souhaitez passer en mode Auto-Execute, modifiez [main.go](main.go) :

```go
gateway, err := gatewayserver.NewAgent(
    ctx,
    gatewayserver.WithAgentCrew(agentCrew, "generic"),
    gatewayserver.WithPort(8080),

    // Activer le mode Auto-Execute
    gatewayserver.WithToolMode(gatewayserver.ToolModeAutoExecute),
    gatewayserver.WithToolsAgent(toolsAgent),
    gatewayserver.WithExecuteFn(executeFunction),
)
```

Et décommenter les définitions d'outils dans `getToolsDefinitions()`.

## Débogage

### Vérifier les requêtes

Activez les logs détaillés :

```go
if err := os.Setenv("NOVA_LOG_LEVEL", "DEBUG"); err != nil {
    panic(err)
}
```

### Messages de diagnostic

Le gateway affiche :
- 📥 `Request received (current agent: X)` : Requête reçue
- 📤 `Response sent (agent used: X)` : Réponse envoyée
- 🔵 `Matching agent for topic: X` : Agent sélectionné

### Erreurs courantes

| Erreur | Cause | Solution |
|--------|-------|----------|
| `400 Invalid request body: json: cannot unmarshal array` | Format content incorrect | ✅ Résolu dans cette version |
| `finish_reason: "tool_calls"` mais pas de tool_calls | LLM mal configuré | Vérifier que le modèle supporte les outils |
| Pas de réponse après tool call | Client n'a pas renvoyé le résultat | Vérifier l'implémentation client |

## Support multi-modal

Le gateway supporte maintenant **trois formats** pour le champ `content` :

### 1. String simple (legacy)
```json
{"role": "user", "content": "Hello"}
```

### 2. Array de strings (qwen-code)
```json
{"role": "user", "content": ["Hello", "world"]}
```

### 3. Array d'objets (multi-modal OpenAI)
```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "Hello"},
    {"type": "image_url", "image_url": {"url": "..."}}
  ]
}
```

Tous les formats sont automatiquement convertis en texte simple par le gateway.

## Ressources

- [OpenAI Tools Documentation](https://platform.openai.com/docs/guides/function-calling)
- [Qwen Code GitHub](https://github.com/QwenLM/qwen-code)
- [Nova SDK Documentation](../../README.md)

---

**Version :** 1.0.0
**Dernière mise à jour :** 2026-02-02
