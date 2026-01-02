## FACT-CHAUD-020: Variation puissance incohérente

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-020 |
| **Catégorie** | Brûleur & Combustion |
| **Système** | Régulation puissance |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières modulantes) |

**Symptômes :**
- Puissance affichée ne correspond pas au comportement
- Flamme visuelle ne correspond pas à % affiché
- Brûleur en puissance max alors que faible demande
- Température départ dépasse consigne largement
- Ou inversement : puissance insuffisante

**Cause racine probable :**
Sonde température défectueuse (valeur erronée), carte électronique défaillante, paramètres mal configurés, vanne gaz défectueuse, capteur pression différentielle défectueux.

**Étapes de résolution :**

1. **Diagnostic comportement**
   - Observer puissance affichée menu diagnostic
   - Comparer avec flamme réelle (hauteur, bruit)
   - Mesurer température départ réelle vs consigne
   - Noter écart et incohérences

2. **Contrôle sondes température**
   - Mesurer résistance sonde départ (CTN/NTC)
   - Comparer à courbe constructeur (température/résistance)
   - Exemple : 10 kΩ à 25°C, 3,3 kΩ à 50°C (selon modèle)
   - Remplacer sonde si valeur incohérente

3. **Contrôle vanne gaz modulante**
   - Mesurer signal commande vanne (tension 0-10V ou PWM)
   - Vérifier que signal varie selon puissance demandée
   - Contrôler servomoteur vanne (bruit, mouvement)
   - Mesurer pression gaz brûleur (doit varier avec puissance)

4. **Contrôle capteur pression différentielle**
   - Sur certaines chaudières : capteur ΔP pour calcul débit
   - Vérifier valeur capteur cohérente
   - Nettoyer prises de pression si bouchées
   - Remplacer capteur si défectueux

5. **Contrôle carte électronique**
   - Vérifier paramètres puissance (mini/maxi)
   - Régler puissance max selon installation
   - Vérifier algorithme modulation (PID)
   - Réinitialiser paramètres usine
   - Mettre à jour firmware si disponible

6. **Test mode manuel**
   - Accéder mode test (forçage puissance)
   - Imposer différents % de puissance
   - Vérifier cohérence flamme/puissance imposée
   - Identifier défaut (capteur, vanne, carte)

7. **Optimisation paramètres**
   - Ajuster paramètres PID régulation
   - Coefficient proportionnel (P)
   - Temps intégration (I)
   - Temps dérivation (D)
   - Demande expertise si nécessaire

**Prévention :**
- Vérification annuelle sondes température
- Contrôle modulation en mode diagnostic
- Mise à jour firmware régulière
- Test cohérence puissance/flamme

**Menus diagnostic utiles (exemples) :**
- Vaillant : d.40 (vitesse ventilateur), d.41 (% puissance)
- Saunier Duval : P01 (puissance demandée)
- Frisquet : températures et puissance via boîtier Vision
- De Dietrich : menu Service → Diagnostic

---

