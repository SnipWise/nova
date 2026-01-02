## FACT-CHAUD-053: Température ECS instable

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-053 |
| **Catégorie** | Eau Chaude Sanitaire (ECS) |
| **Système** | Régulation ECS |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (surtout instantané) |

**Symptômes :**
- Température ECS varie brutalement (chaud/froid)
- Eau brûlante puis tiède en alternance
- Variations synchrones avec débit
- Inconfort important à la douche
- Température stable puis chutes brutales

**Cause racine probable :**
Sonde température ECS défectueuse, régulation PID mal paramétrée, débitmètre encrassé (signal erratique), vanne 3 voies instable, pression eau fluctuante, échangeur partiellement entartré.

**Étapes de résolution :**

1. **Caractérisation instabilité**
   - Observer le cycle : périodicité variations
   - Variations liées débit (ouvrir/fermer robinet)
   - Variations liées température chauffage
   - Consulter température affichée vs ressentie

2. **Contrôle sonde température ECS**
   - Localiser sonde ECS (sortie échangeur sanitaire)
   - Mesurer résistance sonde (CTN/NTC)
   - Comparer à courbe constructeur
   - Vérifier contact thermique sonde (doigt de gant)
   - Remplacer si valeurs erratiques

3. **Contrôle débitmètre**
   - Démonter et nettoyer débitmètre
   - Vérifier rotation fluide turbine
   - Tester signal électrique (oscilloscope si possible)
   - Signal doit être proportionnel au débit
   - Remplacer si signal instable

4. **Optimisation régulation ECS**
   - Accéder paramètres régulation ECS
   - Ajuster paramètres PID (Proportionnel, Intégral, Dérivé)
   - Augmenter temps intégration (stabilité)
   - Réduire gain proportionnel (éviter oscillations)
   - Tester et ajuster progressivement

5. **Contrôle pression eau**
   - Mesurer pression eau froide sanitaire
   - Doit être stable 2-3 bars
   - Si fluctuations importantes : problème réseau
   - Installer réducteur pression si nécessaire
   - Vérifier vase expansion sanitaire (si présent)

6. **Contrôle vanne 3 voies**
   - Vérifier stabilité position vanne en mode ECS
   - Écouter vibrations ou bruits anormaux
   - Lubrifier mécanisme si prévu constructeur
   - Remplacer si oscillations mécaniques

7. **Contrôle échangeur sanitaire**
   - Entartrage partiel : passages bouchés irrégulièrement
   - Crée turbulences et variations température
   - Détartrer chimiquement
   - Vérifier homogénéité après détartrage

8. **Solutions complémentaires**
   - Installer ballon tampon sanitaire (stabilisation)
   - Installer mitigeur thermostatique (sécurité utilisateur)
   - Régler température ECS légèrement supérieure
   - Installer régulateur débit (stabilisation)

**Prévention :**
- Vérification annuelle sonde ECS
- Nettoyage débitmètre
- Contrôle pression eau sanitaire
- Détartrage préventif échangeur
- Réglage régulation optimisé
- Installation mitigeur thermostatique recommandée

**Note :** Mitigeur thermostatique = sécurité anti-brûlure + confort. Obligatoire dans certains ERP (établissements recevant du public).

---

