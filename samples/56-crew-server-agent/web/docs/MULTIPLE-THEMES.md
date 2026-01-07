# Système de Thèmes Multiples

## 📋 Résumé

L'application supporte désormais plusieurs thèmes CSS pour personnaliser l'apparence de l'interface. Vous pouvez facilement basculer entre différents thèmes en modifiant une ligne dans `index.html`.

## 🎨 Thèmes Disponibles

### 1. **VSCode Theme** (Par défaut)
- **Fichier**: `css/styles.css`
- **Description**: Thème inspiré de VSCode avec variables natives
- **Highlight.js**: `vs2015.min.css`
- **Caractéristiques**:
  - Utilise les variables CSS VSCode (`--vscode-*`)
  - Fond éditeur sombre (`#1e1e1e`)
  - Bordures subtiles
  - Boutons style VSCode
  - Coloration syntaxique vs2015

### 2. **Demo Theme** (Ancien thème)
- **Fichier**: `css/styles.demo.css`
- **Description**: Thème démo original avec couleurs personnalisées
- **Highlight.js**: `github-dark.min.css`
- **Caractéristiques**:
  - Couleurs personnalisées fixes
  - Fond sombre (`#1a1a1a`)
  - Style moderne
  - Boutons colorés (bleu, vert, rouge, orange)
  - Coloration syntaxique GitHub Dark

## 🔄 Changer de Thème

### Méthode Simple

Éditez `index.html` et modifiez la ligne de la feuille de style:

**Pour le thème VSCode** (défaut):
```html
<!-- Highlight.js Theme (vs2015 for VSCode theme) -->
<link rel="stylesheet" href="lib/vs2015.min.css">
<!-- Main Stylesheet (switch between styles.css and styles.demo.css) -->
<link rel="stylesheet" href="css/styles.css?v=6">
```

**Pour le thème Demo**:
```html
<!-- Highlight.js Theme (github-dark for Demo theme) -->
<link rel="stylesheet" href="lib/github-dark.min.css">
<!-- Main Stylesheet (switch between styles.css and styles.demo.css) -->
<link rel="stylesheet" href="css/styles.demo.css?v=6">
```

### Cache Busting

Après avoir changé de thème, n'oubliez pas d'incrémenter le numéro de version:

```html
<!-- Avant -->
<link rel="stylesheet" href="css/styles.css?v=6">

<!-- Après -->
<link rel="stylesheet" href="css/styles.css?v=7">
```

Puis faites un **hard refresh** dans le navigateur:
- **macOS**: `Cmd + Shift + R`
- **Windows/Linux**: `Ctrl + Shift + R`

## 📁 Structure des Fichiers

```
web/
├── index.html                      # HTML principal (sélection du thème)
├── lib/
│   ├── vs2015.min.css             # Thème highlight.js pour VSCode
│   └── github-dark.min.css        # Thème highlight.js pour Demo
└── css/
    ├── styles.css                 # Thème VSCode (actif)
    └── styles.demo.css            # Thème Demo (alternatif)
```

## 🎯 Comparaison des Thèmes

| Aspect | VSCode Theme | Demo Theme |
|--------|-------------|------------|
| **Variables CSS** | VSCode natives (`--vscode-*`) | Couleurs fixes |
| **Fond principal** | `#1e1e1e` (editor) | `#1a1a1a` |
| **Fond secondaire** | `#2d2d30` (codeblock) | `#2d2d2d` |
| **Bordures** | `#3c3c3c` (panel) | `#404040` |
| **Bouton primaire** | `#0e639c` (button) | `#1e88e5` |
| **Bouton succès** | `#43a047` (green) | `#43a047` |
| **Bouton danger** | Variable (error) | `#e53935` |
| **Code highlighting** | vs2015 | github-dark |
| **Intégration** | VSCode extension | Standalone |
| **Compatibilité** | VSCode webviews | Tous navigateurs |

## 🔧 Créer un Nouveau Thème

### 1. Dupliquer un Thème Existant

```bash
cd web/css
cp styles.css styles.custom.css
```

### 2. Personnaliser les Couleurs

Éditez `styles.custom.css` et modifiez les couleurs:

**Pour un thème VSCode** (avec variables):
```css
body {
    background-color: var(--vscode-editor-background, #1e1e1e);
    color: var(--vscode-foreground, #cccccc);
}
```

**Pour un thème standalone** (couleurs fixes):
```css
body {
    background-color: #1a1a1a;  /* Votre couleur */
    color: #e0e0e0;             /* Votre couleur */
}
```

### 3. Choisir un Thème Highlight.js

Parcourez les thèmes disponibles:
- [Highlight.js Demo](https://highlightjs.org/static/demo/)

Téléchargez le thème choisi:
```bash
cd web/lib
curl -o my-theme.min.css https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/MY-THEME.min.css
```

### 4. Activer le Nouveau Thème

Modifiez `index.html`:
```html
<link rel="stylesheet" href="lib/my-theme.min.css">
<link rel="stylesheet" href="css/styles.custom.css?v=1">
```

## 🎨 Thèmes Highlight.js Populaires

| Thème | Style | Fichier |
|-------|-------|---------|
| **vs2015** | VSCode dark | `vs2015.min.css` |
| **github-dark** | GitHub dark | `github-dark.min.css` |
| **monokai** | Sublime Text | `monokai.min.css` |
| **atom-one-dark** | Atom editor | `atom-one-dark.min.css` |
| **dracula** | Dracula theme | `dracula.min.css` |
| **nord** | Nord theme | `nord.min.css` |
| **tokyo-night** | Tokyo Night | `tokyo-night-dark.min.css` |

## 📚 Variables VSCode Utilisées

Le thème VSCode utilise les variables CSS suivantes:

### Couleurs de Base
```css
--vscode-font-family              /* Police principale */
--vscode-editor-font-family       /* Police monospace */
--vscode-editor-font-size         /* Taille police */
--vscode-foreground               /* Texte principal */
--vscode-editor-background        /* Fond éditeur */
```

### Composants UI
```css
--vscode-panel-border             /* Bordures */
--vscode-input-background         /* Fond inputs */
--vscode-input-foreground         /* Texte inputs */
--vscode-input-border             /* Bordure inputs */
--vscode-input-placeholderForeground  /* Placeholder */
--vscode-focusBorder              /* Bordure focus */
```

### Boutons
```css
--vscode-button-background        /* Bouton primaire */
--vscode-button-foreground        /* Texte bouton */
--vscode-button-hoverBackground   /* Hover primaire */
--vscode-button-secondaryBackground    /* Bouton secondaire */
--vscode-button-secondaryForeground    /* Texte secondaire */
--vscode-button-secondaryHoverBackground  /* Hover secondaire */
```

### Code et Markdown
```css
--vscode-textCodeBlock-background      /* Fond code blocks */
--vscode-textPreformat-background      /* Fond pre */
--vscode-textBlockQuote-border         /* Bordure quotes */
--vscode-textBlockQuote-background     /* Fond quotes */
--vscode-textLink-foreground           /* Couleur liens */
```

### Messages et États
```css
--vscode-errorForeground               /* Texte erreur */
--vscode-inputValidation-errorBackground   /* Fond erreur */
--vscode-inputValidation-errorBorder       /* Bordure erreur */
--vscode-editorWarning-foreground      /* Texte warning */
--vscode-editorWarning-background      /* Fond warning */
--vscode-terminal-ansiGreen            /* Vert terminal */
--vscode-terminal-ansiCyan             /* Cyan terminal */
```

### Scrollbar
```css
--vscode-scrollbarSlider-background        /* Fond scrollbar */
--vscode-scrollbarSlider-activeBackground  /* Scrollbar active */
--vscode-scrollbarSlider-hoverBackground   /* Scrollbar hover */
```

### Listes et Sélection
```css
--vscode-list-hoverBackground              /* Hover liste */
--vscode-list-activeSelectionBackground    /* Sélection active */
--vscode-list-activeSelectionForeground    /* Texte sélection */
--vscode-editor-inactiveSelectionBackground  /* Sélection inactive */
```

## 🔍 Fallback des Variables

Toutes les variables VSCode ont des valeurs de fallback pour fonctionner hors VSCode:

```css
color: var(--vscode-foreground, #cccccc);
```

Si `--vscode-foreground` n'existe pas, `#cccccc` sera utilisé.

## 🧪 Tester les Thèmes

### 1. Démarrer le Serveur

```bash
cd samples/56-crew-server-agent
go run main.go
```

### 2. Ouvrir dans le Navigateur

```
http://localhost:3000
```

### 3. Changer de Thème

1. Arrêter le serveur (`Ctrl+C`)
2. Modifier `index.html` (changer le lien CSS)
3. Redémarrer le serveur
4. Hard refresh dans le navigateur

### 4. Tester dans VSCode Webview

Le thème VSCode est optimisé pour les webviews VSCode où toutes les variables `--vscode-*` sont automatiquement définies.

## 📊 Performance

### Taille des Thèmes

| Fichier | Taille | Lignes |
|---------|--------|--------|
| `styles.css` (VSCode) | ~35 KB | 698 lignes |
| `styles.demo.css` (Demo) | ~35 KB | 698 lignes |
| `vs2015.min.css` | 1.1 KB | Minifié |
| `github-dark.min.css` | 1.3 KB | Minifié |

### Impact sur le Chargement

- **Première visite**: +36 KB (CSS + highlight theme)
- **Visites suivantes**: Cache hit (0 KB)
- **Changement de thème**: Hard refresh requis

## 🎯 Bonnes Pratiques

### 1. Choisir le Bon Thème

- **VSCode Theme**: Pour intégration dans VSCode extension
- **Demo Theme**: Pour application web standalone
- **Custom Theme**: Pour branding personnalisé

### 2. Maintenir la Cohérence

Assurez-vous que le thème highlight.js correspond au thème CSS:
- VSCode → vs2015
- Demo → github-dark
- Custom → thème compatible

### 3. Documenter les Modifications

Si vous créez un thème custom, documentez:
- Palette de couleurs utilisée
- Thème highlight.js associé
- Raisons du choix de design

### 4. Tester l'Accessibilité

Vérifiez que votre thème respecte:
- Contraste minimum WCAG AA (4.5:1)
- Lisibilité du code
- Visibilité des états (hover, focus, disabled)

## 🚀 Déploiement

### Production avec un Seul Thème

Pour réduire la taille du bundle en production:

1. Supprimer les thèmes inutilisés:
```bash
cd web/css
rm styles.demo.css  # Si vous utilisez VSCode theme
```

2. Supprimer les thèmes highlight.js inutilisés:
```bash
cd web/lib
rm github-dark.min.css  # Si vous utilisez vs2015
```

3. Nettoyer les commentaires dans `index.html`

### Production avec Sélecteur de Thème

Pour permettre à l'utilisateur de choisir:

1. Créer un sélecteur de thème en JavaScript
2. Stocker la préférence dans `localStorage`
3. Charger dynamiquement le CSS au démarrage

Exemple:
```javascript
const theme = localStorage.getItem('theme') || 'vscode';
const themeCSS = theme === 'vscode' ? 'styles.css' : 'styles.demo.css';
const highlightCSS = theme === 'vscode' ? 'vs2015.min.css' : 'github-dark.min.css';
// Charger dynamiquement...
```

## 📝 Historique des Versions

| Version | Date | Thème | Highlight.js |
|---------|------|-------|--------------|
| v1-v5 | 2026-01-07 | Demo (inline puis externe) | github-dark |
| v6 | 2026-01-07 | VSCode (nouveau) | vs2015 |

## ✅ Checklist de Migration

- [x] Renommer `styles.css` → `styles.demo.css`
- [x] Créer `styles.css` avec thème VSCode
- [x] Télécharger `vs2015.min.css`
- [x] Modifier `index.html` pour utiliser vs2015
- [x] Incrémenter cache busting (v5 → v6)
- [x] Documenter les deux thèmes
- [x] Tester le thème VSCode
- [ ] Tester le thème Demo (switch manuel)

## 🔗 Ressources

- [VSCode CSS Variables](https://code.visualstudio.com/api/references/theme-color)
- [Highlight.js Themes](https://highlightjs.org/static/demo/)
- [WCAG Contrast Checker](https://webaim.org/resources/contrastchecker/)

---

**Statut**: ✅ Complété
**Thème actif**: VSCode (styles.css)
**Version**: v6
**Date**: 2026-01-07
