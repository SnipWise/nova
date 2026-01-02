## FACT-CHAUD-050: Vanne 3 voies défectueuse

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-050 |
| **Catégorie** | Hydraulique |
| **Système** | Vanne 3 voies / Inversion |
| **Gravité** | **Moyenne à Élevée** |
| **Marques** | Chaudières mixtes (Saunier Duval, Vaillant, Frisquet, Chaffoteaux) |

**Symptômes :**
- Pas d'eau chaude sanitaire ou ECS tiède malgré chaudière en chauffe
- Radiateurs ne chauffent plus alors que chaudière fonctionne
- Chaudière bascule en ECS mais radiateurs restent chauds (ou inverse)
- Fuite d'eau au niveau de la vanne 3 voies
- Bruit de moteur de vanne (grésillements) sans basculement

**Cause racine probable :**
Moteur de vanne 3 voies HS ou grippé, mécanisme de vanne bloqué par calcaire ou impuretés, joint de tige défectueux (fuite), connectique électrique défectueuse, micro-switch de position HS.

**Étapes de résolution :**

1. **Principe de fonctionnement de la vanne 3 voies**
   - La vanne 3 voies dirige le flux d'eau chaude soit vers le circuit chauffage, soit vers l'échangeur ECS
   - Commandée par un servomoteur électrique (moteur pas-à-pas ou moteur à came)
   - Temps de basculement : 15 à 60 secondes selon modèle
   - Positions : Chauffage / ECS / (parfois position intermédiaire)

2. **Diagnostic du problème**
   - **Test 1** : demander ECS (ouvrir robinet) → observer si basculement (bruit moteur vanne)
   - **Test 2** : toucher les tuyaux : tuyau ECS doit devenir chaud, tuyau chauffage doit refroidir
   - **Test 3** : observer le mouvement de la tige de vanne (tige mobile visible sur certains modèles)

3. **Vérification électrique du moteur de vanne**
   - Couper l'alimentation électrique
   - Accéder au moteur de la vanne 3 voies (démonter capot chaudière)
   - Repérer les connexions électriques (généralement 3 à 5 fils)
   - Remettre sous tension en mode test ECS
   - Mesurer la tension aux bornes du moteur (multimètre) : 230V AC attendu lors de la commande
   - **Si pas de tension** : problème de commande (carte électronique) ou micro-switch
   - **Si tension présente mais pas de mouvement** : moteur grippé ou HS

4. **Test du moteur de vanne (déblocage manuel)**
   - Certains moteurs ont une molette de déblocage manuel (rotation manuelle)
   - Tourner la molette pour basculer manuellement la vanne
   - Si rotation difficile ou impossible : mécanisme de vanne grippé
   - Si rotation facile mais pas de mouvement automatique : moteur HS

5. **Démontage et nettoyage du mécanisme de vanne**
   - Couper électricité, gaz, et fermer les vannes d'isolement
   - Vidanger partiellement le circuit primaire
   - Démonter le moteur de la vanne (2 à 4 vis)
   - Extraire la tige de commande (observer l'état : calcaire, encrassement ?)
   - Nettoyer la tige et le logement (brosse, vinaigre blanc si calcaire)
   - Graisser légèrement la tige avec graisse silicone résistante haute température
   - Actionner manuellement la vanne (tourner l'axe interne) : doit être fluide

6. **Contrôle des joints de tige**
   - Inspecter les joints toriques de la tige de vanne
   - Si joints durcis, déformés ou fuite visible : remplacer (kit joints par marque/modèle)
   - Remonter avec joints neufs graissés

7. **Remplacement du moteur de vanne**
   - Si le moteur est HS (bobinage coupé, engrenages cassés)
   - Commander le moteur de remplacement (référence exacte selon marque/modèle chaudière)
   - Démonter l'ancien moteur
   - Positionner le nouveau moteur sur la vanne (respect du calage : généralement position chauffage)
   - Visser le moteur sur son support
   - Reconnecter les fils électriques (respecter le code couleur/repérage)

8. **Remplacement complet de la vanne 3 voies**
   - Si le corps de vanne est fissuré, corrodé ou irréparable
   - Couper eau, gaz, électricité, vidanger
   - Démonter les raccords hydrauliques (3 raccords : chauffage, ECS, retour commun)
   - Extraire la vanne complète (avec moteur)
   - Installer la nouvelle vanne en respectant les positions (repérage des entrées/sorties)
   - Utiliser des joints neufs ou téflon sur les raccords filetés
   - Remonter le moteur sur la nouvelle vanne

9. **Remise en service et test**
   - Remplir le circuit, purger l'air
   - Remettre sous tension
   - Initialisation de la vanne (certaines chaudières effectuent un auto-test au démarrage)
   - Tester le mode chauffage : radiateurs doivent chauffer, ECS doit être coupée
   - Tester le mode ECS : ouvrir un robinet, observer le basculement (bruit moteur), ECS doit arriver chaude
   - Vérifier l'absence de fuite au niveau de la vanne

10. **Vérification du micro-switch de position**
    - Certains modèles ont un micro-switch qui détecte la position de la vanne
    - Si défectueux : la chaudière ne détecte pas la bonne position
    - Tester la continuité du micro-switch (multimètre) en actionnant manuellement
    - Remplacer si défectueux

**Prévention :**
- Actionner régulièrement la vanne (utiliser ECS même en été) pour éviter grippage
- Traiter l'eau du circuit primaire (inhibiteur) pour limiter les dépôts
- Vérifier le bon fonctionnement lors de l'entretien annuel

**Spécificités techniques :**
- Vanne 3 voies motorisée : temps de basculement 15-60 secondes
- Durée de vie moyenne : 10-15 ans
- Types de vannes : à tournant sphérique, à clapet, à secteur

**Avertissements sécurité :**
- Toujours couper l'alimentation électrique avant de démonter le moteur
- Vidanger le circuit avant de démonter la vanne (risque de brûlure et inondation)
- Respecter le sens de montage et le calage du moteur (risque de fonctionnement inversé)

---

**Fin du fichier 03_Hydraulique.md**
