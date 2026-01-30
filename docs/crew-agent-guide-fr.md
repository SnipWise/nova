# Guide complet du CrewServerAgent - Nova SDK

> **SDK** : [github.com/snipwise/nova](https://github.com/snipwise/nova)
> **Langage** : Go (golang)
> **Package** : `github.com/snipwise/nova/nova-sdk/agents/crewserver`

---

## Table des matieres

1. [Introduction](#1-introduction)
2. [Vue d'ensemble de l'architecture](#2-vue-densemble-de-larchitecture)
3. [Demarrage rapide : Agent unique](#3-demarrage-rapide--agent-unique)
4. [Crew multi-agents avec orchestrateur](#4-crew-multi-agents-avec-orchestrateur)
5. [Ajout d'outils (Function Calling)](#5-ajout-doutils-function-calling)
6. [Parallel Tool Calls](#6-parallel-tool-calls)
7. [Workflow de confirmation (Human-in-the-Loop)](#7-workflow-de-confirmation-human-in-the-loop)
8. [Integration RAG](#8-integration-rag)
9. [Compression de contexte](#9-compression-de-contexte)
10. [Exemple complet](#10-exemple-complet)
11. [Reference API](#11-reference-api)
12. [Reference configuration](#12-reference-configuration)

---

## 1. Introduction

### Qu'est-ce que le CrewServerAgent ?

Le `CrewServerAgent` est un composant central du SDK Nova. Il s'agit d'un **serveur HTTP** capable d'orchestrer **un ou plusieurs agents de chat IA** au sein d'une meme application. Il expose une API REST avec streaming SSE (Server-Sent Events) permettant a des clients web, mobiles ou CLI d'interagir avec les agents de maniere fluide et en temps reel.

Concretement, le `CrewServerAgent` :

- Encapsule un ou plusieurs `chat.Agent` dans une **equipe** (crew).
- Fournit un **routage intelligent** des requetes vers l'agent le plus adapte grace a un orchestrateur.
- Gere le **function calling** (appels d'outils) avec confirmation humaine optionnelle.
- Integre la **recherche de documents** (RAG) pour enrichir les reponses avec des connaissances externes.
- Assure la **compression automatique du contexte** pour maintenir des conversations longues sans depasser les limites du modele.

### Quand utiliser le CrewServerAgent plutot qu'un ServerAgent simple ?

| Cas d'usage | ServerAgent simple | CrewServerAgent |
|---|---|---|
| Un seul agent de chat | Oui | Oui (via `WithSingleAgent`) |
| Plusieurs agents specialises | Non | **Oui** |
| Routage automatique par sujet | Non | **Oui** |
| Orchestration multi-agents | Non | **Oui** |
| Outils + RAG + Compression | Oui | **Oui** |

**Regle generale** : utilisez le `CrewServerAgent` des que vous avez besoin de **plusieurs agents** ou d'un **routage intelligent**. Meme avec un seul agent, le `CrewServerAgent` reste un excellent choix car il offre une API unifiee et evolutive.

### Capacites principales

- **Multi-agents** : gerez une equipe d'agents specialises (codeur, redacteur, cuisinier, etc.).
- **Orchestrateur** : detection automatique du sujet de la requete et routage vers l'agent competent.
- **Function Calling** : integration d'outils externes (API, bases de donnees, calculs) avec execution sequentielle ou parallele.
- **Human-in-the-Loop** : confirmation manuelle des appels d'outils avant execution, via interface web ou fonction personnalisee.
- **RAG** (Retrieval-Augmented Generation) : enrichissement automatique du contexte avec des documents pertinents.
- **Compression de contexte** : reduction automatique de l'historique de conversation pour les sessions longues.
- **Streaming SSE** : reponses en temps reel avec notifications intermediaires (compression, outils, etc.).
- **API REST complete** : endpoints pour la completion, la gestion memoire, les operations et le monitoring.

---

## 2. Vue d'ensemble de l'architecture

### Composants

Le `CrewServerAgent` s'appuie sur plusieurs composants specialises :

```
                    +---------------------+
                    |  CrewServerAgent     |
                    |  (Serveur HTTP)      |
                    +----------+----------+
                               |
          +--------------------+--------------------+
          |         |          |          |          |
    +-----+---+ +--+---+ +----+----+ +---+----+ +--+----------+
    | Chat     | | Tools| | RAG     | | Compre-| | Orchestrator|
    | Agents   | | Agent| | Agent   | | ssor   | | Agent       |
    | (crew)   | |      | |         | | Agent  | |             |
    +----------+ +------+ +---------+ +--------+ +-------------+
```

| Composant | Role | Package |
|---|---|---|
| **Chat Agents** | Agents de conversation (un ou plusieurs) | `agents/chat` |
| **Tools Agent** | Detection et execution d'appels de fonctions | `agents/tools` |
| **RAG Agent** | Recherche de similarite dans des documents | `agents/rag` |
| **Compressor Agent** | Compression de l'historique de conversation | `agents/compressor` |
| **Orchestrator Agent** | Detection du sujet et routage vers l'agent adapte | Implemente `agents.OrchestratorAgent` |

### Flux d'une requete

Lorsqu'un client envoie une requete `POST /completion`, le traitement suit ce flux ordonne :

```
Client HTTP
    |
    v
1. Parse de la requete JSON
    |
    v
2. Configuration du streaming SSE
    |
    v
3. Compression du contexte (si necessaire)
    |  -> Notification SSE : "Compression en cours..."
    |  -> Notification SSE : "Compression terminee"
    |
    v
4. Detection et execution des appels d'outils
    |  -> Notification SSE : outil detecte (kind: "tool_call")
    |  -> Attente de confirmation (si Human-in-the-Loop)
    |  -> Execution de la fonction
    |  -> Ajout des resultats au contexte
    |
    v
5. Enrichissement RAG (recherche de similarite)
    |  -> Ajout du contexte pertinent aux messages
    |
    v
6. Routage vers l'agent adapte (orchestrateur)
    |  -> Detection du sujet
    |  -> Changement d'agent si necessaire
    |
    v
7. Generation de la completion en streaming
    |  -> Chunks SSE : {"message": "..."}
    |  -> Fin : {"message": "", "finish_reason": "stop"}
    |
    v
8. Nettoyage de l'etat des outils
```

**Points importants** :
- Les etapes 3 a 6 ne s'executent que si les composants correspondants sont configures.
- Si les outils detectent un appel de fonction et l'executent avec succes, **aucune completion supplementaire n'est generee** (les resultats d'outils sont directement envoyes au client).
- La compression envoie des notifications SSE au client avant le debut de la generation.

---

## 3. Demarrage rapide : Agent unique

Le moyen le plus simple de demarrer est d'utiliser `WithSingleAgent` pour creer un serveur avec un seul agent de chat.

### Exemple minimal

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	chatAgent, err := chat.NewAgent(ctx,
		agents.Config{
			Name:        "assistant",
			Instruction: "Tu es un assistant utile et amical. Reponds en francais.",
		},
		models.Config{
			Name: os.Getenv("CHAT_MODEL"),
			URL:  os.Getenv("ENGINE_URL"),
		},
	)
	if err != nil {
		fmt.Println("Erreur:", err)
		os.Exit(1)
	}

	crewServerAgent, err := crewserver.NewAgent(ctx,
		crewserver.WithSingleAgent(chatAgent),
		crewserver.WithPort(3500),
	)
	if err != nil {
		fmt.Println("Erreur:", err)
		os.Exit(1)
	}

	fmt.Println("Serveur demarre sur http://localhost:3500")
	if err := crewServerAgent.StartServer(); err != nil {
		fmt.Println("Erreur serveur:", err)
		os.Exit(1)
	}
}
```

### Tester avec curl

```bash
# Envoyer une question
curl -N -X POST http://localhost:3500/completion \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": "Bonjour, comment vas-tu ?"}}'

# Verifier l'etat du serveur
curl http://localhost:3500/health

# Voir les modeles utilises
curl http://localhost:3500/models

# Voir l'agent courant
curl http://localhost:3500/current-agent

# Lister les messages en memoire
curl http://localhost:3500/memory/messages/list

# Reinitialiser la memoire
curl -X POST http://localhost:3500/memory/reset
```

Le flag `-N` de curl desactive le buffering, ce qui est necessaire pour voir les evenements SSE en temps reel.

---

## 4. Crew multi-agents avec orchestrateur

Le veritable potentiel du `CrewServerAgent` se revele lorsqu'il orchestre **plusieurs agents specialises**. Chaque agent possede ses propres instructions et peut etre optimise pour un domaine precis.

### Definir plusieurs agents de chat

Chaque agent est cree independamment avec sa propre configuration :

```go
ctx := context.Background()

modelConfig := models.Config{
	Name: os.Getenv("CHAT_MODEL"),
	URL:  os.Getenv("ENGINE_URL"),
}

// Agent specialise en programmation
coderAgent, _ := chat.NewAgent(ctx,
	agents.Config{
		Name:        "coder",
		Instruction: "Tu es un expert en programmation Go. Fournis du code clair et bien commente.",
	},
	modelConfig,
)

// Agent specialise en philosophie
thinkerAgent, _ := chat.NewAgent(ctx,
	agents.Config{
		Name:        "thinker",
		Instruction: "Tu es un philosophe erudite. Reponds avec profondeur et nuance.",
	},
	modelConfig,
)

// Agent generaliste (agent par defaut)
genericAgent, _ := chat.NewAgent(ctx,
	agents.Config{
		Name:        "generic",
		Instruction: "Tu es un assistant polyvalent et amical.",
	},
	modelConfig,
)
```

### Creer l'equipe avec WithAgentCrew

Les agents sont regroupes dans une `map[string]*chat.Agent` ou la cle est l'identifiant unique de l'agent :

```go
agentCrew := map[string]*chat.Agent{
	"coder":   coderAgent,
	"thinker": thinkerAgent,
	"generic": genericAgent,
}
```

Le second parametre de `WithAgentCrew` est l'identifiant de l'**agent selectionne par defaut** :

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
	crewserver.WithAgentCrew(agentCrew, "generic"),
	crewserver.WithPort(3500),
)
```

### Ajouter un orchestrateur pour la detection de sujets

L'orchestrateur est un agent specialise qui analyse la requete de l'utilisateur pour determiner le sujet de discussion. Il doit implementer l'interface `agents.OrchestratorAgent` et fournir la methode `IdentifyIntent`.

```go
import "github.com/snipwise/nova/nova-sdk/agents/orchestrator"

orchestratorAgent, _ := orchestrator.NewAgent(ctx,
	agents.Config{
		Name: "router",
		Instruction: `Analyse la question de l'utilisateur et identifie le sujet principal.
Les sujets possibles sont : coding, philosophy, general.
Reponds uniquement avec le sujet detecte.`,
	},
	models.Config{
		Name: os.Getenv("ORCHESTRATOR_MODEL"),
		URL:  os.Getenv("ENGINE_URL"),
	},
)
```

### Definir la fonction de routage avec WithMatchAgentIdToTopicFn

La fonction de routage recoit l'identifiant de l'agent courant et le sujet detecte par l'orchestrateur, puis retourne l'identifiant de l'agent a utiliser :

```go
matchFn := func(currentAgentId string, topic string) string {
	switch strings.ToLower(topic) {
	case "coding", "programming", "development", "code", "software":
		return "coder"
	case "philosophy", "thinking", "ideas", "psychology":
		return "thinker"
	default:
		return "generic"
	}
}
```

### Assemblage complet

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
	crewserver.WithAgentCrew(agentCrew, "generic"),
	crewserver.WithOrchestratorAgent(orchestratorAgent),
	crewserver.WithMatchAgentIdToTopicFn(matchFn),
	crewserver.WithPort(3500),
)
if err != nil {
	fmt.Println("Erreur:", err)
	os.Exit(1)
}

crewServerAgent.StartServer()
```

### Transfert de contexte entre agents

Lorsque le `CrewServerAgent` change d'agent en cours de conversation, **chaque agent conserve son propre historique de messages**. Le contexte n'est pas transfere automatiquement d'un agent a l'autre. Cela permet a chaque agent de maintenir une conversation coherente dans son domaine de specialite.

Pour consulter l'agent actuellement selectionne, utilisez l'endpoint `GET /current-agent` :

```bash
curl http://localhost:3500/current-agent
# {"agent_id":"coder","model_id":"ai/qwen2.5:1.5B-F16","agent_name":"coder"}
```

---

## 5. Ajout d'outils (Function Calling)

Le `CrewServerAgent` peut detecter et executer des appels de fonctions grace a un `tools.Agent`. Cela permet a l'IA de faire appel a des services externes (API, bases de donnees, calculs, etc.) de maniere structuree.

### Creer un Tools Agent

```go
import "github.com/snipwise/nova/nova-sdk/agents/tools"

// Definir les outils disponibles
toolsList := []tools.Tool{
	tools.NewTool(
		"get_weather",
		"Obtenir la meteo actuelle pour une ville donnee",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "Nom de la ville",
				},
			},
			"required": []string{"city"},
		},
	),
	tools.NewTool(
		"calculate",
		"Effectuer un calcul mathematique",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "Expression mathematique a evaluer",
				},
			},
			"required": []string{"expression"},
		},
	),
}

// Creer l'agent d'outils
toolsAgent, err := tools.NewAgent(ctx,
	agents.Config{
		Name:        "tools-agent",
		Instruction: "Tu es un assistant avec acces a des outils. Utilise-les quand c'est pertinent.",
	},
	models.Config{
		Name:              os.Getenv("TOOLS_MODEL"),
		URL:               os.Getenv("ENGINE_URL"),
		ParallelToolCalls: models.Bool(false),
	},
	tools.WithTools(toolsList),
)
```

### Definir la fonction d'execution

La fonction d'execution recoit le nom de la fonction appelee et ses arguments au format JSON, puis retourne le resultat :

```go
executeFunction := func(functionName string, arguments string) (string, error) {
	switch functionName {
	case "get_weather":
		// Extraire les arguments et appeler l'API meteo
		return `{"temperature": 22, "condition": "ensoleille", "city": "Paris"}`, nil
	case "calculate":
		// Evaluer l'expression
		return `{"result": 42}`, nil
	default:
		return "", fmt.Errorf("fonction inconnue : %s", functionName)
	}
}
```

### Integrer au CrewServerAgent

```go
crewServerAgent, err := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithToolsAgent(toolsAgent),
	crewserver.WithExecuteFn(executeFunction),
	crewserver.WithPort(3500),
)
```

### Flux d'execution des outils

Voici le detail du flux lorsqu'un utilisateur pose une question qui necessite un outil :

1. La requete de l'utilisateur est envoyee au `tools.Agent`.
2. Le modele detecte qu'un outil est necessaire et genere un `tool_call`.
3. Le `CrewServerAgent` envoie une **notification SSE** au client (type `tool_call`).
4. Si le mode confirmation est actif, le serveur **attend la validation** de l'utilisateur.
5. Une fois confirme (ou en mode automatique), la fonction d'execution est appelee.
6. Le resultat est ajoute au contexte du chat agent courant.
7. Le resultat est egalement streame au client via SSE.

### Comportement par defaut

Lorsqu'aucune `WithConfirmationPromptFn` n'est fournie et que `ParallelToolCalls` est desactive (ou non defini), le `CrewServerAgent` utilise par defaut la methode `DetectToolCallsLoopWithConfirmation` avec la fonction `webConfirmationPrompt`. Cela signifie que :

- Les appels d'outils sont traites en **boucle sequentielle** (le modele peut demander plusieurs outils successivement).
- Chaque appel d'outil necessite une **confirmation via l'interface web** (endpoints `/operation/validate` et `/operation/cancel`).

---

## 6. Parallel Tool Calls

### Qu'est-ce que les Parallel Tool Calls ?

Par defaut, les appels d'outils sont detectes et executes en **boucle sequentielle** : le modele propose un outil, il est execute, puis le modele peut en proposer un autre, et ainsi de suite. Avec les **parallel tool calls**, le modele peut proposer **plusieurs outils en une seule passe**, et ils sont tous executes en parallele.

### Difference entre detection parallele et boucle

| Mode | Description | Methode utilisee |
|---|---|---|
| **Boucle sequentielle** | Le modele propose un outil a la fois, en boucle | `DetectToolCallsLoopWithConfirmation` |
| **Detection parallele (single-pass)** | Le modele propose plusieurs outils en une passe | `DetectParallelToolCalls` |

### Activer les Parallel Tool Calls

L'activation se fait au niveau de la configuration du modele du `tools.Agent` :

```go
toolsAgent, _ := tools.NewAgent(ctx,
	agents.Config{
		Name:        "tools-agent",
		Instruction: "Tu es un assistant avec des outils.",
	},
	models.Config{
		Name:              os.Getenv("TOOLS_MODEL"),
		URL:               os.Getenv("ENGINE_URL"),
		Temperature:       models.Float64(0.0),
		ParallelToolCalls: models.Bool(true), // Activer les appels paralleles
	},
	tools.WithTools(toolsList),
)
```

### Logique de branchement

Le `CrewServerAgent` choisit automatiquement la methode de detection en fonction de deux parametres : la configuration `ParallelToolCalls` du modele et la presence d'une fonction `ConfirmationPromptFn`.

| `ParallelToolCalls` | `ConfirmationPromptFn` | Methode utilisee |
|---|---|---|
| `true` | Non fournie | `DetectParallelToolCalls` |
| `true` | Fournie | `DetectParallelToolCallsWithConfirmation` |
| `false` ou `nil` | Non fournie | `DetectToolCallsLoopWithConfirmation` (confirmation web) |
| `false` ou `nil` | Fournie | `DetectToolCallsLoopWithConfirmation` (confirmation personnalisee) |

### Exemple : appels paralleles sans confirmation

```go
toolsAgent, _ := tools.NewAgent(ctx,
	agents.Config{
		Name:        "tools-agent",
		Instruction: "Tu es un assistant avec des outils. Utilise plusieurs outils simultanement si necessaire.",
	},
	models.Config{
		Name:              os.Getenv("TOOLS_MODEL"),
		URL:               os.Getenv("ENGINE_URL"),
		Temperature:       models.Float64(0.0),
		ParallelToolCalls: models.Bool(true),
	},
	tools.WithTools(toolsList),
)

crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithToolsAgent(toolsAgent),
	crewserver.WithExecuteFn(executeFunction),
	crewserver.WithPort(3500),
	// Pas de WithConfirmationPromptFn -> DetectParallelToolCalls est utilise
)
```

Dans ce mode, lorsque le modele detecte que plusieurs outils sont necessaires, ils sont tous appeles en parallele sans demander de confirmation a l'utilisateur. C'est le mode le plus rapide, ideal pour les scenarios ou les outils sont surs et ne necessitent pas de validation humaine.

### Exemple : appels paralleles avec confirmation

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithToolsAgent(toolsAgent),
	crewserver.WithExecuteFn(executeFunction),
	crewserver.WithConfirmationPromptFn(myConfirmationFn),
	crewserver.WithPort(3500),
	// ParallelToolCalls=true + ConfirmationPromptFn -> DetectParallelToolCallsWithConfirmation
)
```

---

## 7. Workflow de confirmation (Human-in-the-Loop)

Le workflow de confirmation permet a un humain de **valider ou refuser** chaque appel d'outil avant son execution. C'est essentiel pour les operations sensibles (envoi d'emails, modifications de base de donnees, transactions financieres, etc.).

### Confirmation web (comportement par defaut)

Lorsqu'aucune `WithConfirmationPromptFn` n'est fournie, le `CrewServerAgent` utilise automatiquement la **confirmation via interface web**. Le flux est le suivant :

1. Un appel d'outil est detecte.
2. Le serveur cree une **operation en attente** avec un identifiant unique (`operation_id`).
3. Une **notification SSE** est envoyee au client avec les details de l'outil et l'`operation_id`.
4. Le serveur **bloque** en attendant la reponse du client.
5. Le client appelle `/operation/validate` ou `/operation/cancel` avec l'`operation_id`.
6. L'outil est execute (si valide) ou la requete continue sans execution (si annule).

### Confirmation personnalisee avec WithConfirmationPromptFn

Vous pouvez fournir votre propre logique de confirmation :

```go
customConfirmation := func(functionName string, arguments string) tools.ConfirmationResponse {
	// Logique personnalisee : approuver automatiquement certaines fonctions
	if functionName == "get_weather" {
		return tools.Confirmed
	}
	// Refuser les autres
	return tools.Denied
}

crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithToolsAgent(toolsAgent),
	crewserver.WithExecuteFn(executeFunction),
	crewserver.WithConfirmationPromptFn(customConfirmation),
	crewserver.WithPort(3500),
)
```

Les valeurs de retour possibles pour `tools.ConfirmationResponse` sont :

| Valeur | Comportement |
|---|---|
| `tools.Confirmed` | L'outil est execute |
| `tools.Denied` | L'outil n'est pas execute, la conversation continue |
| `tools.Quit` | L'execution des outils est completement arretee |

### Endpoints de gestion des operations

Ces endpoints sont utilises par l'interface web pour gerer les operations en attente :

```bash
# Valider une operation
curl -X POST http://localhost:3500/operation/validate \
  -H "Content-Type: application/json" \
  -d '{"operation_id": "op_0x1234abcd"}'

# Annuler une operation
curl -X POST http://localhost:3500/operation/cancel \
  -H "Content-Type: application/json" \
  -d '{"operation_id": "op_0x1234abcd"}'

# Annuler toutes les operations en attente
curl -X POST http://localhost:3500/operation/reset
```

### Combinaison avec les Parallel Tool Calls

Lorsque `ParallelToolCalls` est active (`true`) et qu'une `ConfirmationPromptFn` est fournie, la methode `DetectParallelToolCallsWithConfirmation` est utilisee. Chaque outil detecte en parallele passe par la confirmation avant execution.

---

## 8. Integration RAG

Le RAG (Retrieval-Augmented Generation) permet d'enrichir les reponses de l'agent avec des informations provenant de documents externes. Le `CrewServerAgent` integre automatiquement le contexte RAG dans le flux de completion.

### Creer un RAG Agent

```go
import "github.com/snipwise/nova/nova-sdk/agents/rag"

ragAgent, err := rag.NewAgent(ctx,
	agents.Config{
		Name: "rag-agent",
	},
	models.Config{
		Name: os.Getenv("EMBEDDING_MODEL"),
		URL:  os.Getenv("ENGINE_URL"),
	},
)
if err != nil {
	fmt.Println("Erreur:", err)
	os.Exit(1)
}

// Ajouter des documents a l'index
ragAgent.AddDocument("Go est un langage de programmation cree par Google en 2009.")
ragAgent.AddDocument("Nova est un SDK Go pour creer des agents IA.")
ragAgent.AddDocument("Le CrewServerAgent permet d'orchestrer plusieurs agents.")
```

### Integrer avec WithRagAgent

Configuration simple (valeurs par defaut : `similarityLimit=0.6`, `maxSimilarities=3`) :

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithRagAgent(ragAgent),
	crewserver.WithPort(3500),
)
```

### Configuration avancee avec WithRagAgentAndSimilarityConfig

Pour un controle precis sur la recherche de similarite :

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithRagAgentAndSimilarityConfig(
		ragAgent,
		0.7,  // similarityLimit : seuil minimum de similarite (0.0 a 1.0)
		5,    // maxSimilarities : nombre maximum de documents retournes
	),
	crewserver.WithPort(3500),
)
```

| Parametre | Description | Valeur par defaut |
|---|---|---|
| `similarityLimit` | Seuil minimum de similarite cosinus. Les documents en dessous de ce seuil sont ignores. | `0.6` |
| `maxSimilarities` | Nombre maximum de documents similaires a injecter dans le contexte. | `3` |

### Comment le contexte RAG est injecte

Lors de chaque requete de completion (etape 5 du flux), si un `ragAgent` est configure :

1. La question de l'utilisateur est utilisee pour effectuer une **recherche de similarite** dans l'index RAG.
2. Les documents correspondants (au-dessus du seuil de similarite) sont recuperes.
3. Ils sont concatenes et ajoutes comme **message systeme** dans le contexte du chat agent courant :
   ```
   Relevant information to help you answer the question:
   [Document 1]
   ---
   [Document 2]
   ---
   ```
4. Le chat agent genere sa reponse en tenant compte de ce contexte enrichi.

---

## 9. Compression de contexte

Les conversations longues accumulent un historique qui peut depasser les limites du modele ou degrader les performances. Le `CrewServerAgent` integre un mecanisme de **compression automatique** du contexte.

### Creer un Compressor Agent

```go
import "github.com/snipwise/nova/nova-sdk/agents/compressor"

compressorAgent, err := compressor.NewAgent(ctx,
	agents.Config{
		Name:        "compressor",
		Instruction: "Resume la conversation en conservant les informations essentielles.",
	},
	models.Config{
		Name: os.Getenv("CHAT_MODEL"),
		URL:  os.Getenv("ENGINE_URL"),
	},
)
```

### Integrer avec WithCompressorAgent

Configuration simple (valeur par defaut du seuil : `8000` caracteres) :

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithCompressorAgent(compressorAgent),
	crewserver.WithPort(3500),
)
```

### Configuration avec seuil personnalise

```go
crewServerAgent, _ := crewserver.NewAgent(ctx,
	crewserver.WithSingleAgent(chatAgent),
	crewserver.WithCompressorAgentAndContextSize(
		compressorAgent,
		4000, // Seuil en caracteres : compresser au-dela de 4000
	),
	crewserver.WithPort(3500),
)
```

### Compression automatique avec notifications SSE

La compression se declenche **automatiquement** au debut de chaque requete de completion, **avant** la detection d'outils et la generation de reponse. Le client recoit des notifications SSE en temps reel :

1. **Debut** : `{"role": "information", "content": "Context size limit reached. Compressing conversation history..."}`
2. **Succes** : `{"role": "information", "content": "Compression completed. Context reduced from 8500 to 2100 bytes."}`
3. **Erreur** (si applicable) : `{"role": "information", "content": "Compression failed: ..."}`

### Fonctionnement de la compression

Lorsque la taille du contexte depasse le seuil configure :

1. L'ensemble des messages de l'agent de chat courant est envoye au `compressor.Agent`.
2. Le compresseur genere un **resume compresse** de la conversation.
3. Les messages de l'agent sont **reinitialises**.
4. Le resume compresse est ajoute comme **message systeme**.
5. La conversation reprend avec un contexte reduit.

**Attention** : si le texte compresse depasse 80% du seuil, une erreur est retournee. Si il depasse 90%, les messages sont reinitialises completement pour eviter une boucle infinie de compression.

---

## 10. Exemple complet

Voici un exemple complet combinant toutes les fonctionnalites : crew multi-agents, orchestrateur, outils avec appels paralleles, RAG et compression de contexte.

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/agents/chat"
	"github.com/snipwise/nova/nova-sdk/agents/compressor"
	"github.com/snipwise/nova/nova-sdk/agents/crewserver"
	"github.com/snipwise/nova/nova-sdk/agents/orchestrator"
	"github.com/snipwise/nova/nova-sdk/agents/rag"
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/models"
)

func main() {
	ctx := context.Background()

	engineURL := os.Getenv("ENGINE_URL")
	chatModel := os.Getenv("CHAT_MODEL")
	toolsModel := os.Getenv("TOOLS_MODEL")
	orchestratorModel := os.Getenv("ORCHESTRATOR_MODEL")
	embeddingModel := os.Getenv("EMBEDDING_MODEL")

	// -----------------------------------------------
	// 1. Creer les agents de chat (crew)
	// -----------------------------------------------
	coderAgent, _ := chat.NewAgent(ctx,
		agents.Config{
			Name:        "coder",
			Instruction: "Tu es un expert en programmation Go. Fournis du code clair et bien commente.",
		},
		models.Config{Name: chatModel, URL: engineURL},
	)

	thinkerAgent, _ := chat.NewAgent(ctx,
		agents.Config{
			Name:        "thinker",
			Instruction: "Tu es un philosophe erudite. Reponds avec profondeur et nuance.",
		},
		models.Config{Name: chatModel, URL: engineURL},
	)

	genericAgent, _ := chat.NewAgent(ctx,
		agents.Config{
			Name:        "generic",
			Instruction: "Tu es un assistant polyvalent et amical.",
		},
		models.Config{Name: chatModel, URL: engineURL},
	)

	agentCrew := map[string]*chat.Agent{
		"coder":   coderAgent,
		"thinker": thinkerAgent,
		"generic": genericAgent,
	}

	// -----------------------------------------------
	// 2. Creer l'orchestrateur
	// -----------------------------------------------
	orchestratorAgent, _ := orchestrator.NewAgent(ctx,
		agents.Config{
			Name: "router",
			Instruction: `Analyse la question et identifie le sujet.
Sujets possibles : coding, philosophy, general.`,
		},
		models.Config{Name: orchestratorModel, URL: engineURL},
	)

	matchFn := func(currentAgentId string, topic string) string {
		switch strings.ToLower(topic) {
		case "coding", "programming", "development", "code":
			return "coder"
		case "philosophy", "thinking", "ideas":
			return "thinker"
		default:
			return "generic"
		}
	}

	// -----------------------------------------------
	// 3. Creer l'agent d'outils (parallel)
	// -----------------------------------------------
	toolsList := []tools.Tool{
		tools.NewTool("get_weather", "Obtenir la meteo", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{"type": "string", "description": "Nom de la ville"},
			},
			"required": []string{"city"},
		}),
		tools.NewTool("search_docs", "Rechercher dans la documentation", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Terme de recherche"},
			},
			"required": []string{"query"},
		}),
	}

	toolsAgent, _ := tools.NewAgent(ctx,
		agents.Config{
			Name:        "tools-agent",
			Instruction: "Utilise les outils quand c'est pertinent.",
		},
		models.Config{
			Name:              toolsModel,
			URL:               engineURL,
			Temperature:       models.Float64(0.0),
			ParallelToolCalls: models.Bool(true),
		},
		tools.WithTools(toolsList),
	)

	executeFunction := func(functionName string, arguments string) (string, error) {
		switch functionName {
		case "get_weather":
			return `{"temperature": 18, "condition": "nuageux"}`, nil
		case "search_docs":
			return `{"results": ["Document 1", "Document 2"]}`, nil
		default:
			return "", fmt.Errorf("fonction inconnue : %s", functionName)
		}
	}

	// -----------------------------------------------
	// 4. Creer le RAG agent
	// -----------------------------------------------
	ragAgent, _ := rag.NewAgent(ctx,
		agents.Config{Name: "rag-agent"},
		models.Config{Name: embeddingModel, URL: engineURL},
	)

	ragAgent.AddDocument("Nova est un SDK Go pour creer des agents IA.")
	ragAgent.AddDocument("Le CrewServerAgent orchestre plusieurs agents de chat.")
	ragAgent.AddDocument("Go a ete cree par Google en 2009.")

	// -----------------------------------------------
	// 5. Creer le compressor agent
	// -----------------------------------------------
	compressorAgent, _ := compressor.NewAgent(ctx,
		agents.Config{
			Name:        "compressor",
			Instruction: "Resume la conversation en conservant les informations essentielles.",
		},
		models.Config{Name: chatModel, URL: engineURL},
	)

	// -----------------------------------------------
	// 6. Assembler le CrewServerAgent
	// -----------------------------------------------
	crewServerAgent, err := crewserver.NewAgent(ctx,
		crewserver.WithAgentCrew(agentCrew, "generic"),
		crewserver.WithOrchestratorAgent(orchestratorAgent),
		crewserver.WithMatchAgentIdToTopicFn(matchFn),
		crewserver.WithToolsAgent(toolsAgent),
		crewserver.WithExecuteFn(executeFunction),
		crewserver.WithRagAgentAndSimilarityConfig(ragAgent, 0.7, 5),
		crewserver.WithCompressorAgentAndContextSize(compressorAgent, 4000),
		crewserver.WithPort(3500),
	)
	if err != nil {
		fmt.Println("Erreur:", err)
		os.Exit(1)
	}

	// -----------------------------------------------
	// 7. Demarrer le serveur
	// -----------------------------------------------
	fmt.Println("CrewServerAgent demarre sur http://localhost:3500")
	if err := crewServerAgent.StartServer(); err != nil {
		fmt.Println("Erreur serveur:", err)
		os.Exit(1)
	}
}
```

---

## 11. Reference API

Le `CrewServerAgent` expose les endpoints HTTP suivants. Toutes les reponses sont au format JSON, sauf les endpoints de streaming qui utilisent le protocole SSE (Server-Sent Events).

### POST /completion

Envoie un message a l'agent et recoit une reponse en streaming SSE.

**Requete** :
```json
{
  "data": {
    "message": "Bonjour, comment vas-tu ?"
  }
}
```

**Reponse** (flux SSE) :
```
data: {"message":"Bonjour"}

data: {"message":" ! Je"}

data: {"message":" vais"}

data: {"message":" bien."}

data: {"message":"","finish_reason":"stop"}
```

**Notifications intermediaires possibles** :

Compression :
```
data: {"role":"information","content":"Context size limit reached. Compressing..."}
data: {"role":"information","content":"Compression completed. Context reduced from 8500 to 2100 bytes."}
```

Appel d'outil :
```
data: {"kind":"tool_call","status":"pending","operation_id":"op_0x1234","message":"Tool call detected: get_weather"}
```

Resultat d'outil :
```
data: {"message":"<hr>La temperature a Paris est de 18 degres.<hr>"}
```

Erreur :
```
data: {"error":"description de l'erreur"}
```

---

### POST /completion/stop

Interrompt la generation en cours.

**Requete** : aucun corps necessaire.

**Reponse** :
```json
{
  "status": "ok",
  "message": "Stream stopped"
}
```
ou :
```json
{
  "status": "ok",
  "message": "No stream to stop"
}
```

---

### POST /memory/reset

Reinitialise l'historique de conversation de l'agent courant et de l'agent d'outils.

**Requete** : aucun corps necessaire.

**Reponse** :
```json
{
  "status": "ok",
  "message": "Memory reset successfully"
}
```

---

### GET /memory/messages/list

Retourne l'historique complet des messages de l'agent courant.

**Reponse** :
```json
{
  "messages": [
    {"role": "system", "content": "Tu es un assistant..."},
    {"role": "user", "content": "Bonjour"},
    {"role": "assistant", "content": "Bonjour ! Comment puis-je vous aider ?"}
  ]
}
```

---

### GET /memory/messages/context-size

Retourne les statistiques de taille du contexte courant.

**Reponse** :
```json
{
  "messages_count": 5,
  "characters_count": 1250,
  "limit": 8000
}
```

---

### POST /operation/validate

Valide une operation en attente (confirmation d'un appel d'outil).

**Requete** :
```json
{
  "operation_id": "op_0x1234abcd"
}
```

**Reponse** (flux SSE) :
```
data: {"message":"Operation op_0x1234abcd validated"}
```

---

### POST /operation/cancel

Annule une operation en attente.

**Requete** :
```json
{
  "operation_id": "op_0x1234abcd"
}
```

**Reponse** (flux SSE) :
```
data: {"message":"Operation op_0x1234abcd cancelled"}
```

---

### POST /operation/reset

Annule toutes les operations en attente.

**Requete** : aucun corps necessaire.

**Reponse** (flux SSE) :
```
data: {"message":"All pending operations cancelled (2 operations)"}
```

---

### GET /models

Retourne les informations sur les modeles utilises.

**Reponse** :
```json
{
  "status": "ok",
  "chat_model": "ai/qwen2.5:1.5B-F16",
  "embeddings_model": "ai/mxbai-embed-large",
  "tools_model": "hf.co/menlo/jan-nano-gguf:q4_k_m"
}
```

Si un composant n'est pas configure, la valeur est `"none"`.

---

### GET /health

Verification de l'etat du serveur.

**Reponse** :
```json
{
  "status": "ok"
}
```

---

### GET /current-agent

Retourne les informations sur l'agent actuellement selectionne.

**Reponse** :
```json
{
  "agent_id": "coder",
  "model_id": "ai/qwen2.5:1.5B-F16",
  "agent_name": "coder"
}
```

---

## 12. Reference configuration

Toutes les options de configuration disponibles lors de la creation d'un `CrewServerAgent` via `crewserver.NewAgent()`.

### WithAgentCrew

```go
func WithAgentCrew(agentCrew map[string]*chat.Agent, selectedAgentId string) CrewServerAgentOption
```

Definit l'equipe d'agents de chat et l'identifiant de l'agent selectionne par defaut. La map ne peut pas etre vide et l'identifiant doit correspondre a une cle existante dans la map.

**Parametres** :
- `agentCrew` : map associant un identifiant unique a chaque `chat.Agent`.
- `selectedAgentId` : identifiant de l'agent actif au demarrage.

---

### WithSingleAgent

```go
func WithSingleAgent(chatAgent *chat.Agent) CrewServerAgentOption
```

Cree une equipe contenant un seul agent avec l'identifiant `"single"`. C'est un raccourci pour `WithAgentCrew(map[string]*chat.Agent{"single": chatAgent}, "single")`.

**Parametre** :
- `chatAgent` : l'agent de chat a utiliser.

---

### WithPort

```go
func WithPort(port int) CrewServerAgentOption
```

Definit le port HTTP du serveur. La valeur par defaut est `3500`.

**Parametre** :
- `port` : numero de port (entier).

---

### WithMatchAgentIdToTopicFn

```go
func WithMatchAgentIdToTopicFn(fn func(string, string) string) CrewServerAgentOption
```

Definit la fonction de routage qui associe un sujet detecte a un identifiant d'agent.

**Parametre** :
- `fn` : fonction recevant `(currentAgentId, detectedTopic)` et retournant l'identifiant de l'agent cible.

Si non fournie, une fonction par defaut est utilisee qui retourne le premier agent de la map.

---

### WithExecuteFn

```go
func WithExecuteFn(fn func(string, string) (string, error)) CrewServerAgentOption
```

Definit la fonction d'execution des appels d'outils.

**Parametre** :
- `fn` : fonction recevant `(functionName, argumentsJSON)` et retournant `(resultJSON, error)`.

Si non fournie, une fonction par defaut retourne une erreur `"executeFunction not implemented"`.

---

### WithToolsAgent

```go
func WithToolsAgent(toolsAgent *tools.Agent) CrewServerAgentOption
```

Attache un agent d'outils pour activer le function calling.

**Parametre** :
- `toolsAgent` : instance de `tools.Agent` configuree avec les outils disponibles.

---

### WithConfirmationPromptFn

```go
func WithConfirmationPromptFn(fn func(string, string) tools.ConfirmationResponse) CrewServerAgentOption
```

Definit une fonction de confirmation personnalisee pour les appels d'outils. Lorsqu'elle est fournie, elle remplace la confirmation web par defaut.

**Parametre** :
- `fn` : fonction recevant `(functionName, argumentsJSON)` et retournant `tools.Confirmed`, `tools.Denied` ou `tools.Quit`.

Lorsque cette option est combinee avec `ParallelToolCalls: models.Bool(true)` sur le `tools.Agent`, la methode `DetectParallelToolCallsWithConfirmation` est utilisee.

---

### WithCompressorAgent

```go
func WithCompressorAgent(compressorAgent *compressor.Agent) CrewServerAgentOption
```

Attache un agent de compression de contexte. Le seuil de taille par defaut est `8000` caracteres.

**Parametre** :
- `compressorAgent` : instance de `compressor.Agent`.

---

### WithCompressorAgentAndContextSize

```go
func WithCompressorAgentAndContextSize(compressorAgent *compressor.Agent, contextSizeLimit int) CrewServerAgentOption
```

Attache un agent de compression et definit le seuil de declenchement.

**Parametres** :
- `compressorAgent` : instance de `compressor.Agent`.
- `contextSizeLimit` : taille en caracteres au-dela de laquelle la compression se declenche.

---

### WithRagAgent

```go
func WithRagAgent(ragAgent *rag.Agent) CrewServerAgentOption
```

Attache un agent RAG pour la recherche de documents. Les valeurs par defaut sont `similarityLimit=0.6` et `maxSimilarities=3`.

**Parametre** :
- `ragAgent` : instance de `rag.Agent` avec des documents indexes.

---

### WithRagAgentAndSimilarityConfig

```go
func WithRagAgentAndSimilarityConfig(ragAgent *rag.Agent, similarityLimit float64, maxSimilarities int) CrewServerAgentOption
```

Attache un agent RAG et configure les parametres de recherche de similarite.

**Parametres** :
- `ragAgent` : instance de `rag.Agent`.
- `similarityLimit` : seuil minimum de similarite cosinus (entre `0.0` et `1.0`).
- `maxSimilarities` : nombre maximum de documents a injecter dans le contexte.

---

### WithOrchestratorAgent

```go
func WithOrchestratorAgent(orchestratorAgent agents.OrchestratorAgent) CrewServerAgentOption
```

Attache un agent orchestrateur pour la detection de sujets et le routage automatique vers l'agent adapte.

**Parametre** :
- `orchestratorAgent` : instance implementant l'interface `agents.OrchestratorAgent`.

Doit etre utilise conjointement avec `WithMatchAgentIdToTopicFn` pour que le routage fonctionne correctement.
