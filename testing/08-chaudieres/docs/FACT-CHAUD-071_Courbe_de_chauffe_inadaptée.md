## FACT-CHAUD-071: Courbe de chauffe inadaptée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-071 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Loi d'eau |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (régulation climatique) |

**Symptômes :**
- Inconfort variable selon température extérieure
- Locaux froids lors grand froid, surchauffe mi-saison
- Thermostat ambiance sollicité en permanence
- Consommation énergétique excessive
- Cycles marche/arrêt trop fréquents en mi-saison

**Cause racine probable :**
Pente courbe incorrecte pour type émetteurs ou isolation, mise en service bâclée, modification bâtiment non prise en compte, réglages usine non adaptés.

**Étapes de résolution :**

1. **Identification problème spécifique**
   - **Symptôme A :** Froid quand T°ext basse, OK quand doux
     - → Pente insuffisante (courbe trop "plate")
   - **Symptôme B :** OK quand froid, trop chaud quand doux
     - → Pente excessive (courbe trop "raide")
   - **Symptôme C :** Toujours trop froid
     - → Décalage parallèle insuffisant ou pente trop faible
   - **Symptôme D :** Toujours trop chaud
     - → Décalage parallèle excessif ou pente trop forte

2. **Détermination pente théorique**
   - **Méthode calcul :**
     - Pente = (T°départ grand froid - T°départ mi-saison) / (T°mi-saison - T°grand froid)
     - Exemple : (70°C - 40°C) / (10°C - (-10°C)) = 30/20 = 1,5
   - **Méthode empirique selon émetteurs :**
     - Plancher chauffant : 0,3-0,6
     - Ventilo-convecteurs BT : 0,6-0,9
     - Radiateurs alu BT : 0,8-1,2
     - Radiateurs fonte MT : 1,2-1,8
     - Radiateurs fonte HT : 1,8-2,5

3. **Calcul température départ nécessaire**
   - Identifier température dimensionnement (T°ext base)
   - Exemple : T°base = -7°C dans région (donnée climatique)
   - Identifier puissance/température émetteurs
   - Exemple radiateurs : 20°C ambiance à 60°C départ
   - Déduire T°départ max nécessaire

4. **Traçage courbe adaptée**
   - **Point 1 (grand froid) :**
     - T°ext = température base région (-5 à -15°C selon zone)
     - T°départ = température max émetteurs (35-80°C)
   - **Point 2 (arrêt chauffage) :**
     - T°ext = 15-18°C (seuil arrêt chauffage)
     - T°départ = arrêt ou mini (30-35°C)
   - **Pente = (T°départ P1 - T°départ P2) / (T°ext P2 - T°ext P1)**

5. **Ajustement progressif**
   - Partir de pente calculée ou recommandée
   - Programmer dans régulation chaudière
   - Laisser fonctionner 48h minimum
   - Relever températures intérieures matin/soir
   - Ajuster par incrément 0,1 pente
   - Répéter jusqu'à confort optimal

6. **Optimisation décalage parallèle**
   - Une fois pente correcte, ajuster décalage
   - Décalage = ajustement fin ±2 à ±5°C
   - Ne compense pas une pente incorrecte
   - Utiliser pour petits ajustements saisonniers

7. **Prise en compte spécificités**
   - **Inertie bâtiment :**
     - Bâtiment lourd : inertie forte, anticipation nécessaire
     - Bâtiment léger : réaction rapide, pente peut être plus faible
   - **Apports gratuits :**
     - Forte occupation : réduire légèrement pente
     - Apports solaires importants : idem
   - **Altitude :**
     - Augmenter pente en altitude (T°base plus basse)

8. **Validation sur cycle complet**
   - Tester sur période froide (-5°C ou moins)
   - Tester sur période mi-saison (5-15°C)
   - Vérifier confort stable
   - Mesurer consommation (comparaison avant/après)
   - Ajustement final si nécessaire

**Prévention :**
- Mise en service rigoureuse par professionnel qualifié
- Réévaluation après travaux isolation
- Ajustement après changement émetteurs
- Documentation paramètres (traçabilité)

**Cas particuliers :**

**Installation mixte (radiateurs + plancher) :**
- Créer 2 zones avec courbes différentes
- Ou compromis pente intermédiaire + mélange
- Préférer séparation hydraulique (2 départs)

**Rénovation énergétique :**
- Isolation renforcée → réduire pente (besoin moins important)
- Remplacement fenêtres → idem
- Après travaux : reprendre réglage à zéro

**Régulation pièce par pièce :**
- Têtes thermostatiques radiateurs
- Ajustent localement selon besoins
- Courbe chauffe = base, têtes = ajustement fin

**Outils aide réglage :**
- Applications constructeur (Vaillant, Viessmann, etc.)
- Abaques courbes chauffe (documentation technique)
- Régulateur Lernen/Auto-adapt (apprentissage automatique)

**Formule simplifiée pente :**
Pente ≈ (ΔT eau émetteur × 1,3) / 30
- Exemple radiateurs MT (ΔT = 50°C) : (50 × 1,3) / 30 ≈ 2,2
- Exemple plancher chauffant (ΔT = 15°C) : (15 × 1,3) / 30 ≈ 0,65

---

