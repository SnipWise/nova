# Debug: UI Reste Bloqué Après Validation

## Problème

Après avoir validé l'exécution d'un tool:
- ✅ La validation fonctionne
- ✅ La carte de notification devient verte et disparaît
- ❌ Le frontend reste en état "loading" (spinner visible)
- ❌ Impossible de saisir un nouveau message

## Cause Probable

Le stream SSE continue après la validation, mais il manque probablement l'événement final `finish_reason: "stop"` qui indique au frontend que la réponse est complète.

### Flux Attendu

```
1. User: "Say hello to Alice"
   ↓
2. Backend détecte tool call
   ↓
3. SSE: data: {"kind":"tool_call","status":"pending",...}
   ↓
4. Frontend affiche notification (WAIT)
   ↓
5. User clique "Validate"
   ↓
6. POST /operation/validate
   ↓
7. Backend exécute le tool
   ↓
8. SSE: data: {"message":"👋 Hello, Alice!🙂"}
   ↓
9. SSE: data: {"message":"","finish_reason":"stop"}  ⬅️ MANQUANT ?
   ↓
10. Frontend débloque UI
```

## Logging Ajouté

J'ai ajouté des logs pour diagnostiquer le problème:

### Dans api.js

```javascript
// Ligne 91-95: Détecte la fin du stream
if (done) {
    console.log('Stream completed (done=true)');
    onChunk('', true);  // Force unlock UI
    break;
}

// Ligne 125: Log tous les événements SSE
console.log('SSE event received:', parsed);

// Ligne 141: Log les chunks de message
console.log('Message chunk:', {chunk: chunk.substring(0, 50), finishReason});
```

## Comment Tester

### 1. Ouvrir la Console Développeur

1. Ouvrir http://localhost:3000
2. F12 → Onglet "Console"
3. Garder la console visible

### 2. Envoyer un Message avec Tool

Envoyer: **"Say hello to TestUser"**

### 3. Observer les Logs

Vous devriez voir dans la console:

```javascript
// Au début du stream
SSE event received: {kind: "tool_call", status: "pending", operation_id: "op_0x..."}
Tool call notification: {...}

// Après validation
Validating operation: op_0x...
Validation raw response: data: {"message":"✅ Operation validated"}
Validation parsed: {message: "✅ Operation validated"}

// Puis... quoi ?
// Est-ce qu'on voit d'autres événements SSE ?
SSE event received: {message: "👋 Hello, TestUser!🙂", finish_reason: ???}
```

### 4. Cas à Vérifier

**Cas 1: Stream se termine correctement**
```javascript
SSE event received: {message: "👋 Hello, TestUser!🙂", finish_reason: "stop"}
Message chunk: {chunk: "👋 Hello, TestUser!🙂", finishReason: "stop"}
Stream finished (stop reason)
```
✅ UI devrait se débloquer

**Cas 2: Stream se termine sans finish_reason**
```javascript
SSE event received: {message: "👋 Hello, TestUser!🙂"}
Message chunk: {chunk: "👋 Hello, TestUser!🙂", finishReason: undefined}
Stream completed (done=true)
```
✅ UI devrait quand même se débloquer (fix ajouté ligne 93-95)

**Cas 3: Stream ne se termine pas**
```javascript
SSE event received: {kind: "tool_call", ...}
// ... puis plus rien
```
❌ UI reste bloquée → **Problème backend**

## Solutions Possibles

### Solution 1: Backend N'envoie Pas finish_reason

Si les logs montrent que le message arrive mais sans `finish_reason: "stop"`:

**Problème**: Le Nova SDK ne renvoie pas de finish_reason après l'exécution du tool

**Solution**: Vérifier dans `main.go` si le stream est correctement fermé après l'exécution du tool

### Solution 2: Stream Se Bloque

Si les logs montrent que plus rien n'arrive après la validation:

**Problème**: Le stream attend quelque chose du backend

**Solution temporaire**: Ajouter un timeout côté frontend

```javascript
// Dans api.js, ajouter un timeout
const timeout = setTimeout(() => {
    console.warn('Stream timeout - forcing completion');
    if (onChunk) {
        onChunk('', true);
    }
    this.closeStream();
}, 30000); // 30 secondes

// Annuler le timeout si le stream se termine normalement
clearTimeout(timeout);
```

### Solution 3: Événement SSE Perdu

Si le `finish_reason: "stop"` est envoyé mais pas reçu:

**Problème**: Buffering dans le proxy ou parsing incorrect

**Solution**: Vérifier que le proxy flush correctement (déjà fait)

## Test Rapide

Pour tester immédiatement, essayez ceci dans la console du navigateur pendant que l'UI est bloquée:

```javascript
// Forcer le déblocage
api.closeStream();
// Puis dans Vue DevTools ou console
app.isLoading = false;
app.streamingMessageIndex = -1;
```

Si ça débloque l'UI, alors le problème est bien que le stream ne se termine pas correctement.

## Prochaines Étapes

1. **Tester** avec les logs activés
2. **Copier** les logs de la console ici
3. **Identifier** lequel des 3 cas se produit
4. **Appliquer** la solution correspondante

---

**Note**: J'ai déjà ajouté un fix préventif (ligne 91-95) qui devrait débloquer l'UI même si `finish_reason` n'est pas envoyé, dès que le stream se ferme côté backend.
