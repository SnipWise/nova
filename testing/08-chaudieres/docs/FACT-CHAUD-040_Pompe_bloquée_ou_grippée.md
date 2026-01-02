## FACT-CHAUD-040: Pompe bloquée ou grippée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-040 |
| **Catégorie** | Hydraulique |
| **Système** | Circulateur / Pompe |
| **Gravité** | **Élevée** |
| **Marques** | Grundfos, Wilo, DAB, Salmson (tous circulateurs) |

**Symptômes :**
- Circulateur ne tourne pas, aucun bruit de fonctionnement
- Le corps du circulateur est chaud (surchauffe du moteur)
- Radiateurs restent froids malgré chaudière en chauffe
- Voyant de défaut ou erreur pompe sur la chaudière
- Disjonction électrique ou fusible grillé

**Cause racine probable :**
Rotor grippé par manque d'utilisation (arrêt prolongé), dépôts calcaires ou oxydes métalliques bloquant l'axe, roulements grippés, corps étranger dans la pompe, bobinage moteur HS.

**Étapes de résolution :**

1. **Vérification de l'alimentation électrique**
   - Contrôler que le circulateur est bien alimenté (multimètre : 230V)
   - Vérifier l'état du fusible ou disjoncteur dédié
   - Tester la continuité des connexions électriques

2. **Diagnostic du blocage mécanique**
   - Couper l'alimentation électrique (sécurité)
   - Poser la main sur le corps du circulateur : s'il est très chaud, le moteur a tenté de tourner sans y parvenir (blocage mécanique)
   - Écouter : aucun bruit = blocage total ou défaut électrique

3. **Déblocage manuel du rotor**
   - Maintenir l'alimentation coupée
   - Dévisser la vis centrale du capot avant du circulateur (vis de 5-8 mm)
   - Retirer le capot avec précaution (attention : un peu d'eau peut s'écouler)
   - Observer l'axe du rotor au centre (fente pour tournevis plat)
   - Insérer un tournevis plat dans la fente de l'axe
   - Tourner manuellement dans les deux sens pour débloquer
   - **Astuce** : si très dur, appliquer quelques gouttes de dégrippant (WD-40) autour de l'axe, attendre 10 min, réessayer

4. **Vérification de la rotation libre**
   - Une fois débloqué, l'axe doit tourner librement à la main (360° sans résistance)
   - Si rotation toujours difficile ou par à-coups : roulements usés ou encrassement important

5. **Nettoyage interne (si encrassement)**
   - Vidanger le circuit ou fermer les vannes d'isolement du circulateur
   - Déposer complètement le circulateur
   - Démonter le corps (selon modèle) pour accéder au rotor et à la volute
   - Nettoyer les dépôts (calcaire, boues, oxydes) avec brosse douce et vinaigre blanc
   - Rincer abondamment à l'eau claire
   - Remonter avec joints neufs si disponibles

6. **Test de remise en service**
   - Remonter le capot du circulateur
   - Remettre sous tension électrique
   - Observer si le circulateur démarre (légère vibration, bruit de fonctionnement)
   - Purger le circulateur pour évacuer l'air introduit
   - Vérifier que les radiateurs chauffent

7. **Si le déblocage échoue : test du bobinage moteur**
   - Couper l'alimentation, débrancher les fils du circulateur
   - Mesurer la résistance du bobinage avec un multimètre (ohmmètre)
   - Valeur normale : 50-500 ohms selon modèle (vérifier notice)
   - Si résistance infinie (∞) : bobinage coupé → **remplacement obligatoire**
   - Si résistance nulle (0) : court-circuit → **remplacement obligatoire**
   - Tester l'isolement moteur/masse : doit être > 1 MΩ

8. **Remplacement du circulateur**
   - Si le déblocage est impossible ou le moteur est HS
   - Suivre la procédure décrite dans FACT-CHAUD-039 étape 8
   - Choisir un circulateur de performances équivalentes (hauteur manométrique, débit)

**Prévention :**
- Faire fonctionner le circulateur au moins 1 fois par mois hors saison de chauffe (évite grippage)
- Utiliser un inhibiteur de corrosion dans le circuit
- Entretenir le circuit (désembouage si nécessaire)
- Installer un filtre à boues en amont du circulateur

**Spécificités techniques :**
- Les circulateurs modernes à aimants permanents (Grundfos Alpha, Wilo Stratos) sont moins sujets au grippage
- Fonction anti-blocage sur certains circulateurs électroniques (rotation automatique périodique)

**Avertissements sécurité :**
- Toujours couper l'alimentation avant d'ouvrir le capot du circulateur (risque électrique)
- Ne jamais faire fonctionner un circulateur bloqué plus de quelques minutes (destruction du moteur par surchauffe)

---

