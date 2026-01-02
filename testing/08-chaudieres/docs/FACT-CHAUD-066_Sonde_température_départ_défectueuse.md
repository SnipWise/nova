## FACT-CHAUD-066: Sonde température départ défectueuse

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-066 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Sonde température départ |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Affichage température incohérente ou absente
- Code erreur sonde départ (ex: Saunier Duval F73, Vaillant F75)
- Température affichée bloquée ou erratique
- Chaudière en sécurité ou fonctionne en mode dégradé
- Pas de modulation de puissance

**Cause racine probable :**
Sonde CTN/NTC défectueuse, oxydation connectique, câble coupé ou court-circuité, connecteur déconnecté, corrosion doigt de gant.

**Étapes de résolution :**

1. **Diagnostic initial**
   - Relever température affichée et comparer au réel (thermomètre contact)
   - Vérifier code erreur spécifique dans menu diagnostic
   - Noter si valeur bloquée ou variable anormalement
   - Tester en mode chauffage et ECS si applicable

2. **Mesure résistance sonde**
   - Couper alimentation électrique chaudière
   - Déconnecter sonde au niveau carte électronique
   - Mesurer résistance avec multimètre (ohmmètre)
   - Valeurs courantes CTN 10k : 10 kΩ à 25°C, 3,3 kΩ à 50°C, 1,5 kΩ à 70°C
   - Comparer avec courbe constructeur (notice technique)

3. **Test variation résistance**
   - Chauffer légèrement la sonde (chaleur main ou eau chaude)
   - La résistance doit diminuer progressivement (CTN)
   - Si résistance constante : sonde morte
   - Si résistance infinie : sonde coupée ou câble sectionné
   - Si résistance nulle : court-circuit

4. **Contrôle câblage**
   - Vérifier continuité câble entre sonde et carte
   - Contrôler isolation câble (pas de court-circuit à la masse)
   - Inspecter connectique (oxydation, corrosion)
   - Nettoyer contacts avec bombe contact électronique
   - Vérifier fixation connecteur (bien clipsé)

5. **Contrôle doigt de gant**
   - Déposer la sonde du doigt de gant
   - Vérifier absence d'eau dans doigt de gant (corrosion)
   - Nettoyer et sécher si humidité
   - Appliquer pâte thermique pour bon contact thermique
   - Vérifier fixation doigt de gant (pas de jeu)

6. **Remplacement sonde**
   - Identifier référence exacte selon modèle chaudière
   - Déposer ancienne sonde (noter position, longueur immergée)
   - Installer sonde neuve avec pâte thermique
   - Reconnecter en respectant polarité si applicable
   - Isoler connexion de l'humidité

7. **Test et calibration**
   - Remettre en service la chaudière
   - Vérifier température affichée cohérente
   - Comparer affichage avec thermomètre de référence
   - Laisser chauffer : vérifier évolution température
   - Tester modulation puissance
   - Calibrer si fonction disponible (menu installateur)

**Prévention :**
- Vérification annuelle valeur sonde (menu diagnostic)
- Contrôle connectique (oxydation)
- Protection connexions contre humidité (graisse silicone)
- Remplacement préventif si valeurs dérivent (± 3°C)
- Application pâte thermique lors entretien

**Valeurs typiques sondes CTN :**
- 10 kΩ à 25°C
- 5,8 kΩ à 40°C
- 3,3 kΩ à 50°C
- 2,0 kΩ à 60°C
- 1,5 kΩ à 70°C
- 0,8 kΩ à 90°C

**Note :** En cas de panne sonde, certaines chaudières passent en mode sécurité (arrêt), d'autres en mode dégradé (puissance fixe, température estimée).

---

