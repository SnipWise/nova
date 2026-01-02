## FACT-CHAUD-011: Modulation puissance défectueuse

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-011 |
| **Catégorie** | Brûleur & Combustion |
| **Système** | Modulation puissance |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières modulantes) |

**Symptômes :**
- Chaudière fonctionne uniquement en tout ou rien
- Pas de modulation de puissance
- Cycles marche/arrêt fréquents
- Rendement dégradé
- Température départ instable

**Cause racine probable :**
Vanne gaz à modulation défectueuse, carte électronique, sonde température défaillante, paramètres mal configurés, servomoteur vanne bloqué.

**Étapes de résolution :**

1. **Diagnostic modulation**
   - Observer le comportement : puissance fixe ou variable
   - Consulter le menu diagnostic (affichage % puissance)
   - Vérifier si la puissance demandée varie selon besoin
   - Écouter le bruit du brûleur (doit varier)

2. **Contrôle vanne gaz modulante**
   - Vérifier le type de vanne (tout ou rien vs modulante)
   - Contrôler le servomoteur (bruit de fonctionnement)
   - Mesurer la tension de commande (0-10V ou PWM selon modèle)
   - Vérifier la mécanique (pas de blocage)

3. **Contrôle carte électronique**
   - Vérifier les paramètres de puissance (mini/maxi)
   - Contrôler le signal de commande vanne (oscilloscope ou multimètre)
   - Tester le mode manuel si disponible
   - Réinitialiser les paramètres usine

4. **Contrôle sondes température**
   - Vérifier sonde départ (valeur cohérente)
   - Contrôler sonde retour si présente
   - Tester la variation de consigne (impact sur modulation)
   - Remplacer sonde si défectueuse

5. **Contrôle ventilateur modulant**
   - Sur chaudières à ventilateur modulant : vérifier variation vitesse
   - Consulter vitesse dans menu diagnostic
   - Vérifier que la vitesse suit la puissance demandée

6. **Vérification hydraulique**
   - S'assurer d'un débit suffisant (pompe, réseau)
   - Vérifier que le ΔT permet la modulation
   - Contrôler bypass si présent

7. **Optimisation paramètres**
   - Ajuster la puissance maximale selon installation
   - Régler la puissance minimale (éviter cycles courts)
   - Configurer les temps de stabilisation
   - Paramétrer la rampe de montée en puissance

**Prévention :**
- Vérification annuelle de la modulation
- Contrôle paramètres après modification installation
- Lubrification servomoteur si prévu
- Mise à jour firmware carte

**Avantages modulation correcte :**
- Rendement amélioré (+5 à 10%)
- Confort accru (température stable)
- Moins de cycles (longévité accrue)
- Économies d'énergie

---

