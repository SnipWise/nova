## FACT-CHAUD-087: Sécurité manque d'eau - Débitmètre ou pressostat

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-087 |
| **Catégorie** | Sécurité Gaz & Sécurités Générales |
| **Système** | Détection manque eau / débit |
| **Gravité** | **Critique** |
| **Marques** | Multi-marques (surtout instantanés ECS) |

**Symptômes :**
- Blocage chaudière sur appel ECS (eau chaude sanitaire)
- Code erreur "débit insuffisant" ou "manque d'eau ECS"
- Brûleur ne démarre pas à l'ouverture robinet
- Ou extinction brutale 10-30 secondes après démarrage
- Température ECS monte très vite (surchauffe)

**Cause racine probable :**
Débit eau sanitaire insuffisant (< 1,5-2 L/min), débitmètre encrassé ou défectueux, filtre eau bouché, pression eau réseau basse, échangeur ECS bouché (tartre).

**Étapes de résolution :**

1. **Comprendre sécurité manque d'eau**
   - **Chaudière ECS instantané** : chauffe eau à la demande
   - **Débit mini requis** : 1,5-2,5 L/min selon modèle
   - **Capteur** : débitmètre (turbine) ou pressostat différentiel
   - **Rôle** : autoriser allumage uniquement si débit suffisant
   - **Protection** : évite surchauffe échangeur ECS (destruction)

2. **Mesure débit eau sanitaire**
   - Ouvrir robinet eau chaude à fond
   - Remplir récipient gradué (1 L) avec chronomètre
   - Calculer débit : volume (L) / temps (min) = L/min
   - **Débit correct** : > 2 L/min
   - **Débit faible** : < 1,5 L/min (blocage attendu)
   - Tester sur plusieurs points puisage

3. **Contrôle pression eau réseau**
   - Mesurer pression eau froide (manomètre)
   - Pression normale : 2,5-4 bar
   - Si < 2 bar : peut causer débit insuffisant
   - Contacter distributeur eau si pression réseau basse
   - Installer surpresseur si nécessaire (habitat isolé)

4. **Contrôle filtres et crépine**
   - Localiser filtre amont chaudière (entrée eau froide)
   - Fermer vanne amont, vidanger
   - Démonter et nettoyer filtre (grillage)
   - Vérifier tamis crépine (pas bouché)
   - Nettoyer ou remplacer cartouche
   - Réduction débit de 50% fréquente si filtre colmaté

5. **Contrôle débitmètre (turbine)**
   - Localiser débitmètre ECS (entrée eau froide sanitaire)
   - Type : turbine à ailettes (génère impulsions électriques)
   - **Nettoyage** :
     * Fermer vannes ECS, vidanger
     * Démonter débitmètre (2 vis généralement)
     * Extraire turbine délicatement
     * Nettoyer ailettes (brosse douce, vinaigre blanc)
     * Éliminer calcaire, impuretés
     * Vérifier rotation libre (souffle ou eau)
     * Remonter avec joint neuf
   - **Test électrique** :
     * Ouvrir robinet, vérifier rotation turbine
     * Mesurer signal sortie (impulsions 0-5V ou contact)
     * Remplacer débitmètre si pas de signal

6. **Contrôle pressostat différentiel**
   - Sur certains modèles : pressostat ΔP (pas turbine)
   - Mesure différence pression entrée/sortie échangeur
   - Vérifier tubes prises pression (pas bouchés)
   - Nettoyer tubes (soufflette)
   - Tester commutation pressostat (continuité)
   - Remplacer si défectueux

7. **Contrôle échangeur ECS (tartre)**
   - Échangeur à plaques bouché = débit réduit
   - Symptômes :
     * Débit correct en eau froide
     * Débit faible en eau chaude (contre-pression)
     * ΔT très élevé (> 30°C)
   - **Détartrage** :
     * Isoler échangeur (vannes)
     * Détartrage chimique (pompe + produit)
     * Ou échangeur démonté (bain acide)
     * Rinçage abondant
   - Zones dureté eau > 25°F : détartrage annuel

8. **Contrôle réducteur pression**
   - Réducteur pression amont défectueux :
     * Pression sortie trop basse
     * Ou fluctuations pression
   - Vérifier réglage réducteur (2,5-3 bar sortie)
   - Remplacer si défaillant

9. **Vérification installation**
   - Tuyauterie ECS trop étroite (Ø < 12 mm) : perte charge
   - Vannes 1/4 tour (pas robinet ancien) : débit limité
   - Nombre coudes excessif : pertes charge
   - Calcul dimensionnement si installation défectueuse

**Prévention :**
- Nettoyage filtre et débitmètre annuel
- Détartrage échangeur ECS (si eau dure > 20°F)
- Contrôle pression réseau
- Adoucisseur d'eau si dureté > 30°F
- Vérification débit à chaque entretien

**Débit ECS requis selon chaudière :**
| Puissance ECS | Débit mini | Température ΔT |
|---------------|-----------|----------------|
| 18-20 kW | 1,5 L/min | +40°C |
| 24-28 kW | 2,0 L/min | +40°C |
| 30-35 kW | 2,5 L/min | +40°C |

**Dureté eau et fréquence détartrage :**
- **Eau douce** (< 15°F) : tous les 5 ans
- **Eau moyennement dure** (15-25°F) : tous les 3 ans
- **Eau dure** (25-35°F) : tous les 2 ans
- **Eau très dure** (> 35°F) : annuel + adoucisseur recommandé

**Signes entartrage échangeur ECS :**
- Débit ECS diminue progressivement
- Température ECS instable
- Bruits dans échangeur (claquements)
- ΔT excessif eau chaude
- Blocages fréquents manque d'eau

**⚠️ SÉCURITÉ :**
- Sécurité manque d'eau = PROTECTION VITALE
- Fonctionnement sans eau = destruction échangeur (quelques secondes)
- Ne JAMAIS shunter débitmètre ou pressostat
- Surchauffe ECS = risque brûlure grave (> 80°C possible)
- Risque explosion vapeur si manque eau prolongé

**Solutions si problème récurrent :**
- Installation adoucisseur d'eau (TH < 15°F)
- Filtre anti-tartre magnétique/polyphosphate
- Remplacement échangeur ECS si très entartré
- Upgrade chaudière mixte vers ballon ECS (moins sensible)

---

