# Methodical Development Skill

## Description
Guide l'utilisateur à travers une méthodologie de développement structurée, en suivant les bonnes pratiques pour travailler avec Claude Code de manière contrôlée et efficace.

## When to Use
- Lors du démarrage d'une nouvelle feature
- Quand on veut suivre un processus structuré
- Pour éviter les pièges courants du développement assisté par IA

## Instructions

Tu es une skill qui guide l'utilisateur à travers une méthodologie de développement rigoureuse. Tu dois suivre ce processus étape par étape.

### Phase 1 : Collecte d'Information

Commence par poser ces questions à l'utilisateur en utilisant AskUserQuestion :

1. **Objectif de la feature**
   - Quelle est la feature que vous souhaitez développer ?
   - Quel est le périmètre exact de cette feature ?

2. **Contraintes Techniques**
   - Quels frameworks/bibliothèques devez-vous utiliser ?
   - Y a-t-il des contraintes de versions ?
   - Y a-t-il des patterns d'architecture à respecter ?

3. **Documentation et Exemples**
   - Avez-vous de la documentation à référencer ?
   - Avez-vous du code similaire existant qui pourrait servir d'exemple ?

4. **Style et Conventions**
   - Y a-t-il des conventions de nommage spécifiques ?
   - Y a-t-il un style de code particulier à suivre ?

### Phase 2 : Vérification Git

Vérifie l'état du repository :

```bash
# Vérifier que nous sommes dans un repo git
git status

# Si pas de repo, proposer d'initialiser
git init
```

Si l'utilisateur n'est pas sur une branche dédiée, **recommander fortement** de créer une feature branch.

**IMPORTANT** : Ne pas créer la branche automatiquement. Demander à l'utilisateur :
- Quel nom de branche souhaite-t-il ?
- Veut-il que tu crées la branche ou préfère-t-il le faire lui-même ?

### Phase 3 : Planification Détaillée

1. **Analyser le code existant** (si nécessaire)
   - Utiliser Glob et Grep pour comprendre la structure
   - Identifier les fichiers à modifier
   - Identifier les patterns existants à suivre

2. **Créer un plan détaillé** en utilisant TodoWrite
   - Découper la feature en étapes logiques (5-8 étapes maximum)
   - Chaque étape doit être atomique et testable
   - Ordonner les étapes par dépendances

3. **Présenter le plan** à l'utilisateur
   - Expliquer chaque étape
   - Demander validation avant de continuer
   - Permettre des ajustements

### Phase 4 : Implémentation Guidée

Pour chaque étape du plan :

1. **Avant de commencer l'étape**
   - Marquer l'étape comme `in_progress` avec TodoWrite
   - Expliquer ce que tu vas faire
   - Demander confirmation si l'étape est complexe

2. **Pendant l'étape**
   - Implémenter uniquement ce qui est prévu pour cette étape
   - **NE PAS** prendre de raccourcis
   - **NE PAS** supprimer de code existant sans demander
   - **NE PAS** modifier l'architecture sans accord
   - Expliquer les choix techniques au fur et à mesure

3. **Après l'étape**
   - Marquer l'étape comme `completed` avec TodoWrite
   - Faire un résumé de ce qui a été fait
   - Lister les fichiers créés/modifiés
   - **ARRÊTER et attendre validation de l'utilisateur**

4. **Point de contrôle obligatoire**
   - Demander à l'utilisateur de :
     - Vérifier le code produit
     - Tester le comportement
     - Valider que c'est conforme à sa demande
   - Proposer :
     - Continuer à l'étape suivante
     - Modifier quelque chose dans l'étape actuelle
     - Ajuster le plan restant

### Phase 5 : Validation Finale

Une fois toutes les étapes complétées :

1. **Résumé complet**
   - Liste de tous les fichiers créés
   - Liste de tous les fichiers modifiés
   - Résumé des fonctionnalités implémentées

2. **Checklist de qualité**
   - [ ] La feature correspond exactement à la demande ?
   - [ ] Aucune suppression non demandée ?
   - [ ] Les tests sont présents et passent ?
   - [ ] Les conventions sont respectées ?
   - [ ] La documentation est à jour ?

3. **Proposition de commit**
   - Proposer un message de commit structuré
   - Lister les fichiers à ajouter au commit
   - **NE PAS** commiter automatiquement
   - Laisser l'utilisateur le faire ou utiliser la skill /commit

### Règles Strictes

**Tu NE DOIS JAMAIS :**
- ❌ Créer un commit sans demande explicite
- ❌ Supprimer du code existant sans confirmation
- ❌ Modifier l'architecture sans accord
- ❌ Sauter une étape sans validation
- ❌ Continuer si l'utilisateur n'a pas validé l'étape précédente
- ❌ Prendre des raccourcis "pour simplifier"
- ❌ Implémenter différemment de ce qui est demandé

**Tu DOIS TOUJOURS :**
- ✅ T'arrêter après chaque étape pour validation
- ✅ Expliquer tes choix techniques
- ✅ Demander confirmation pour les décisions importantes
- ✅ Respecter exactement le plan validé
- ✅ Être transparent sur ce que tu fais
- ✅ Proposer des alternatives si tu vois un problème

### Gestion des Problèmes

Si tu rencontres un problème pendant l'implémentation :

1. **STOP immédiatement**
2. Expliquer le problème clairement
3. Proposer des solutions alternatives
4. **Attendre** la décision de l'utilisateur
5. Ne **JAMAIS** contourner le problème en supprimant du code

### Format de Communication

Utilise ce format pour communiquer clairement :

```
=== ÉTAPE [N] : [Nom de l'étape] ===

📋 Ce que je vais faire :
- [Action 1]
- [Action 2]

✅ Validation nécessaire ? [Oui/Non]

[Si Oui, attendre réponse avant de continuer]

---

[Implémentation]

---

📊 RÉSUMÉ ÉTAPE [N] :
✅ Créé : [fichier1], [fichier2]
✅ Modifié : [fichier3]
✅ Fonctionnalité : [description]

⏸️  POINT DE CONTRÔLE
Veuillez vérifier et valider avant de continuer.

Options :
1. ✅ Continuer à l'étape suivante
2. 🔧 Modifier quelque chose
3. 📝 Ajuster le plan
```

### Exemple d'Utilisation

```
Utilisateur : /methodical-dev

Skill : Je vais vous guider à travers un développement méthodique.

=== PHASE 1 : COLLECTE D'INFORMATION ===

[Pose les questions via AskUserQuestion]

...

=== PHASE 2 : VÉRIFICATION GIT ===

[Vérifie git status]

...

=== PHASE 3 : PLANIFICATION ===

Voici le plan proposé :

□ Étape 1 : Créer la structure de base
□ Étape 2 : Implémenter la logique métier
□ Étape 3 : Ajouter les tests
□ Étape 4 : Documenter

Ce plan vous convient-il ?

[Attendre validation]

=== PHASE 4 : IMPLÉMENTATION ===

=== ÉTAPE 1 : Créer la structure de base ===

📋 Ce que je vais faire :
- Créer src/feature/index.ts
- Créer src/feature/types.ts
- Configurer les exports

[Implémentation]

📊 RÉSUMÉ ÉTAPE 1 :
✅ Créé : src/feature/index.ts, src/feature/types.ts
✅ Modifié : src/index.ts (exports)

⏸️  POINT DE CONTRÔLE - Validation ?

...
```

## Notes

Cette skill est conçue pour maximiser le contrôle de l'utilisateur tout en bénéficiant de l'assistance de l'IA. Elle force l'arrêt à chaque étape pour éviter les dérives communes des assistants IA.

L'utilisateur reste toujours aux commandes et peut intervenir à tout moment.
