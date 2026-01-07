# Cache Busting - Gestion des Versions JavaScript

## Qu'est-ce que le Cache Busting?

Le **cache busting** est une technique qui force le navigateur à télécharger la nouvelle version d'un fichier JavaScript/CSS au lieu d'utiliser la version en cache.

## Problème

Lorsque vous modifiez un fichier JavaScript, le navigateur peut continuer à utiliser l'ancienne version mise en cache, ce qui empêche vos modifications d'apparaître immédiatement.

### Exemple du Problème

```javascript
// Vous modifiez InputBar.js pour enlever confirm()
const handleResetMemory = () => {
    emit('reset-memory');  // Nouvelle version
};

// Mais le navigateur utilise encore l'ancienne version en cache:
const handleResetMemory = () => {
    if (confirm('...')) {  // Ancienne version
        emit('reset-memory');
    }
};
```

## Solution: Query Parameters de Version

### Concept

Ajouter un numéro de version comme query parameter à l'URL du script:

```html
<!-- Sans version (peut utiliser le cache) -->
<script src="js/app.js"></script>

<!-- Avec version (force le téléchargement) -->
<script src="js/app.js?v=1"></script>
```

Pour le navigateur, `app.js?v=1` et `app.js?v=2` sont des URLs **différentes**, donc il télécharge le nouveau fichier.

## Implémentation dans Notre Application

### index.html

```html
<!-- App Scripts -->
<script src="js/api.js?v=3"></script>
<script src="js/markdown.js?v=3"></script>
<script src="js/components/ChatMessage.js?v=3"></script>
<script src="js/components/OperationControls.js?v=3"></script>
<script src="js/components/StatusBar.js?v=3"></script>
<script src="js/components/InputBar.js?v=3"></script>
<script src="js/components/Modal.js?v=3"></script>
<script src="js/app.js?v=3"></script>
```

### Quand Incrémenter la Version?

#### ✅ Incrémenter Quand:

1. **Vous modifiez du code JavaScript**
   ```javascript
   // Avant modification: ?v=3
   // Après modification: ?v=4
   ```

2. **Les utilisateurs ne voient pas vos changements**
   - Symptôme: "J'ai corrigé le bug mais il est toujours là"
   - Action: Incrémenter la version

3. **Vous déployez en production**
   - Toujours incrémenter pour les déploiements

#### ❌ Ne PAS Incrémenter Quand:

1. Vous modifiez uniquement le HTML (index.html)
2. Vous modifiez uniquement le CSS inline
3. Vous ne touchez qu'aux fichiers backend (Go)

### Processus de Modification

```bash
# 1. Modifier le code JavaScript
vim js/components/InputBar.js

# 2. Incrémenter la version dans index.html
# Changer ?v=3 en ?v=4

# 3. Recharger le navigateur
# Simple F5 suffit maintenant
```

## Historique des Versions

### v1 (Version Initiale)
- Première version de l'application
- Pas de query parameters

### v2 (Ajout des Modals)
- Suppression des `confirm()` natifs
- Ajout du système de modals personnalisées
- Modal pour Clear Memory
- Modal pour View Messages
- Modal pour View Models
- Modal pour Reset Operations

### v3 (Fix Label Unknown)
- Suppression du label "UNKNOWN" dans la liste des messages
- Ajout de `v-if="msg.role"` pour masquer les labels vides

## Alternatives au Versioning Manuel

### 1. Hash de Contenu (Build Tools)

Avec des outils de build comme Webpack/Vite:

```html
<!-- Le hash change automatiquement quand le fichier change -->
<script src="js/app.a3d8f9b2.js"></script>
```

**Avantages**: Automatique, précis par fichier
**Inconvénients**: Nécessite un outil de build

### 2. Timestamp

```html
<script src="js/app.js?t=1704654321"></script>
```

**Avantages**: Unique à chaque build
**Inconvénients**: Cache invalidé même sans changement

### 3. Git Commit Hash

```html
<script src="js/app.js?v=a3d8f9b"></script>
```

**Avantages**: Traçable dans Git
**Inconvénients**: Nécessite automation

## Notre Choix: Versioning Manuel Simple

Nous utilisons le versioning manuel (`?v=1`, `?v=2`, etc.) car:

✅ **Simplicité**: Pas d'outil de build requis
✅ **Contrôle**: Vous décidez quand invalider le cache
✅ **CDN-friendly**: Fonctionne avec Vue.js et libraries CDN
✅ **Production-ready**: Suffisant pour la plupart des cas

## Commandes Utiles

### Rechercher la Version Actuelle

```bash
grep "js/app.js?v=" web/index.html
```

### Remplacer Toutes les Versions

```bash
# Passer de v=3 à v=4 partout
sed -i '' 's/\?v=3/\?v=4/g' web/index.html
```

### Vérifier que Tous les Scripts Ont la Même Version

```bash
grep "\.js?v=" web/index.html | grep -o "v=[0-9]*" | sort -u
# Devrait afficher une seule ligne: v=3
```

## Bonnes Pratiques

### ✅ DO

- Incrémenter systématiquement après modification JS
- Utiliser le même numéro pour tous les scripts
- Documenter les changements importants
- Tester après chaque incrémentation

### ❌ DON'T

- Oublier d'incrémenter après modification
- Utiliser des versions différentes pour chaque fichier
- Sauter des numéros (v=3 → v=5)
- Réutiliser un ancien numéro

## Troubleshooting

### Les Modifications Ne Sont Toujours Pas Visibles

1. **Vérifier la version dans index.html**
   ```bash
   grep "js/app.js?v=" web/index.html
   ```

2. **Hard Refresh du Navigateur**
   - Mac: `Cmd + Shift + R`
   - Windows/Linux: `Ctrl + Shift + F5`

3. **Inspecter dans DevTools**
   - F12 → Network → Filtrer "js" → Vérifier les URLs chargées
   - Chercher `app.js?v=X` dans la liste

4. **Vider le Cache Manuellement**
   - Chrome: Settings → Privacy → Clear browsing data → Cached images and files

5. **Mode Incognito**
   - Tester dans une fenêtre incognito (pas de cache)

### Le Fichier Est Chargé Mais Les Changements Ne Fonctionnent Pas

- Vérifiez que vous avez bien modifié le bon fichier
- Vérifiez qu'il n'y a pas d'erreurs JavaScript dans la console (F12)
- Vérifiez que le serveur web sert bien les derniers fichiers

## Exemple Complet

### Scénario: Ajouter une Nouvelle Fonctionnalité

```bash
# 1. Modifier le code
echo "console.log('New feature');" >> js/app.js

# 2. Ouvrir index.html
vim web/index.html

# 3. Changer manuellement
# Avant:
# <script src="js/app.js?v=3"></script>
# Après:
# <script src="js/app.js?v=4"></script>

# 4. Appliquer à tous les scripts
# Remplacer ?v=3 par ?v=4 pour tous les scripts

# 5. Recharger le navigateur
# F5 ou Cmd+R

# 6. Vérifier dans DevTools
# Network → Voir app.js?v=4 chargé
```

## Conclusion

Le cache busting avec query parameters est une solution simple et efficace pour notre application Vue.js CDN-based. En incrémentant systématiquement le numéro de version après chaque modification JavaScript, nous garantissons que les utilisateurs reçoivent toujours la dernière version du code.

**Version actuelle**: `v=3` (au 2026-01-07)

**Prochaine version**: `v=4` (lors de la prochaine modification JS)
