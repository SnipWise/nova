# Guide de Test - Validation des Outils (Human-in-the-Loop)

Ce guide explique comment tester la fonctionnalité de validation des appels d'outils.

## 🎯 Qu'est-ce que la Validation d'Outils ?

La validation d'outils (Human-in-the-Loop) permet de **confirmer manuellement** les appels de fonctions avant leur exécution. Ceci est utile pour :
- Opérations critiques (suppression, modification de données)
- Actions qui ont un impact réel (envoi d'email, paiement)
- Sécurité et contrôle

## 🔧 Outils Disponibles

Le serveur expose 3 outils de test :

| Outil | Description | Arguments |
|---|---|---|
| `say_hello` | Dit bonjour à une personne | `name` (string) |
| `calculate_sum` | Calcule la somme de 2 nombres | `a`, `b` (numbers) |
| `say_exit` | Commande d'arrêt | aucun |

## 📋 Flux de Validation

```
1. User envoie un message qui nécessite un outil
   ↓
2. Agent détecte l'outil à utiliser
   ↓
3. Agent envoie une notification SSE de type "tool_call"
   ↓
4. Interface web affiche les contrôles de validation
   ↓
5. User clique "Validate" ou "Cancel"
   ↓
6. API envoie la validation au serveur
   ↓
7. Serveur exécute (ou annule) l'outil
   ↓
8. Résultat inclus dans la réponse finale
```

## 🧪 Tests à Effectuer

### Test 1 : Say Hello (Validation)

**Message** :
```
Say hello to Alice
```

**Comportement attendu** :
1. ✅ Message utilisateur s'affiche
2. ✅ Notification d'opération apparaît avec :
   - Message : "Requesting confirmation for function: say_hello"
   - Operation ID : `op_0x...`
   - Arguments : `{"name":"Alice"}`
   - Boutons : "Validate" et "Cancel"
3. ✅ Click sur "Validate"
4. ✅ Réponse de l'agent contient : "👋 Hello, Alice!🙂"

**Logs serveur attendus** :
```
[INFO] ⁉️ Requesting confirmation for function: say_hello with args: {"name":"Alice"}
[INFO] 🟡 Tool call detected: say_hello with args: {"name":"Alice"} (ID: op_0x...)
[INFO] ⏳ Waiting for validation of operation op_0x...
[INFO] ✅ Operation validated: op_0x...
[INFO] 🟢 Executing function: say_hello with arguments: {"name":"Alice"}
```

### Test 2 : Say Hello (Annulation)

**Message** :
```
Say hello to Bob
```

**Comportement attendu** :
1. ✅ Notification d'opération apparaît
2. ✅ Click sur "Cancel"
3. ✅ Opération annulée
4. ✅ Agent répond : "Operation was cancelled"

**Logs serveur attendus** :
```
[INFO] ⁉️ Requesting confirmation for function: say_hello with args: {"name":"Bob"}
[INFO] 🟡 Tool call detected: say_hello (ID: op_0x...)
[INFO] ⏳ Waiting for validation of operation op_0x...
[INFO] ❌ Operation cancelled: op_0x...
```

### Test 3 : Calculate Sum

**Message** :
```
What is 42 plus 58?
```

**Comportement attendu** :
1. ✅ Notification d'opération apparaît
2. ✅ Arguments affichés : `{"a":42,"b":58}`
3. ✅ Click sur "Validate"
4. ✅ Réponse contient : "100"

### Test 4 : Opérations Multiples

**Message** :
```
Say hello to Alice and then calculate the sum of 10 and 20
```

**Comportement attendu** :
1. ✅ Première notification pour `say_hello`
2. ✅ Valider la première
3. ✅ Deuxième notification pour `calculate_sum`
4. ✅ Valider la deuxième
5. ✅ Réponse finale contient les deux résultats

### Test 5 : Timeout (si configuré)

Si un timeout est configuré sur le serveur :

**Message** :
```
Say hello to Charlie
```

**Comportement attendu** :
1. ✅ Notification apparaît
2. ⏱️ Attendre le timeout (ne rien faire)
3. ✅ Opération timeout automatiquement
4. ✅ Message d'erreur ou comportement de fallback

## 🔍 Vérification DevTools

### Console Tab

Pendant un appel d'outil, vous devriez voir :

```javascript
Tool call notification: {
  kind: "tool_call",
  status: "pending",
  operation_id: "op_0x...",
  message: "Requesting confirmation for function: say_hello with args: {\"name\":\"Alice\"}"
}
```

### Network Tab

1. **Requête initiale** :
   - URL : `http://localhost:8081/completion`
   - Type : `EventStream`
   - Status : `200`

2. **SSE Events** :
   - Voir les events `data: {...}` arriver
   - Chercher `"kind":"tool_call"` dans les events

3. **Requête de validation** :
   - URL : `http://localhost:8081/operation/validate`
   - Method : `POST`
   - Body : `{"operation_id":"op_0x..."}`

## 🐛 Problèmes Courants

### Notification ne s'affiche pas

**Symptômes** :
- Logs serveur montrent "Waiting for validation"
- Interface web ne montre rien

**Causes possibles** :
1. ❌ Proxy CORS ne flush pas les SSE → **Vérifier la correction dans `proxy/main.go`**
2. ❌ API JavaScript ne parse pas les notifications → **Vérifier `js/api.js` ligne ~72**
3. ❌ Composant OperationControls non monté → **Vérifier `js/app.js`**

**Solution** :
```bash
# Redémarrer le proxy avec la version corrigée
cd web/proxy
go run main.go
```

### Validation ne fonctionne pas

**Symptômes** :
- Click sur "Validate" ne fait rien
- Serveur ne reçoit pas la validation

**Diagnostic** :
```javascript
// Dans la console du navigateur
api.validateOperation('op_0x...')
  .then(result => console.log('Success:', result))
  .catch(err => console.error('Error:', err));
```

**Causes possibles** :
1. ❌ Operation ID incorrect
2. ❌ Endpoint `/operation/validate` ne répond pas
3. ❌ CORS bloque la requête

### Multiples Notifications

**Symptômes** :
- Plusieurs cartes d'opération s'affichent pour le même outil

**Cause** :
- Notifications dupliquées dans le stream

**Solution** :
- Vérifier que `pendingOperations` utilise bien `operation_id` comme clé unique

## 📊 Format des Notifications SSE

### Notification Pending

```
data: {"kind":"tool_call","status":"pending","operation_id":"op_0x140003dcbe0","message":"Requesting confirmation for function: say_hello with args: {\"name\":\"Alice\"}"}
```

### Notification Completed

```
data: {"kind":"tool_call","status":"completed","operation_id":"op_0x140003dcbe0","message":"Operation validated"}
```

### Notification Cancelled

```
data: {"kind":"tool_call","status":"cancelled","operation_id":"op_0x140003dcbe0","message":"Operation cancelled"}
```

## 🎓 Architecture Technique

### Côté Serveur (Go)

```go
// 1. Détection de l'outil
toolsAgent.DetectToolCallsLoopWithConfirmation(...)

// 2. Création de l'opération
operation := &PendingOperation{
    ID: operationID,
    Status: "pending",
}

// 3. Envoi de la notification SSE
notification := ToolCallNotification{
    Kind: "tool_call",
    Status: "pending",
    OperationID: operationID,
    Message: "Requesting confirmation...",
}
notificationChan <- notification

// 4. Attente de validation
<-operation.ValidationChan

// 5. Exécution si validé
if operation.Status == "validated" {
    result := executeFn(functionName, arguments)
}
```

### Côté Client (JavaScript)

```javascript
// 1. Reception de la notification
if (data.kind === 'tool_call') {
    pendingOperations.push(data);
}

// 2. Affichage du composant
<OperationControls :operation="op" @validate="handleValidate" />

// 3. Envoi de la validation
async handleValidate(operationId) {
    await api.validateOperation(operationId);
}
```

## ✅ Checklist de Test

- [ ] Notification s'affiche quand outil détecté
- [ ] Operation ID est présent et unique
- [ ] Boutons "Validate" et "Cancel" sont visibles
- [ ] Click "Validate" exécute l'outil
- [ ] Click "Cancel" annule l'opération
- [ ] Notification disparaît après validation/annulation
- [ ] Résultat de l'outil inclus dans la réponse
- [ ] Multiples outils gérés correctement
- [ ] Logs serveur cohérents avec l'interface
- [ ] Pas d'erreur dans la console navigateur

---

**Bon test ! 🧪**
