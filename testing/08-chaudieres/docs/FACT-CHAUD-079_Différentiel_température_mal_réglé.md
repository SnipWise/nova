## FACT-CHAUD-079: Différentiel température mal réglé

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-079 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Différentiel régulation |
| **Gravité** | Faible |
| **Marques** | Multi-marques |

**Symptômes :**
- Cycles marche/arrêt trop fréquents (différentiel trop faible)
- Variation température intérieure excessive (différentiel trop élevé)
- Inconfort thermique (oscillations température)
- Chaudière ne redémarre pas assez vite (différentiel trop grand)

**Cause racine probable :**
Paramètre différentiel/hystérésis mal configuré, valeur usine inadaptée installation, absence réglage lors mise en service.

**Étapes de résolution :**

1. **Compréhension différentiel**
   - **Différentiel = écart entre seuil marche et seuil arrêt**
   - **Exemple :**
     - Consigne 20°C, différentiel 1°C
     - Arrêt chauffage à 20°C
     - Redémarrage chauffage à 19°C
   - **Hystérésis = autre terme pour différentiel**

2. **Identification problème**
   - **Différentiel trop faible (< 0,5°C) :**
     - Cycles marche/arrêt très fréquents
     - Usure composants
     - Nuisance sonore
     - Surconsommation
   - **Différentiel trop élevé (> 2°C) :**
     - Variation température ressentie
     - Inconfort (trop chaud puis trop froid)
     - Cycles longs (économique mais moins confortable)

3. **Mesure différentiel actuel**
   - Observer température départ ou intérieure
   - Noter température arrêt chaudière (T_stop)
   - Noter température redémarrage (T_start)
   - Calculer : Différentiel = T_stop - T_start
   - Exemple : arrêt 20°C, redémarrage 18,5°C → différentiel 1,5°C

4. **Localisation réglage différentiel**
   - **Thermostat d'ambiance :**
     - Menu paramètres/réglages
     - Souvent nommé "Hystérésis", "Différentiel", "Écart"
   - **Chaudière :**
     - Menu installateur
     - Paramètre "Hystérésis température"
   - **Régulation climatique :**
     - Parfois différentiel fixe (non modifiable)

5. **Réglage optimal différentiel**
   - **Valeurs recommandées :**
     - **Thermostat ambiance classique :** 0,5-1,5°C
     - **Thermostat précis/modulant :** 0,3-0,5°C
     - **Régulation climatique :** 0,5-1°C
     - **Plancher chauffant :** 0,5-1°C (forte inertie)
     - **Radiateurs :** 1-2°C (acceptable)
   - **Compromis :**
     - Faible différentiel : confort max, risque cycles courts
     - Différentiel élevé : cycles longs, confort réduit

6. **Ajustement progressif**
   - Partir valeur actuelle
   - Modifier par incrément 0,2-0,5°C
   - Observer 24-48h
   - Mesurer :
     - Fréquence cycles
     - Variation température ressentie
     - Confort occupants
   - Réajuster si nécessaire

7. **Adaptation selon inertie**
   - **Faible inertie (bâtiment léger, radiateurs alu) :**
     - Réaction rapide
     - Différentiel faible possible (0,5-1°C)
   - **Forte inertie (bâtiment lourd, plancher chauffant) :**
     - Réaction lente
     - Différentiel plus élevé acceptable (1-2°C)
     - Inertie compense variations

8. **Différentiel sur température départ**
   - Certaines chaudières : différentiel sur T°départ (pas T°ambiance)
   - Valeurs plus élevées : 5-10°C typique
   - Exemple : consigne 60°C, arrêt à 60°C, redémarrage à 50°C
   - Évite cycles courts sur production chaleur

9. **Cas particulier régulation modulante**
   - Thermostat modulant (OpenTherm) : différentiel moins critique
   - Modulation progressive puissance
   - Évite marche/arrêt franc
   - Différentiel peut être faible (0,3-0,5°C)

10. **Vérification anti-cycles courts**
    - En complément différentiel : temporisation
    - Temps minimum arrêt avant redémarrage
    - Paramètre chaudière "Temps anti-tact"
    - Typique : 3-5 minutes
    - Combine différentiel + temporisation = régulation optimale

**Prévention :**
- Réglage lors mise en service (professionnel)
- Réévaluation après modification installation
- Ajustement saisonnier si nécessaire
- Documentation valeur retenue

**Exemples configuration :**

**Configuration confort max (faibles variations) :**
- Différentiel : 0,5°C
- Anti-cycles : 5 minutes
- Régulation modulante recommandée

**Configuration économique (cycles longs) :**
- Différentiel : 1,5-2°C
- Anti-cycles : 3 minutes
- Acceptable si inertie bâtiment

**Configuration équilibrée :**
- Différentiel : 1°C
- Anti-cycles : 3-4 minutes
- Bon compromis général

**Interaction autres paramètres :**
- Différentiel + Anticipation sonde = régulation fine
- Différentiel + Courbe chauffe = optimisation globale
- Différentiel + Programmation horaire = confort programmé

**Mesure confort :**
- Variation température < 1°C : excellent
- Variation 1-2°C : bon
- Variation 2-3°C : acceptable
- Variation > 3°C : inconfortable, revoir réglages

**Relation différentiel/cycles :**
| Différentiel | Cycles/heure (typique) | Confort |
|--------------|------------------------|---------|
| 0,3°C | 6-8 | Excellent |
| 0,5°C | 4-6 | Très bon |
| 1°C | 3-4 | Bon |
| 1,5°C | 2-3 | Correct |
| 2°C | 1-2 | Acceptable |
| > 2°C | < 1 | Variations perceptibles |

---

