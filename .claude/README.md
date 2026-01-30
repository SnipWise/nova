# Configuration Claude Code - Méthodologie de Développement

Ce dossier contient une skill et un agent implémentant la méthodologie de développement contrôlé décrite dans [`docs/methodologie-claude-code.md`](../docs/methodologie-claude-code.md).

## Structure

```
.claude/
├── README.md                      # Ce fichier
├── skills/
│   └── methodical-dev.md         # Skill de développement méthodique
└── agents/
    └── methodical-dev.md         # Agent de développement méthodique
```

## 1. Skill : `/methodical-dev`

### Qu'est-ce que c'est ?

Une **skill** est une commande interactive qui guide l'utilisateur étape par étape. Elle pose des questions, crée un plan, et implémente avec validation à chaque étape.

### Caractéristiques

- 🎯 **Interactive** : Pose des questions pour clarifier les besoins
- 📋 **Planification** : Crée un plan détaillé avant de coder
- ✋ **Validation systématique** : S'arrête après chaque étape
- 🎮 **Contrôle total** : L'utilisateur valide chaque action

### Comment l'utiliser ?

#### Méthode 1 : Invocation directe

```bash
# Dans Claude Code, simplement taper :
/methodical-dev
```

La skill va alors vous poser des questions pour :
1. Comprendre votre objectif
2. Identifier les contraintes techniques
3. Collecter la documentation nécessaire
4. Établir les conventions à respecter

#### Méthode 2 : Avec contexte initial

Si vous avez déjà les informations, vous pouvez les fournir directement :

```
/methodical-dev

Je veux créer une feature d'authentification JWT pour mon API Express.

Contraintes :
- Utiliser jsonwebtoken et bcrypt
- Suivre les patterns de src/auth/
- TypeScript strict
- Tests avec Jest

Documentation : https://jwt.io/introduction
```

### Workflow de la Skill

```
1. COLLECTE D'INFORMATION
   ↓ (questions interactives)

2. VÉRIFICATION GIT
   ↓ (propose création de branche)

3. PLANIFICATION
   ↓ (analyse + plan détaillé)
   ↓ (validation du plan)

4. IMPLÉMENTATION
   ↓ (étape 1) → VALIDATION
   ↓ (étape 2) → VALIDATION
   ↓ (étape N) → VALIDATION

5. RÉSUMÉ FINAL
   ↓ (checklist + proposition commit)
```

### Exemple de Session

```
Vous : /methodical-dev

Claude : Je vais vous guider à travers un développement méthodique.

=== PHASE 1 : COLLECTE D'INFORMATION ===

[Série de questions interactives]

Vous : [Réponses]

Claude : === PHASE 3 : PLANIFICATION ===

Voici le plan proposé :

□ Étape 1 : Créer les types TypeScript pour JWT
□ Étape 2 : Implémenter la génération de tokens
□ Étape 3 : Implémenter la validation de tokens
□ Étape 4 : Créer les middleware Express
□ Étape 5 : Ajouter les tests unitaires
□ Étape 6 : Intégrer dans les routes existantes

Ce plan vous convient-il ?

Vous : Oui, c'est bon

Claude : === ÉTAPE 1 : Créer les types TypeScript ===

📋 Ce que je vais faire :
- Créer src/auth/types.ts avec interfaces JWTPayload, AuthToken
- Définir les types pour les configs JWT

[Implémentation...]

📊 RÉSUMÉ ÉTAPE 1 :
✅ Créé : src/auth/types.ts
✅ Fonctionnalité : Types complets pour JWT

⏸️  POINT DE CONTRÔLE
Veuillez vérifier avant de continuer.

Vous : OK continue

[...et ainsi de suite pour chaque étape...]
```

### Avantages de la Skill

✅ Guidance interactive
✅ Validation à chaque étape
✅ Évite les dérives et raccourcis
✅ Transparence totale
✅ Contrôle utilisateur maximal

## 2. Agent : `methodical-developer`

### Qu'est-ce que c'est ?

Un **agent** est un processus autonome qui exécute une tâche complexe avec supervision. Il est plus autonome qu'une skill mais s'arrête toujours pour validation.

### Caractéristiques

- 🤖 **Autonome supervisé** : Prend des décisions mais demande validation
- 🔄 **Multi-étapes** : Gère des workflows complexes
- 🛡️ **Garde-fous** : Interdictions strictes sur suppressions/modifications
- 📊 **Rapports détaillés** : Génère des rapports après chaque étape

### Comment l'utiliser ?

#### Via le Task tool

```bash
# Demander à Claude Code de lancer l'agent
"Lance l'agent methodical-developer pour créer une feature de cache Redis"
```

Claude Code va utiliser le Task tool en interne :

```python
Task(
    subagent_type="feature-dev:methodical-developer",
    prompt="Développe une feature de cache Redis pour l'API. Suis les patterns de src/database/. Utilise ioredis.",
    description="Cache Redis avec méthodologie"
)
```

#### Exemple complet

```
Vous : Je veux développer une feature complète d'export de données en CSV et JSON.
Utilise l'agent methodical-developer pour le faire de manière structurée.

Claude Code : Je vais lancer l'agent methodical-developer pour cette tâche.

[Lance l'agent...]

Agent : === ANALYSE INITIALE ===

J'ai analysé le code existant et identifié :
- Utilisation de Express avec TypeScript
- Pattern de services dans src/services/
- Pattern de routes dans src/routes/

Questions de clarification :
1. Format des données source ? (depuis DB ? API ? fichiers ?)
2. Options d'export requises ? (filtres, tri, pagination ?)
3. Génération synchrone ou async (files de jobs) ?

[Vous répondez...]

Agent : === PLAN PROPOSÉ ===

Je propose de découper en 6 étapes :

1. Créer les types et interfaces pour l'export
2. Implémenter le service d'export CSV
3. Implémenter le service d'export JSON
4. Créer les routes API avec validation
5. Ajouter les tests unitaires et d'intégration
6. Documenter l'API

Validation du plan ?

Vous : OK

Agent : === ÉTAPE 1 : Types et interfaces ===

📋 Actions :
- src/export/types.ts (interfaces ExportOptions, ExportResult, ExportFormat)
- src/export/config.ts (configuration export)

[Implémentation...]

📊 RAPPORT ÉTAPE 1 :

✅ RÉALISATIONS :
- Créé : src/export/types.ts (85 lignes)
  - Interface ExportOptions avec filtres, pagination, sorting
  - Type ExportFormat = 'csv' | 'json'
  - Interface ExportResult avec data, metadata

- Créé : src/export/config.ts (30 lignes)
  - Limites d'export (max 10000 rows)
  - Options par défaut

🔧 CHOIX TECHNIQUES :
- Utilisation de zod pour validation (cohérent avec le reste du code)
- Support de streams pour gros volumes (via options.stream)

⏸️  VALIDATION REQUISE

Vous : Parfait, continue

[...suite...]
```

### Workflow de l'Agent

```
1. ANALYSE
   - Lit le code existant
   - Identifie les patterns
   - Pose questions de clarification

2. PLANIFICATION
   - Génère plan détaillé
   - Demande validation

3. IMPLÉMENTATION (boucle)
   Pour chaque étape :
   ├─ Marque "in_progress"
   ├─ Implémente
   ├─ Génère rapport détaillé
   ├─ STOP et attend validation
   └─ Si OK, continue

4. VALIDATION FINALE
   - Résumé complet
   - Checklist qualité
   - Proposition commit
```

### Commandes Pendant l'Exécution

Vous pouvez interagir avec l'agent :

```bash
/pause          # Pause pour révision
/show-plan      # Affiche plan et progression
/rollback       # Annule dernière étape
/modify-plan    # Ajuste le plan restant
/report         # Rapport détaillé complet
/resume         # Reprend après pause
```

### Avantages de l'Agent

✅ Plus autonome que la skill
✅ Gère la complexité (exploration, analyse)
✅ Rapports détaillés automatiques
✅ Garde-fous intégrés
✅ Toujours supervision utilisateur

## Comparaison Skill vs Agent

| Critère | Skill `/methodical-dev` | Agent `methodical-developer` |
|---------|------------------------|----------------------------|
| **Complexité** | Features simples/moyennes | Features complexes |
| **Autonomie** | Guidée (vous pilotez) | Supervisée (agent pilote) |
| **Interaction** | Questions fréquentes | Questions au début + validations |
| **Exploration** | Limitée | Approfondie (analyse code) |
| **Rapidité** | Plus rapide (moins d'analyse) | Plus lent (analyse complète) |
| **Contrôle** | Maximal | Élevé |
| **Cas d'usage** | Features claires | Features nécessitant analyse |

## Quand Utiliser Quoi ?

### Utilisez la Skill `/methodical-dev` si :

- ✅ Vous savez exactement ce que vous voulez
- ✅ La feature est relativement simple (3-5 étapes)
- ✅ Vous voulez un maximum de contrôle interactif
- ✅ Vous avez déjà analysé le code existant

**Exemple** : "Ajouter un endpoint GET /users/:id à mon API Express existante"

### Utilisez l'Agent `methodical-developer` si :

- ✅ La feature est complexe (6+ étapes)
- ✅ Vous avez besoin d'analyse du code existant
- ✅ Vous voulez des rapports détaillés automatiques
- ✅ La feature touche plusieurs parties du code

**Exemple** : "Implémenter un système complet d'authentification avec JWT, refresh tokens, et gestion des sessions"

## Principes Communs

Que ce soit la skill ou l'agent, ils partagent ces principes :

### ❌ Ce qu'ils ne feront JAMAIS :

- Supprimer du code sans confirmation
- Prendre des raccourcis non demandés
- Modifier l'architecture sans accord
- Commiter automatiquement
- Contourner un problème en supprimant
- Implémenter différemment "pour simplifier"

### ✅ Ce qu'ils feront TOUJOURS :

- S'arrêter après chaque étape pour validation
- Expliquer leurs choix techniques
- Respecter strictement le plan validé
- Rapporter tout changement
- Demander confirmation pour décisions importantes
- Privilégier le code correct sur le code qui compile

## Configuration Avancée

### Customiser la Skill

Vous pouvez modifier [`.claude/skills/methodical-dev.md`](skills/methodical-dev.md) pour :
- Ajouter des questions spécifiques à votre projet
- Adapter les checklist de qualité
- Modifier les étapes du workflow

### Customiser l'Agent

Vous pouvez créer des variantes de l'agent avec des settings différents :

```json
// .claude/agents/methodical-dev-fast.json
{
  "name": "methodical-developer-fast",
  "type": "feature-dev",
  "settings": {
    "require_validation_after_each_step": false,
    "max_steps_without_validation": 3,
    "generate_step_reports": true
  }
}
```

### Hooks

Vous pouvez configurer des hooks pour automatiser des vérifications :

```json
// .claude/hooks.json
{
  "hooks": {
    "before-step": {
      "command": "npm run type-check",
      "on_failure": "notify"
    },
    "after-step": {
      "command": "npm test -- --bail",
      "on_failure": "stop"
    }
  }
}
```

## Troubleshooting

### La skill ne s'active pas

Vérifier que le fichier est bien dans `.claude/skills/methodical-dev.md` et que Claude Code a redémarré.

### L'agent ne s'arrête pas après chaque étape

L'agent devrait s'arrêter. Si ce n'est pas le cas, rappeler explicitement :
```
"Arrête-toi maintenant et attend ma validation avant de continuer"
```

### Comment revenir en arrière ?

Si Git est utilisé :
```bash
git status
git diff
git restore <fichier>  # pour annuler un fichier
```

### L'agent est trop verbeux

Vous pouvez demander :
```
"Fait des rapports plus concis, juste l'essentiel"
```

### L'agent prend trop de temps

C'est normal avec la méthodologie rigoureuse. Si vous préférez la vitesse :
- Utilisez Claude Code standard (sans cette méthodologie)
- Mais acceptez les risques de dérives

## Ressources

- 📄 [Méthodologie complète](../docs/methodologie-claude-code.md)
- 📊 [Slides de présentation](../docs/slides/)
- 🎯 [Code source de la skill](skills/methodical-dev.md)
- 🤖 [Configuration de l'agent](agents/methodical-dev.md)

## Feedback et Améliorations

Cette méthodologie et ces outils sont en évolution. N'hésitez pas à :
- Adapter à vos besoins spécifiques
- Créer des variantes pour votre équipe
- Partager vos retours d'expérience

## Philosophie

> "Un bon développeur avec un bon assistant bat toujours un assistant seul."

Cette méthodologie vise à combiner :
- L'efficacité de l'IA
- Le jugement humain
- La rigueur du processus

Pour obtenir la meilleure qualité de code possible.

---

**Version** : 1.0.0
**Dernière mise à jour** : 2026-01-29
