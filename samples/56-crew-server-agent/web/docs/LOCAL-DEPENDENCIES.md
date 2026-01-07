# Dépendances JavaScript Locales

## 📋 Résumé

Les dépendances JavaScript (Vue.js, Marked.js, Highlight.js) ont été téléchargées localement pour éliminer la dépendance aux CDN externes.

## 🎯 Objectif

- **Indépendance**: Pas besoin de connexion Internet pour développer
- **Performance**: Chargement plus rapide (pas de requêtes externes)
- **Fiabilité**: Pas de risque d'indisponibilité du CDN
- **Sécurité**: Contrôle total sur le code exécuté
- **Déploiement**: Application self-contained, facile à déployer

## 📁 Structure des Fichiers

### Avant (CDN)
```
web/
├── index.html (liens CDN)
└── js/
    └── ...
```

### Après (Local)
```
web/
├── index.html (liens locaux)
├── lib/                              ← NOUVEAU
│   ├── vue.global.prod.js           (144 KB)
│   ├── marked.min.js                (34 KB)
│   ├── highlight.min.js             (119 KB)
│   └── github-dark.min.css          (1.3 KB)
└── js/
    └── ...
```

## 📦 Dépendances Téléchargées

### 1. Vue.js 3.4.15
- **Fichier**: `lib/vue.global.prod.js`
- **Taille**: 144 KB
- **Source**: https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js
- **Usage**: Framework Vue.js 3 (Composition API)

### 2. Marked.js 11.1.1
- **Fichier**: `lib/marked.min.js`
- **Taille**: 34 KB
- **Source**: https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js
- **Usage**: Parsing et rendu Markdown

### 3. Highlight.js 11.9.0
- **Fichier**: `lib/highlight.min.js`
- **Taille**: 119 KB
- **Source**: https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js
- **Usage**: Coloration syntaxique du code

### 4. Highlight.js Theme (GitHub Dark)
- **Fichier**: `lib/github-dark.min.css`
- **Taille**: 1.3 KB
- **Source**: https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css
- **Usage**: Thème de coloration sombre

## 📝 Modifications du HTML

### index.html

**Avant**:
```html
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">
<link rel="stylesheet" href="css/styles.css?v=5">
<!-- ... -->
<script src="https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js"></script>
<script src="https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
```

**Après**:
```html
<link rel="stylesheet" href="lib/github-dark.min.css">
<link rel="stylesheet" href="css/styles.css?v=5">
<!-- ... -->
<!-- Dependencies (Local) -->
<script src="lib/vue.global.prod.js"></script>
<script src="lib/marked.min.js"></script>
<script src="lib/highlight.min.js"></script>
```

## 🚀 Avantages

### 1. Développement Offline
- ✅ Pas besoin de connexion Internet
- ✅ Développement en local sans dépendances externes
- ✅ Fonctionne sur des réseaux isolés

### 2. Performance
```
Avant (CDN):
- Requête DNS vers CDN
- Latence réseau variable
- Dépend de la vitesse Internet

Après (Local):
- Fichiers servis localement
- Latence minimale
- Toujours rapide
```

### 3. Fiabilité
- ✅ Pas de risque d'indisponibilité du CDN
- ✅ Pas de changements inattendus (versions figées)
- ✅ Contrôle total sur les versions

### 4. Sécurité
- ✅ Pas de requêtes vers des domaines tiers
- ✅ Contrôle total sur le code exécuté
- ✅ Pas de risque de compromission du CDN
- ✅ Conforme aux politiques de sécurité strictes

### 5. Déploiement
- ✅ Application self-contained
- ✅ Un seul répertoire à déployer
- ✅ Fonctionne sans accès Internet
- ✅ Facile à packager (Docker, etc.)

## 📊 Comparaison

| Aspect | CDN | Local | Gagnant |
|--------|-----|-------|---------|
| **Première visite** | Rapide (cache CDN) | Rapide (local) | Égalité |
| **Visites suivantes** | Très rapide (cache) | Très rapide (cache) | Égalité |
| **Offline** | ❌ Ne fonctionne pas | ✅ Fonctionne | **Local** |
| **Fiabilité** | Dépend du CDN | Toujours disponible | **Local** |
| **Sécurité** | Dépendance externe | Contrôle total | **Local** |
| **Taille bundle** | 0 KB initial | +298 KB | CDN |
| **Requêtes réseau** | +3 requêtes | 0 requêtes externes | **Local** |

## 🔧 Mise à Jour des Dépendances

### Mettre à Jour Vue.js

```bash
cd web/lib
curl -o vue.global.prod.js https://cdn.jsdelivr.net/npm/vue@3.5.0/dist/vue.global.prod.js
```

### Mettre à Jour Marked.js

```bash
cd web/lib
curl -o marked.min.js https://cdn.jsdelivr.net/npm/marked@12.0.0/marked.min.js
```

### Mettre à Jour Highlight.js

```bash
cd web/lib
curl -o highlight.min.js https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/highlight.min.js
curl -o github-dark.min.css https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/styles/github-dark.min.css
```

## 🧪 Vérification

### 1. Vérifier que les fichiers existent

```bash
ls -lh web/lib/
```

**Résultat attendu**:
```
-rw-r--r--  github-dark.min.css    (1.3K)
-rw-r--r--  highlight.min.js       (119K)
-rw-r--r--  marked.min.js          (34K)
-rw-r--r--  vue.global.prod.js     (144K)
```

### 2. Tester le chargement

```bash
# Démarrer le serveur
cd samples/56-crew-server-agent
go run main.go

# Ouvrir http://localhost:3000
```

### 3. Vérifier dans DevTools

**Network Tab**:
- ✅ `vue.global.prod.js` chargé depuis `localhost:3000`
- ✅ `marked.min.js` chargé depuis `localhost:3000`
- ✅ `highlight.min.js` chargé depuis `localhost:3000`
- ✅ Aucune requête vers CDN externes

**Console**:
- ✅ `Vue` est défini
- ✅ `marked` est défini
- ✅ `hljs` est défini

### 4. Tester Offline

1. Démarrer l'application
2. Couper la connexion Internet
3. Rafraîchir la page
4. ✅ L'application fonctionne toujours

## 📦 Taille Totale

| Dépendance | Taille | Pourcentage |
|------------|--------|-------------|
| Vue.js | 144 KB | 48% |
| Highlight.js | 119 KB | 40% |
| Marked.js | 34 KB | 11% |
| GitHub Dark CSS | 1.3 KB | 1% |
| **Total** | **298 KB** | **100%** |

**Note**: Toutes les dépendances sont minifiées et en production.

## 🔒 Intégrité des Fichiers

Pour vérifier l'intégrité des fichiers (optionnel):

```bash
# Générer les checksums
cd web/lib
shasum -a 256 *.js *.css > checksums.txt

# Vérifier les checksums
shasum -a 256 -c checksums.txt
```

## 📚 Versions Utilisées

| Library | Version | Date de release |
|---------|---------|-----------------|
| Vue.js | 3.4.15 | Jan 2024 |
| Marked.js | 11.1.1 | Dec 2023 |
| Highlight.js | 11.9.0 | Nov 2023 |

## 🎯 Bonne Pratiques

### 1. Versionner les Dépendances

Les fichiers dans `lib/` doivent être commités dans Git:

```bash
git add web/lib/
git commit -m "Add local JavaScript dependencies"
```

### 2. Documenter les Versions

Garder trace des versions dans un fichier `lib/VERSIONS.md`:

```markdown
# Versions des Dépendances

- Vue.js: 3.4.15
- Marked.js: 11.1.1
- Highlight.js: 11.9.0
```

### 3. Tester Après Mise à Jour

Toujours tester l'application après avoir mis à jour une dépendance:

```bash
# Mise à jour
curl -o lib/vue.global.prod.js https://...

# Test
go run main.go
# Ouvrir http://localhost:3000
# Vérifier que tout fonctionne
```

## 🚫 Retour aux CDN (si nécessaire)

Si vous voulez revenir aux CDN:

```html
<!-- Dans index.html, remplacer -->
<link rel="stylesheet" href="lib/github-dark.min.css">
<!-- par -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">

<!-- Et pareil pour les scripts -->
<script src="https://cdn.jsdelivr.net/npm/vue@3.4.15/dist/vue.global.prod.js"></script>
<script src="https://cdn.jsdelivr.net/npm/marked@11.1.1/marked.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
```

## 🌐 Déploiement

### Docker

Les dépendances locales facilitent le déploiement Docker:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./samples/56-crew-server-agent

FROM alpine:latest
COPY --from=builder /app/server /server
COPY ./samples/56-crew-server-agent/web /web
EXPOSE 3000 8080
CMD ["/server"]
```

**Avantage**: Tout est inclus, pas besoin d'Internet au runtime.

### Production

En production, les dépendances locales offrent:
- ✅ Déploiement reproductible
- ✅ Pas de dépendance réseau externe
- ✅ Contrôle total des versions
- ✅ Meilleure sécurité

## 📈 Impact sur la Performance

### Première Visite

| Métrique | CDN | Local |
|----------|-----|-------|
| Requêtes DNS | 3 | 0 |
| Requêtes HTTP | 3 externes | 3 locales |
| Latence | Variable | Minimale |
| Temps total | ~500ms | ~50ms |

### Visites Suivantes

| Métrique | CDN | Local |
|----------|-----|-------|
| Cache hit | ✅ (si même CDN) | ✅ (toujours) |
| Temps total | ~10ms | ~10ms |

## ✅ Checklist de Migration

- [x] Créer le dossier `web/lib/`
- [x] Télécharger Vue.js
- [x] Télécharger Marked.js
- [x] Télécharger Highlight.js
- [x] Télécharger le thème CSS
- [x] Modifier `index.html` pour utiliser les fichiers locaux
- [x] Tester l'application
- [x] Vérifier dans DevTools (pas de requêtes CDN)
- [x] Tester offline
- [x] Documenter les versions

## 📝 Conclusion

L'utilisation de dépendances JavaScript locales rend l'application:
- Plus **fiable** (pas de dépendance CDN)
- Plus **sécurisée** (contrôle total)
- Plus **performante** (pas de latence réseau)
- Plus **simple à déployer** (self-contained)

**Coût**: +298 KB de fichiers statiques (négligeable)

**Bénéfice**: Application complètement autonome 🎉

---

**Statut**: ✅ Complété
**Date**: 2026-01-07
**Taille totale**: 298 KB (minifié)
