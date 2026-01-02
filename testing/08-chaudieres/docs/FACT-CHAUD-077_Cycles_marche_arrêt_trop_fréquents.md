## FACT-CHAUD-077: Cycles marche/arrêt trop fréquents

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-077 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Régulation cycles |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Chaudière démarre et s'arrête toutes les 2-5 minutes
- Bruit cycles allumage répétés
- Brûleur ne tient pas en fonctionnement
- Code erreur cycles courts (certaines marques)
- Usure prématurée composants

**Cause racine probable :**
Sur-dimensionnement chaudière, absence anti-cycles courts, régulation inadaptée, débit hydraulique insuffisant, thermostat mal positionné, installation non équilibrée.

**Étapes de résolution :**

1. **Mesure fréquence cycles**
   - Chronométrer cycles marche/arrêt
   - Noter durée marche et durée arrêt
   - **Cycles courts = marche < 3 minutes**
   - Compter nombre cycles/heure
   - > 6 cycles/heure = problématique

2. **Identification cause principale**
   - **Sur-dimensionnement :**
     - Chaudière trop puissante pour besoins
     - Atteint température rapidement
     - S'arrête, refroidit, redémarre
   - **Différentiel trop faible :**
     - Écart marche/arrêt insuffisant
     - Oscillation rapide autour consigne
   - **Débit insuffisant :**
     - Chaleur non évacuée
     - Surchauffe rapide chaudière

3. **Contrôle dimensionnement**
   - Calculer besoins réels (déperditions bâtiment)
   - Comparer puissance chaudière
   - Exemple problématique : chaudière 25 kW pour besoin 10 kW
   - Sur-dimensionnement > 150% : problème cycles courts

4. **Réduction puissance chaudière**
   - Régler puissance maximale dans menu installateur
   - Paramètre "Puissance max chauffage"
   - Réduire progressivement (ex: de 100% à 60%)
   - Tester et ajuster jusqu'à cycles > 5 minutes
   - Vérifier capacité chauffage grand froid

5. **Ajustement différentiel/hystérésis**
   - Augmenter différentiel thermostat
   - Exemple : 0,5°C → 1,5°C
   - Compromis : cycles plus longs mais variation température
   - Différentiel 1-2°C : bon équilibre général

6. **Temporisation anti-cycles courts**
   - Paramètre chaudière : temps mini arrêt
   - Empêche redémarrage immédiat
   - Régler à 3-5 minutes minimum
   - Certaines chaudières : fonction "anti-tact" intégrée

7. **Contrôle débit hydraulique**
   - Vérifier vitesse pompe circulation (pas trop lente)
   - Mesurer ΔT départ-retour :
     - ΔT faible (< 10°C) : débit excessif ou puissance faible
     - ΔT élevé (> 25°C) : débit insuffisant (cycles courts)
   - Ajuster vitesse pompe
   - Objectif : ΔT 15-20°C

8. **Équilibrage installation**
   - Vérifier ouverture vannes radiateurs
   - Équilibrer réseau (débit homogène)
   - Fermer bypass si trop ouvert
   - Purger air circuit (air = débit réduit)

9. **Vérification thermostat**
   - Contrôler positionnement (voir FACT-CHAUD-072)
   - Thermostat sur radiateur ou source chaleur : cycles courts
   - Déplacer en position neutre
   - Vérifier différentiel thermostat

10. **Solution ballon tampon**
    - Si impossible autre solution : installer ballon tampon
    - Volume 50-100L selon puissance chaudière
    - Stocke chaleur produite
    - Lisse cycles, allonge durée fonctionnement
    - Particulièrement utile chaudière bois, PAC

11. **Optimisation régulation**
    - Passer en régulation climatique si possible
    - Modulation puissance (chaudière modulante)
    - Régulation anticipative
    - Évite cycles tout ou rien

12. **Vérification paramètres avancés**
    - Temps minimal fonctionnement brûleur
    - Temps minimal arrêt
    - Gradient montée puissance (rampe)
    - Hystérésis températures

**Prévention :**
- Dimensionnement correct dès installation
- Calcul déperditions sérieux (bureau étude)
- Régulation adaptée
- Entretien régulier (débit optimal)

**Conséquences cycles courts :**
- **Usure prématurée :**
  - Électrode allumage
  - Vanne gaz
  - Ventilateur
  - Échangeur (chocs thermiques)
- **Surconsommation :**
  - Rendement dégradé (phase démarrage)
  - Pertes balayage pré-ventilation
- **Nuisance sonore :** "tac-tac-tac" permanent
- **Inconfort :** température instable

**Cycles normaux selon système :**
| Type installation | Durée cycle marche | Cycles/heure |
|-------------------|-------------------|--------------|
| Idéal | 10-15 min | 2-4 |
| Acceptable | 5-10 min | 4-6 |
| Problématique | < 5 min | > 6 |
| Critique | < 3 min | > 10 |

**Solutions selon cause :**
| Cause | Solution |
|-------|----------|
| Sur-dimensionnement | Réduire puissance max, ballon tampon |
| Différentiel faible | Augmenter hystérésis |
| Débit insuffisant | Augmenter vitesse pompe, purger |
| Thermostat mal placé | Déplacer en position neutre |
| Absence anti-cycles | Activer temporisation |

**Chaudière modulante vs TOR :**
- **TOR (Tout Ou Rien) :** plus sensible cycles courts
- **Modulante :** adapte puissance, moins de cycles
- Modulation 30-100% : amélioration notable
- Modulation 10-100% : cycles quasi éliminés

---

