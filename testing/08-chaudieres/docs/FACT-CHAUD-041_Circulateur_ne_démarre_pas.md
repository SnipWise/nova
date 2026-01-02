## FACT-CHAUD-041: Circulateur ne démarre pas

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-041 |
| **Catégorie** | Hydraulique |
| **Système** | Circulateur / Commande électrique |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques (tous systèmes) |

**Symptômes :**
- Chaudière en chauffe mais radiateurs froids
- Aucun bruit ni vibration du circulateur
- Le corps du circulateur reste froid (moteur non alimenté)
- Pas de déclenchement électrique ni de surchauffe

**Cause racine probable :**
Défaut de commande électrique (relais carte, thermostat d'ambiance, aquastat), câblage défectueux, fusible grillé, circulateur HS (bobinage coupé), mauvais paramétrage régulation.

**Étapes de résolution :**

1. **Vérification du mode de fonctionnement**
   - Contrôler que la chaudière est bien en mode chauffage (pas seulement ECS)
   - Vérifier le thermostat d'ambiance : température de consigne > température ambiante
   - Sur les régulations électroniques : vérifier que le mode chauffage est activé
   - Vérifier la programmation horaire (pas en période d'arrêt programmé)

2. **Contrôle de l'alimentation électrique du circulateur**
   - Couper l'alimentation générale de la chaudière
   - Accéder au bornier du circulateur (retirer le capot électrique)
   - Remettre sous tension en mode chauffage
   - Mesurer la tension aux bornes du circulateur avec un multimètre (230V AC attendu)
   - **Si 230V présent** : défaut du circulateur lui-même → passer à l'étape 6
   - **Si 0V** : problème de commande en amont → passer à l'étape 3

3. **Vérification du thermostat d'ambiance**
   - Tester le contact du thermostat (multimètre en position continuité/ohmmètre)
   - En mode chauffage avec consigne > température : contact doit être fermé (0 ohm)
   - Si contact ouvert (∞ ohm) : thermostat défectueux ou mal réglé
   - Test de contournement : ponter les bornes du thermostat (fil volant) pour tester
   - Si le circulateur démarre : thermostat HS → remplacement

4. **Contrôle du relais de commande circulateur (carte électronique)**
   - Localiser le relais de commande circulateur sur la carte (voir schéma électrique)
   - En mode chauffage, le relais doit être enclenché (clic audible)
   - Mesurer la tension de sortie relais (doit être 230V)
   - Si relais n'enclenche pas : défaut de la carte électronique ou de l'aquastat

5. **Vérification de l'aquastat ou sonde de température**
   - L'aquastat autorise le circulateur uniquement si l'eau de chaudière est chaude
   - Vérifier la température de l'eau chaudière (affichage digital)
   - Si température > 40°C et circulateur ne démarre pas : aquastat ou carte défectueux
   - Tester l'aquastat (mesure de résistance, comparaison avec courbe constructeur)

6. **Test du circulateur (bobinage moteur)**
   - Couper l'alimentation, débrancher les fils du circulateur
   - Mesurer la résistance du bobinage avec un ohmmètre
   - Valeur normale : 50-500 ohms selon modèle
   - **Si résistance infinie (∞)** : bobinage coupé → **circulateur HS, remplacement**
   - **Si résistance correcte** : vérifier qu'il n'est pas bloqué mécaniquement (voir FACT-CHAUD-040)

7. **Contrôle du câblage et des connexions**
   - Inspecter visuellement les fils et connexions (oxydation, brûlures, déconnexions)
   - Vérifier la continuité des fils entre carte et circulateur
   - Resserrer les connexions desserrées
   - Remplacer les cosses ou dominos défectueux

8. **Test avec circulateur de secours (diagnostic)**
   - Si disponible, brancher temporairement un circulateur fonctionnel connu
   - Si ce circulateur démarre : confirme que le circulateur d'origine est HS
   - Si ce circulateur ne démarre pas non plus : problème de commande (carte, thermostat)

9. **Remplacement selon diagnostic**
   - **Circulateur HS** : remplacement (voir FACT-CHAUD-039 étape 8)
   - **Thermostat HS** : remplacement du thermostat d'ambiance
   - **Carte électronique HS** : remplacement de la carte (pièce coûteuse, vérifier garantie)

**Prévention :**
- Tester le bon fonctionnement du chauffage en début de saison
- Vérifier les connexions électriques lors de l'entretien annuel
- Protéger la carte électronique de l'humidité

**Spécificités techniques :**
- Certains circulateurs ont une protection thermique interne qui peut se déclencher
- Les circulateurs à commande PWM (modulation de largeur d'impulsion) nécessitent un signal spécifique

**Avertissements sécurité :**
- Toujours couper l'alimentation avant de toucher les connexions électriques
- Ne jamais ponter de sécurité de manière permanente (test uniquement)

---

