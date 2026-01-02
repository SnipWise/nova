## FACT-CHAUD-080: Capteur pression différentielle défectueux

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-080 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Capteur ΔP |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières récentes, haute performance) |

**Symptômes :**
- Code erreur débit/pression différentielle
- Chaudière ne démarre pas (pas de débit détecté)
- Modulation puissance incohérente
- Régulation hydraulique défaillante
- Protection manque d'eau activée sans raison

**Cause racine probable :**
Capteur ΔP défectueux, prises de pression bouchées, membrane capteur encrassée, câblage défectueux, défaut carte électronique.

**Étapes de résolution :**

1. **Compréhension rôle capteur ΔP**
   - **ΔP = différence pression entre 2 points**
   - Mesure généralement : pression départ - pression retour
   - Fonction :
     - Détection circulation eau (débit)
     - Calcul débit volumique
     - Protection manque d'eau
     - Optimisation modulation selon débit
   - Remplace pressostat eau sur chaudières récentes

2. **Localisation capteur**
   - Généralement sur corps de chauffe
   - 2 tubes de prise pression (départ et retour)
   - Capteur électronique (signal 0-10V ou 4-20mA)
   - Connecté à carte électronique

3. **Diagnostic initial**
   - Consulter menu diagnostic chaudière
   - Relever valeur ΔP affichée
   - Comparer avec valeur attendue
   - Valeur typique fonctionnement : 100-300 mbar (varie selon modèle)
   - Pompe arrêtée : ΔP ≈ 0
   - Pompe en marche : ΔP doit augmenter

4. **Test fonctionnement pompe**
   - Vérifier pompe circulation fonctionne
   - Écouter bruit pompe
   - Vérifier rotation (rotor non bloqué)
   - Sans pompe : pas de ΔP (normal)

5. **Contrôle prises de pression**
   - Localiser 2 tubes prise pression (petits tubes 4-6mm)
   - Vérifier non bouchés :
     - Déconnecter du capteur
     - Souffler légèrement (doivent être dégagés)
     - Nettoyer si bouchés (aiguille fine, air comprimé)
   - Vérifier non pincés ou percés
   - Reconnecter solidement

6. **Contrôle membrane capteur**
   - Certains capteurs : membrane mesure ΔP
   - Membrane peut s'encrasser (tartre, boues)
   - Démonter capteur (couper pression, vidanger localement)
   - Nettoyer membrane délicatement (eau, brosse douce)
   - Rincer abondamment
   - Remonter avec joints neufs

7. **Test électrique capteur**
   - Mesurer alimentation capteur (multimètre)
   - Typique : 5V ou 12V DC
   - Mesurer signal sortie capteur (tension ou courant)
   - Exemple : 0-10V proportionnel à ΔP
   - Vérifier variation signal si on modifie débit (vitesse pompe)
   - Si signal fixe ou absent : capteur HS

8. **Contrôle câblage**
   - Vérifier connectique capteur (bien enfichée)
   - Contrôler continuité câbles (pas de coupure)
   - Vérifier absence court-circuit
   - Contrôler connexions carte électronique

9. **Test croisé**
   - Si possible, tester avec capteur identique (autre chaudière)
   - Ou simuler signal capteur (générateur tension/courant)
   - Si chaudière fonctionne avec autre capteur : capteur HS
   - Si problème persiste : carte électronique

10. **Remplacement capteur**
    - Identifier référence exacte (notice technique)
    - Commander pièce origine constructeur
    - Procédure remplacement :
      - Couper électricité et eau
      - Vidanger localement (circuit chaudière)
      - Déconnecter tubes pression
      - Déconnecter câble électrique
      - Dévisser capteur
      - Monter nouveau capteur (joints neufs, pâte d'étanchéité)
      - Reconnecter tubes et câble
      - Remplir et purger
      - Tester fonctionnement

11. **Paramétrage après remplacement**
    - Certains systèmes : calibration nécessaire
    - Procédure menu installateur
    - Vérifier valeur ΔP cohérente
    - Tester modulation puissance
    - Vérifier sécurités (arrêt si pompe HS)

12. **Contrôle installation hydraulique**
    - Vérifier pression circuit (1-1,5 bar)
    - Purger air (air = faux débit)
    - Contrôler filtre (pas bouché)
    - Vérifier vase expansion (pas de membrane percée)

**Prévention :**
- Vérification annuelle valeur ΔP (menu diagnostic)
- Nettoyage prises pression tous les 2-3 ans
- Traitement eau circuit (éviter encrassement)
- Contrôle câblage lors entretien

**Valeurs ΔP typiques :**
| Situation | ΔP attendu |
|-----------|------------|
| Pompe arrêtée | 0-20 mbar |
| Pompe vitesse 1 | 50-150 mbar |
| Pompe vitesse 2 | 100-250 mbar |
| Pompe vitesse 3 | 150-400 mbar |
| Filtre bouché | ΔP excessive (> 500 mbar) |

**Diagnostic selon valeur ΔP :**
- **ΔP = 0 en permanence :** capteur HS, pompe arrêtée, tubes bouchés
- **ΔP excessive :** filtre bouché, obstruction circuit, vanne fermée
- **ΔP erratique :** air dans circuit, capteur défectueux
- **ΔP ne varie pas :** capteur bloqué, membrane encrassée

**Codes erreur fréquents :**
- Vaillant F22 : manque d'eau (ΔP insuffisant)
- Saunier Duval F34 : défaut pression différentielle
- Frisquet : sécurité manque d'eau (ΔP)

**Fonction avancée capteur ΔP :**
- **Calcul débit :** Q = k × √ΔP (k = constante installation)
- **Optimisation modulation :** adapter puissance au débit
- **Détection anomalies :** filtre bouché, vanne fermée
- **Protection hydraulique :** arrêt si débit nul
- **Efficacité énergétique :** meilleur rendement

**Coût remplacement :**
- Capteur ΔP : 80-200€ selon modèle
- Main d'œuvre : 1-2h
- Total intervention : 150-400€

**Alternative si panne :**
- Certaines chaudières : mode dégradé sans capteur ΔP
- Ou by-pass temporaire (shunt sécurité, avec précautions)
- **DANGER :** ne pas neutraliser sécurité définitivement (risque surchauffe)

---

