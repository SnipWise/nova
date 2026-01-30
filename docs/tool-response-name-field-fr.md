# Champ `name` dans les reponses d'outils — Correctif de compatibilite pour les modeles locaux

## Probleme

Certains modeles LLM locaux (par exemple **FunctionGemma**, **jan-nano**) utilisent des templates de chat Jinja qui exigent un champ `name` dans les messages de reponse d'outils. Lorsqu'un appel d'outil est execute et que son resultat est renvoye au modele, le message de reponse doit inclure le nom de la fonction appelee.

L'API standard OpenAI utilise `tool_call_id` pour correler les reponses d'outils avec les appels d'outils, donc le champ `name` n'est pas requis. Le helper `ToolMessage()` du SDK `openai-go` ne definit que `content` et `tool_call_id`.

Cela provoque une **erreur 500 Internal Server Error** avec les moteurs locaux (par exemple llama.cpp) lorsque le template de chat tente d'acceder a `message['name']` et qu'il n'existe pas :

```
Invalid tool response: 'name' must be provided.
```

## Solution appliquee (Option 3 — Toujours inclure `name`)

Nous avons choisi de **toujours inclure le nom de la fonction** dans les messages de reponse d'outils. C'est l'approche implementee dans le SDK Nova.

### Fonctionnement

Une nouvelle fonction helper `createToolResponseMessage` dans `nova-sdk/agents/tools/tools.helpers.go` construit les messages de reponse d'outils en utilisant le mecanisme `SetExtraFields` du SDK OpenAI Go pour injecter le champ `name` :

```go
func createToolResponseMessage(content string, toolCallID string, functionName string) openai.ChatCompletionMessageParamUnion {
    toolMsg := openai.ToolMessage(content, toolCallID)
    toolMsg.OfTool.SetExtraFields(map[string]any{
        "name": functionName,
    })
    return toolMsg
}
```

Le JSON resultant envoye au moteur LLM ressemble a :

```json
{
  "role": "tool",
  "content": "{\"result\": 42}",
  "tool_call_id": "call_abc123",
  "name": "calculate_sum"
}
```

### Pourquoi c'est sans risque

- Les API compatibles OpenAI **ignorent les champs inconnus** — le champ `name` supplementaire ne cause aucun probleme.
- Les modeles locaux qui **exigent** `name` fonctionnent maintenant correctement.
- Aucune configuration necessaire — cela fonctionne directement pour tous les modeles.

## Approches alternatives

### Option 1 — Ajouter `name` uniquement dans `processToolCalls()`

Au lieu de creer un helper reutilisable, le champ `name` pourrait etre injecte directement au point d'appel dans `processToolCalls()` :

```go
toolMsg := openai.ToolMessage(result.Content, toolCall.ID)
toolMsg.OfTool.SetExtraFields(map[string]any{
    "name": functionName,
})
messages = append(messages, toolMsg)
```

**Avantages :**
- Changement minimal, localise sur une seule ligne.
- Pas de nouvelle fonction a maintenir.

**Inconvenients :**
- Si d'autres parties du SDK construisent egalement des messages de reponse d'outils a l'avenir, le correctif ne s'appliquera pas la-bas.
- Moins facilement decouvrable — les developpeurs ne trouveront pas un helper clairement nomme expliquant *pourquoi* le `name` est necessaire.

### Option 2 — Conditionnel via la configuration du modele

Ajouter un flag a `models.Config` comme `IncludeNameInToolResponse` :

```go
models.Config{
    Name:                     "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature:              models.Float64(0.0),
    IncludeNameInToolResponse: models.Bool(true),
}
```

Le SDK verifierait ensuite ce flag avant d'injecter `name` :

```go
if agent.ModelConfig.IncludeNameInToolResponse != nil && *agent.ModelConfig.IncludeNameInToolResponse {
    toolMsg.OfTool.SetExtraFields(map[string]any{
        "name": functionName,
    })
}
```

**Avantages :**
- Opt-in explicite — aucun risque d'effets secondaires sur les modeles qui n'en ont pas besoin.
- Clair dans le code de l'exemple que ce modele a une exigence speciale.

**Inconvenients :**
- Ajoute de la complexite de configuration — les utilisateurs doivent savoir quels modeles ont besoin de ce flag.
- Facile a oublier, ce qui conduit a la meme erreur 500 cryptique.
- Comme `name` est inoffensif pour les API qui ne l'utilisent pas, le flag n'apporte aucun benefice reel.

## Modeles concernes

Modeles connus necessitant ce correctif :

| Modele | Exigence du template |
|--------|---------------------|
| FunctionGemma (`functiongemma-270m-it-gguf`) | `message['name']` dans la reponse d'outil |
| jan-nano (`jan-nano-gguf`) | `message['name']` dans la reponse d'outil |

Tout modele utilisant un template de chat Jinja qui accede a `message['name']` sur les messages de reponse d'outils beneficiera de ce correctif.

## Fichiers modifies

- `nova-sdk/agents/tools/tools.helpers.go` — Ajout du helper `createToolResponseMessage()` et mise a jour de `processToolCalls()` pour l'utiliser.
