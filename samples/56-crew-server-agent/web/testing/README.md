# Testing Scripts for Web Interface

Scripts de test curl pour vérifier que l'API fonctionne correctement avec le proxy CORS.

## 🎯 Objectif

Ces scripts testent l'API à travers le **proxy CORS (port 8081)** pour s'assurer que tous les endpoints fonctionnent correctement avec l'interface web.

## 📋 Prérequis

1. **Backend démarré** :
   ```bash
   cd samples/56-crew-server-agent
   go run main.go
   ```

2. **Proxy CORS démarré** :
   ```bash
   cd samples/56-crew-server-agent/web/proxy
   go run main.go
   ```

3. **Outil jq installé** (pour parser JSON) :
   ```bash
   # macOS
   brew install jq

   # Ubuntu/Debian
   sudo apt install jq
   ```

## 🧪 Scripts de Test

### 1. Test de Santé
```bash
./test-health.sh
```
Vérifie que le serveur répond et que CORS fonctionne.

### 2. Test de Streaming
```bash
./test-stream.sh
```
Envoie un message et affiche la réponse en streaming.

### 3. Test des Modèles
```bash
./test-models.sh
```
Récupère les informations sur les modèles utilisés.

### 4. Test de la Mémoire
```bash
./test-memory.sh
```
Liste les messages, affiche la taille du contexte, et réinitialise la mémoire.

### 5. Test des Outils (Function Calling)
```bash
./test-tools.sh
```
Teste l'appel d'outils avec validation.

### 6. Test Complet
```bash
./run-all-tests.sh
```
Exécute tous les tests dans l'ordre.

## 📊 Structure des Tests

```
web/testing/
├── README.md                 # Ce fichier
├── test-health.sh           # Test santé + CORS
├── test-stream.sh           # Test streaming
├── test-models.sh           # Test récupération modèles
├── test-memory.sh           # Test gestion mémoire
├── test-tools.sh            # Test function calling
└── run-all-tests.sh         # Lance tous les tests
```

## 🔍 Vérification CORS

Tous les scripts vérifient automatiquement les headers CORS :

```bash
# Le script affiche :
✓ CORS Headers present:
  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

## 🎨 Format de Sortie

Les scripts utilisent des couleurs pour faciliter la lecture :
- 🟢 **Vert** : Succès
- 🔵 **Bleu** : Information
- 🟡 **Jaune** : Avertissement
- 🔴 **Rouge** : Erreur

## 📝 Variables d'Environnement

Tous les scripts supportent la variable `PROXY_URL` :

```bash
# Par défaut : http://localhost:8081
export PROXY_URL=http://localhost:8081

# Ou directement :
PROXY_URL=http://localhost:8081 ./test-health.sh
```

## 🐛 Dépannage

### "Connection refused"

**Problème** : Le proxy n'est pas démarré

**Solution** :
```bash
cd samples/56-crew-server-agent/web/proxy
go run main.go
```

### "502 Bad Gateway"

**Problème** : Le backend n'est pas démarré

**Solution** :
```bash
cd samples/56-crew-server-agent
go run main.go
```

### "jq: command not found"

**Problème** : jq n'est pas installé

**Solution** :
```bash
brew install jq  # macOS
sudo apt install jq  # Linux
```

## 📚 Référence API

Les scripts testent les endpoints suivants :

| Endpoint | Méthode | Description |
|---|---|---|
| `/health` | GET | Santé du serveur |
| `/completion` | POST | Génération avec streaming |
| `/completion/stop` | POST | Arrêt du streaming |
| `/memory/reset` | POST | Réinitialisation mémoire |
| `/memory/messages/list` | GET | Liste des messages |
| `/memory/messages/context-size` | GET | Taille du contexte |
| `/models` | GET | Informations modèles |
| `/operation/validate` | POST | Validation d'opération |
| `/operation/cancel` | POST | Annulation d'opération |
| `/operation/reset` | POST | Reset des opérations |

## 🎯 Utilisation pour le Debug

Ces scripts sont utiles pour :
1. Vérifier que le proxy CORS fonctionne
2. Tester l'API sans l'interface web
3. Debugger les problèmes de connexion
4. Comprendre le format des requêtes/réponses

## 📖 Exemples

### Test Simple
```bash
cd samples/56-crew-server-agent/web/testing
./test-health.sh
```

### Test avec Question Personnalisée
```bash
# Modifier test-stream.sh
USER_CONTENT="What is the meaning of life?"
./test-stream.sh
```

### Test de Validation d'Outil
```bash
# 1. Lancer l'appel d'outil
./test-tools.sh

# 2. Noter l'operation_id affiché
# 3. Valider l'opération
./validate-operation.sh <operation_id>
```

---

**Bon test ! 🧪**
