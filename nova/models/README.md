# Package Models

Le package `models` fournit des structures et fonctions pour configurer les paramètres des modèles de langage de manière indépendante du SDK sous-jacent.

## Structure Principale

### Config

`Config` représente la configuration d'un modèle avec tous ses paramètres optionnels.

```go
type Config struct {
    Name             string    // Nom du modèle (requis)
    Temperature      *float64  // Température d'échantillonnage
    TopP             *float64  // Nucleus sampling
    TopK             *int64    // Top-K sampling
    MinP             *float64  // Seuil de probabilité minimum
    MaxTokens        *int64    // Tokens maximum
    FrequencyPenalty *float64  // Pénalité de fréquence
    PresencePenalty  *float64  // Pénalité de présence
    RepeatPenalty    *float64  // Pénalité de répétition
    Seed             *int64    // Graine aléatoire
    Stop             []string  // Séquences d'arrêt
    N                *int64    // Nombre de complétions
}
```

## Fonctions Principales

### NewConfig

Crée une nouvelle configuration avec juste le nom du modèle.

```go
config := models.NewConfig("model-name")
```

### Float et Int

Fonctions helpers pour créer des pointeurs vers des valeurs.

```go
temp := models.Float(0.7)
maxTokens := models.Int(2000)
```

## Méthodes Fluent API

La structure `Config` supporte une API fluent (chainable) pour configurer facilement tous les paramètres :

```go
config := models.NewConfig("model-name").
    WithTemperature(0.7).
    WithTopP(0.95).
    WithTopK(40).
    WithMinP(0.05).
    WithMaxTokens(2000).
    WithFrequencyPenalty(0.5).
    WithPresencePenalty(0.3).
    WithRepeatPenalty(1.1).
    WithSeed(42).
    WithStop("</s>", "\n\n").
    WithN(1)
```

### Méthodes Disponibles

| Méthode | Paramètre | Description |
|---------|-----------|-------------|
| `WithTemperature(float64)` | Temperature | Température d'échantillonnage |
| `WithTopP(float64)` | TopP | Nucleus sampling |
| `WithTopK(int64)` | TopK | Top-K sampling |
| `WithMinP(float64)` | MinP | Seuil de probabilité minimum |
| `WithMaxTokens(int64)` | MaxTokens | Tokens maximum |
| `WithFrequencyPenalty(float64)` | FrequencyPenalty | Pénalité de fréquence |
| `WithPresencePenalty(float64)` | PresencePenalty | Pénalité de présence |
| `WithRepeatPenalty(float64)` | RepeatPenalty | Pénalité de répétition |
| `WithSeed(int64)` | Seed | Graine aléatoire |
| `WithStop(...string)` | Stop | Séquences d'arrêt |
| `WithN(int64)` | N | Nombre de complétions |

## Exemples d'Utilisation

### Configuration Simple

```go
import "github.com/snipwise/nova/nova/models"

config := models.NewConfig("ai/qwen2.5:1.5B-F16")
```

### Configuration avec Paramètres

```go
config := models.NewConfig("ai/qwen2.5:1.5B-F16").
    WithTemperature(0.7).
    WithMaxTokens(2000)
```

### Configuration Complète

```go
config := models.Config{
    Name:             "ai/qwen2.5:1.5B-F16",
    Temperature:      models.Float(0.7),
    TopP:             models.Float(0.95),
    TopK:             models.Int(40),
    MinP:             models.Float(0.05),
    MaxTokens:        models.Int(2000),
    FrequencyPenalty: models.Float(0.5),
    PresencePenalty:  models.Float(0.3),
    RepeatPenalty:    models.Float(1.1),
    Seed:             models.Int(42),
    Stop:             []string{"</s>", "\n\n"},
    N:                models.Int(1),
}
```

### Configurations Réutilisables

```go
var (
    // Configuration déterministe
    Deterministic = models.NewConfig("model-name").
        WithTemperature(0.0).
        WithSeed(42)

    // Configuration créative
    Creative = models.NewConfig("model-name").
        WithTemperature(0.9).
        WithTopP(0.95).
        WithPresencePenalty(0.6)

    // Configuration équilibrée
    Balanced = models.NewConfig("model-name").
        WithTemperature(0.7).
        WithMaxTokens(2000)
)
```

## Utilisation avec Chat Agent

```go
import (
    "github.com/snipwise/nova/nova/chat"
    "github.com/snipwise/nova/nova/models"
)

agent, err := chat.NewAgent(
    ctx,
    agentConfig,
    models.NewConfig("model-name").
        WithTemperature(0.7).
        WithMaxTokens(2000),
)
```

## Notes

- Tous les paramètres (sauf `Name`) sont optionnels
- Utilisez des pointeurs (`*float64`, `*int64`) pour distinguer les valeurs non définies
- Les fonctions helpers `Float()` et `Int()` créent automatiquement les pointeurs
- Certains paramètres (`TopK`, `MinP`, `RepeatPenalty`, `Stop`) peuvent être spécifiques à certains modèles/engines

## Documentation Complète

Pour une documentation détaillée sur chaque paramètre, consultez :
- [SIMPLE_AGENT_README.md](../chat/SIMPLE_AGENT_README.md)
- [MODEL_CONFIG_GUIDE.md](../chat/MODEL_CONFIG_GUIDE.md)
