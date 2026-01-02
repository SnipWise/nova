## FACT-CHAUD-078: Température départ trop élevée/faible

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-078 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Régulation température départ |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- **Température trop élevée :** surchauffe locaux, inconfort, surconsommation, cycles courts
- **Température trop faible :** locaux froids, inconfort, plainte occupants
- Température départ ne correspond pas à consigne
- Température départ inadaptée aux besoins

**Cause racine probable :**
Consigne mal réglée, courbe chauffe inadaptée, sonde température défectueuse, régulation défaillante, sur/sous-dimensionnement émetteurs.

**Étapes de résolution - Température TROP ÉLEVÉE :**

1. **Diagnostic situation**
   - Mesurer température départ réelle (thermomètre)
   - Comparer avec consigne affichée
   - Mesurer température intérieure
   - Relever température extérieure
   - Identifier si problème permanent ou contextuel

2. **Vérification consigne départ**
   - Consulter consigne température départ (menu chaudière)
   - **Valeurs normales selon émetteurs :**
     - Plancher chauffant : 35-45°C
     - Radiateurs BT : 45-55°C
     - Radiateurs MT : 55-65°C
     - Radiateurs HT : 70-80°C
   - Réduire consigne si trop élevée

3. **Contrôle courbe de chauffe**
   - Si régulation climatique active
   - Vérifier pente et décalage (voir FACT-CHAUD-070, 071)
   - Pente trop forte → température départ excessive
   - Réduire pente ou décalage parallèle

4. **Vérification sonde départ**
   - Température affichée vs réelle
   - Si sonde indique trop froid → chaudière chauffe trop
   - Tester sonde (voir FACT-CHAUD-066)
   - Remplacer si défectueuse

5. **Contrôle modulation**
   - Chaudière doit moduler selon besoin
   - Vérifier modulation active (menu diagnostic)
   - Si bloquée puissance max : problème vanne gaz ou carte

6. **Solutions température trop élevée**
   - Réduire consigne température départ
   - Ajuster courbe de chauffe (pente, décalage)
   - Activer régulation climatique si disponible
   - Installer vannes thermostatiques radiateurs
   - Réduire puissance max chaudière si surdimensionnée

**Étapes de résolution - Température TROP FAIBLE :**

1. **Diagnostic situation**
   - Mesurer température départ
   - Comparer avec consigne
   - Vérifier température intérieure (toutes pièces)
   - Noter température extérieure
   - Identifier radiateurs tièdes ou froids

2. **Vérification consigne départ**
   - Consulter consigne
   - Augmenter si inférieure aux besoins
   - Vérifier consigne max autorisée (limite protection)

3. **Contrôle courbe de chauffe**
   - Pente trop faible → température départ insuffisante froid
   - Augmenter pente ou décalage parallèle
   - Vérifier température extérieure cohérente (sonde OK)

4. **Vérification sonde départ**
   - Si sonde indique trop chaud → chaudière ne chauffe pas assez
   - Exemple : 70°C affiché mais 50°C réel
   - Tester et remplacer sonde si défectueuse

5. **Contrôle puissance chaudière**
   - Vérifier que chaudière atteint puissance max
   - Consulter menu diagnostic (% puissance)
   - Si limitée : problème gaz, vanne, modulation

6. **Vérification hydraulique**
   - Pompe circulation fonctionne (vitesse correcte)
   - Débit suffisant (pas d'air, pas d'obstruction)
   - Vanne mélangeuse ouverte (si présente)
   - By-pass pas trop ouvert

7. **Contrôle émetteurs**
   - Radiateurs correctement dimensionnés
   - Tous radiateurs chauffent
   - Vannes ouvertes
   - Radiateurs non entartrés (si eau dure)
   - Pas de poche d'air

8. **Solutions température trop faible**
   - Augmenter consigne température départ
   - Ajuster courbe de chauffe (pente, décalage)
   - Augmenter puissance max chaudière
   - Vérifier dimensionnement installation
   - Améliorer isolation (réduire besoins)
   - Envisager émetteurs supplémentaires

**Adaptation température selon conditions :**

**Grand froid (T°ext < -5°C) :**
- Température départ maximale nécessaire
- Vérifier que chaudière atteint consigne

**Mi-saison (T°ext 5-15°C) :**
- Température départ réduite
- Éviter surchauffe
- Régulation climatique optimale

**Nuit/réduit :**
- Réduction température départ possible
- Ou arrêt complet selon stratégie

**ECS prioritaire :**
- Température départ peut baisser temporairement (vanne 3 voies)
- Normal, retour chauffage après production ECS

**Prévention :**
- Réglage minutieux mise en service
- Vérification saisonnière (automne)
- Contrôle sondes température annuel
- Ajustement après travaux isolation/émetteurs

**Tableau températures recommandées :**
| Émetteurs | T°ext -10°C | T°ext 0°C | T°ext +10°C |
|-----------|-------------|-----------|-------------|
| Plancher chauffant | 40°C | 35°C | 30°C |
| Radiateurs BT | 50°C | 45°C | 35°C |
| Radiateurs MT | 65°C | 55°C | 45°C |
| Radiateurs HT | 75°C | 65°C | 50°C |

**Impact température départ :**
- **Trop élevée :**
  - Surconsommation (+10-30%)
  - Inconfort (surchauffe)
  - Cycles courts
  - Perte condensation (chaudières condensation)
- **Trop faible :**
  - Inconfort (froid)
  - Insatisfaction occupants
  - Humidité (pas assez chaud pour sécher)

**Optimisation condensation :**
- Chaudière condensation : T°retour < 54°C
- Baisser T°départ favorise condensation
- Rendement optimal : T°départ < 55°C
- Compromis confort/rendement à trouver

---

