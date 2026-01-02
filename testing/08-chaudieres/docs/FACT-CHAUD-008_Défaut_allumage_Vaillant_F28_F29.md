## FACT-CHAUD-008: Défaut allumage Vaillant F28/F29

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-008 |
| **Catégorie** | Brûleur & Combustion |
| **Système** | Séquence allumage Vaillant |
| **Gravité** | **Élevée** |
| **Marques** | Vaillant (ecoTEC, turboTEC, atmoTEC) |

**Symptômes :**
- Code F28 : échec allumage (3 tentatives sans succès)
- Code F29 : extinction flamme en cours de fonctionnement
- Réarmement manuel nécessaire
- Parfois alternance F28/F29

**Cause racine probable :**
F28 : absence gaz, vanne gaz défectueuse, électrode, transformateur HT. F29 : ionisation défaillante, coupure gaz, instabilité flamme, carte électronique.

**Étapes de résolution F28 :**

1. **Contrôle gaz**
   - Vérifier robinet gaz ouvert
   - Mesurer pression gaz réseau (20-25 mbar)
   - Purger l'air si installation récente
   - Contrôler compteur gaz (pas de coupure)

2. **Contrôle vanne gaz**
   - Écouter le clic d'ouverture de la vanne
   - Mesurer tension bobines vanne gaz (230V)
   - Mesurer résistance bobines (3-5 kΩ)
   - Vérifier pression sortie vanne en fonctionnement

3. **Contrôle allumage**
   - Vérifier étincelle franche et régulière
   - Nettoyer et ajuster électrode (écartement 4 mm)
   - Contrôler transformateur HT (sortie ~8-10 kV)
   - Vérifier câble HT (pas de fuite à la masse)

4. **Contrôle ventilateur**
   - Vérifier démarrage ventilateur avant allumage (pré-ventilation)
   - Contrôler vitesse ventilateur (menu d.40)
   - Nettoyer si encrassé

**Étapes de résolution F29 :**

1. **Contrôle ionisation**
   - Nettoyer électrode ionisation (papier abrasif fin)
   - Vérifier position dans flamme
   - Mesurer courant ionisation (menu d.48 : > 2 µA)
   - Contrôler câble et connexion masse

2. **Contrôle combustion**
   - Vérifier couleur flamme (bleue)
   - Analyser fumées (CO₂, O₂, CO)
   - Nettoyer brûleur si flamme instable
   - Vérifier débit gaz (pas de variation)

3. **Contrôle carte électronique**
   - Vérifier paramètres ionisation (menu d.96)
   - Réinitialiser paramètres usine si nécessaire
   - Mettre à jour firmware si disponible
   - Remplacer carte si défaut persiste

**Prévention :**
- Contrôle annuel électrodes et transformateur
- Vérification pression gaz
- Analyse combustion
- Mise à jour firmware régulière

**Spécificités Vaillant :**
- Accès menu diagnostic : maintenir touches Mode + OK
- d.40 : vitesse ventilateur (tours/min)
- d.48 : courant ionisation (µA)
- d.96 : paramètres ionisation

---

