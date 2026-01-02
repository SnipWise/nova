## FACT-CHAUD-042: Débit insuffisant, radiateurs froids

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-042 |
| **Catégorie** | Hydraulique |
| **Système** | Circuit chauffage / Distribution |
| **Gravité** | **Moyenne** |
| **Marques** | Multi-marques (tous systèmes) |

**Symptômes :**
- Radiateurs tièdes ou froids, surtout les plus éloignés de la chaudière
- Chauffage lent à monter en température
- Température départ chaudière correcte mais retour froid
- Certains radiateurs chauffent bien, d'autres non (déséquilibre)

**Cause racine probable :**
Circulateur sous-dimensionné ou vitesse trop faible, embouage du circuit, filtre encrassé, vannes radiateurs fermées ou grippées, air dans le circuit, by-pass hydraulique mal réglé, pertes de charge excessives.

**Étapes de résolution :**

1. **Vérification de la température de départ/retour**
   - Relever la température de départ chaudière (affichage ou thermomètre)
   - Mesurer la température de retour (tuyau retour avant chaudière)
   - Écart normal : 15-20°C (par ex : départ 70°C, retour 50-55°C)
   - Si écart > 25°C : débit trop faible, mauvaise circulation

2. **Contrôle de la vitesse du circulateur**
   - Vérifier le réglage de vitesse sur le circulateur (sélecteur 1-2-3 ou Auto)
   - Passer à la vitesse supérieure si débit insuffisant
   - Observer si l'amélioration est sensible (radiateurs plus chauds)
   - Attention : augmenter la vitesse augmente aussi la consommation électrique et le bruit

3. **Purge complète du circuit**
   - Purger méthodiquement tous les radiateurs en commençant par les plus proches
   - Maintenir la pression du circuit pendant la purge (rajouter eau si nécessaire)
   - Purger également le circulateur (voir FACT-CHAUD-039)
   - Purger les points hauts du circuit (purgeurs automatiques s'ils existent)

4. **Vérification de l'ouverture des vannes radiateurs**
   - Contrôler que tous les robinets thermostatiques ou manuels sont bien ouverts
   - Démonter la tête thermostatique : vérifier que le pointeau n'est pas grippé en position fermée
   - Actionner manuellement le pointeau (appuyer dessus, il doit revenir par ressort)
   - Si grippé : débloquer avec pince multiprise ou WD-40, actionner plusieurs fois

5. **Contrôle du filtre à boues**
   - Localiser le filtre (généralement en amont de la chaudière ou du circulateur)
   - Fermer les vannes d'isolement du filtre
   - Dévisser le corps du filtre, retirer la cartouche
   - Nettoyer la cartouche sous l'eau (brosse si très encrassée)
   - Remonter, ouvrir les vannes, purger l'air, tester

6. **Vérification de l'embouage du circuit**
   - Observer la couleur de l'eau lors d'une purge : doit être claire
   - Si eau noire, marron ou rouge : circuit embué (boues magnétite)
   - Test : placer un aimant sur un radiateur froid, s'il colle fortement = boues magnétiques
   - **Solution** : désembouage complet du circuit (voir FACT-CHAUD-043)

7. **Contrôle du by-pass hydraulique**
   - Certains circuits ont un by-pass entre départ et retour (vanne réglable ou automatique)
   - Si by-pass trop ouvert : une partie de l'eau chauffe contourne les radiateurs
   - Fermer partiellement le by-pass et observer l'amélioration
   - Réglage optimal : débit juste suffisant pour éviter le blocage du circulateur vannes fermées

8. **Équilibrage hydraulique des radiateurs**
   - Les radiateurs éloignés peuvent manquer de débit si pas d'équilibrage
   - Technique : réduire légèrement le débit des radiateurs proches (vis de réglage retour)
   - Augmenter le débit des radiateurs éloignés (ouvrir complètement)
   - Procéder par itérations jusqu'à obtenir une température homogène

9. **Vérification du dimensionnement du circulateur**
   - Calculer la hauteur manométrique nécessaire (pertes de charge circuit)
   - Comparer avec les performances du circulateur installé (voir courbe constructeur)
   - Si circulateur sous-dimensionné : **remplacement par modèle plus puissant**

10. **Contrôle des tuyauteries (obstruction)**
    - Si un seul radiateur reste froid : obstruction locale probable
    - Démonter le radiateur, rincer à contre-courant au tuyau d'arrosage
    - Vérifier l'absence d'obstruction sur les vannes et té de dérivation

**Prévention :**
- Installer un filtre magnétique pour capturer les boues
- Ajouter un inhibiteur de corrosion dans le circuit
- Effectuer un désembouage préventif tous les 5-7 ans
- Équilibrer le circuit lors de l'installation ou modification

**Spécificités techniques :**
- Formule de débit : Q (m³/h) = P (kW) / (1,16 × ΔT) où ΔT est l'écart départ/retour
- Exemple : 15 kW avec ΔT 20°C → Q = 15/(1,16×20) = 0,65 m³/h = 650 L/h

---

