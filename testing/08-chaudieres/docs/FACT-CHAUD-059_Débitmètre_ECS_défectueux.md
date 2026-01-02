## FACT-CHAUD-059: Débitmètre ECS défectueux

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-059 |
| **Catégorie** | Eau Chaude Sanitaire (ECS) |
| **Système** | Détection débit ECS |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques (chaudières instantanées) |

**Symptômes :**
- Chaudière ne détecte pas demande ECS
- Pas de démarrage brûleur à l'ouverture robinet
- Démarrage aléatoire ou tardif
- Code erreur débit ECS insuffisant
- Brûleur s'arrête en cours de puisage

**Cause racine probable :**
Turbine débitmètre bloquée (calcaire, impuretés), aimant turbine manquant ou cassé, capteur Hall défectueux, câblage débitmètre coupé, corps débitmètre fissuré (fuite).

**Étapes de résolution :**

1. **Compréhension fonctionnement**
   - Débitmètre = turbine + aimant + capteur Hall
   - Eau fait tourner turbine
   - Aimant génère impulsions électriques (capteur Hall)
   - Carte électronique compte impulsions → calcul débit
   - Seuil déclenchement : 2-4 L/min selon modèles

2. **Test fonctionnement débitmètre**
   - Ouvrir robinet ECS progressivement
   - Observer afficheur chaudière :
     - Symbole robinet qui s'affiche = détection OK
     - Ou affichage débit (L/min) si disponible
   - Si aucune détection : débitmètre HS

3. **Démontage débitmètre**
   - Fermer vanne entrée eau froide chaudière
   - Purger pression (ouvrir robinet ECS)
   - Déconnecter câble électrique débitmètre
   - Dévisser débitmètre (clé, selon modèle)
   - Attention : eau résiduelle (prévoir récipient)

4. **Inspection visuelle**
   - Extraire turbine du corps débitmètre
   - Vérifier rotation turbine (souffler ou tourner doigt)
   - Turbine bloquée : calcaire, impuretés, cheveu
   - Vérifier présence aimant sur turbine
   - Aimant cassé/absent : turbine inefficace

5. **Nettoyage débitmètre**
   - Démonter turbine complètement
   - Nettoyer corps débitmètre (brosse, eau)
   - Détartrer si calcaire (vinaigre blanc)
   - Rincer abondamment
   - Nettoyer turbine (aimant, pales)
   - Vérifier axe turbine (pas de jeu excessif)
   - Remonter turbine : doit tourner librement

6. **Test électrique capteur**
   - Mesurer résistance capteur Hall (si accessible)
   - Valeur typique : 500-1500 Ω (selon modèle)
   - Ou tester signal électrique :
     - Reconnecter câble électrique
     - Souffler dans débitmètre (turbine tourne)
     - Mesurer impulsions (oscilloscope ou multimètre AC)
   - Signal absent : capteur HS → remplacement débitmètre

7. **Contrôle câblage**
   - Vérifier continuité câble débitmètre → carte
   - Contrôler connecteur (oxydation, mauvais contact)
   - Vérifier alimentation capteur (5-12V DC selon modèle)

8. **Remontage et test**
   - Remonter débitmètre avec joint neuf (obligatoire)
   - Serrer modérément (pas trop : risque casse plastique)
   - Reconnecter câble électrique
   - Ouvrir vanne eau froide progressivement
   - Vérifier absence fuite
   - Tester démarrage ECS (ouvrir robinet)
   - Vérifier affichage débit ou symbole

9. **Calibration débit (si paramétrable)**
   - Certaines chaudières : réglage seuil déclenchement
   - Ajuster selon installation (pression, tuyauterie)
   - Seuil trop élevé : pas de détection petits débits
   - Seuil trop bas : déclenchements intempestifs

**Prévention :**
- Nettoyage annuel débitmètre (eau dure)
- Installation filtre entrée eau froide (protection)
- Vérification test ECS lors entretien
- Traitement anti-tartre
- Remplacement préventif (durée vie ~10 ans)

**Modèles débitmètres courants :**
- Saunier Duval : 0020133280
- Vaillant : 0020039057
- Elm Leblanc : 87167808640
- Débitmètres souvent interchangeables (vérifier raccords)

**Coût remplacement :**
- Débitmètre seul : 30-80€
- Main d'œuvre : 1 heure

---

