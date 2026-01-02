## FACT-CHAUD-068: Sonde extérieure défaillante

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-068 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Régulation climatique |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (sur installations avec régulation climatique) |

**Symptômes :**
- Température extérieure affichée aberrante (ex: -50°C ou +80°C)
- Régulation climatique inopérante (retour loi d'eau fixe)
- Température départ ne s'adapte pas aux conditions extérieures
- Surchauffe ou sous-chauffe locaux
- Code erreur sonde extérieure (certaines marques)

**Cause racine probable :**
Sonde CTN défectueuse, câble sectionné (longueur importante), oxydation connexion, infiltration eau dans sonde, sonde mal positionnée (soleil direct, source chaleur).

**Étapes de résolution :**

1. **Vérification activation régulation climatique**
   - S'assurer que fonction est activée (menu installateur)
   - Vérifier paramétrage courbe de chauffe
   - Contrôler que sonde est bien déclarée dans système
   - Noter température extérieure affichée

2. **Diagnostic valeur affichée**
   - Comparer température affichée avec météo locale
   - Si écart > 5°C : problème sonde ou positionnement
   - Si valeur bloquée : sonde ou câble défectueux
   - Si valeur erratique : mauvaise connexion

3. **Contrôle positionnement sonde**
   - **Emplacement correct :**
     - Face nord ou nord-ouest du bâtiment
     - À l'ombre permanente (pas de soleil direct)
     - Éloignée fenêtres, bouches ventilation, cheminées
     - Hauteur 2-3 mètres
     - Protégée vent dominant mais bien ventilée
   - **Vérifier :**
     - Pas d'exposition soleil (fausse température +10 à +20°C)
     - Pas de source chaleur proximité
     - Pas d'eau stagnante (corrosion)

4. **Mesure résistance sonde**
   - Déconnecter sonde au niveau carte régulation
   - Mesurer résistance (CTN typique 10 kΩ à 25°C)
   - Comparer avec température réelle extérieure et courbe
   - Exemples : 15 kΩ à 15°C, 10 kΩ à 25°C, 28 kΩ à 0°C
   - Résistance infinie : câble coupé ou sonde HS

5. **Contrôle câblage**
   - Vérifier continuité sur toute longueur (peut être > 20m)
   - Contrôler isolation (pas de court-circuit)
   - Inspecter passage gaines, traversées murs
   - Vérifier absence écrasement, coupure
   - Tester connexions intermédiaires si présentes

6. **Contrôle boîtier sonde extérieure**
   - Ouvrir boîtier sonde (vis)
   - Vérifier étanchéité (joint, presse-étoupe câble)
   - Contrôler absence infiltration eau
   - Nettoyer et sécher si humidité
   - Vérifier fixation sonde dans boîtier

7. **Remplacement sonde**
   - Utiliser référence constructeur ou compatible
   - Installer en position optimale (voir point 3)
   - Assurer étanchéité boîtier et passage câble
   - Fixer solidement (pas de vibration vent)
   - Protéger connexions

8. **Paramétrage et test**
   - Vérifier température affichée cohérente
   - Activer régulation climatique
   - Ajuster courbe de chauffe selon besoin (pente + décalage)
   - Tester sur 24-48h : température départ doit évoluer
   - Affiner courbe selon retour occupants

**Prévention :**
- Vérification annuelle température affichée vs météo
- Contrôle étanchéité boîtier sonde
- Vérification positionnement (végétation, modification bâtiment)
- Test câble (continuité, isolation)

**Impact panne sonde extérieure :**
- Perte régulation climatique (retour thermostat seul)
- Inconfort thermique (température intérieure instable)
- Surconsommation (surchauffe par sécurité)
- Usure prématurée chaudière (cycles fréquents)

**Avantages régulation climatique fonctionnelle :**
- Anticipation besoins selon météo
- Confort amélioré (température stable)
- Économies 10-25% selon installation
- Moins de cycles marche/arrêt
- Adaptation automatique saison

**Courbe de chauffe typique :**
- Pente : relation entre T°ext et T°départ
- Exemple : T°ext -10°C → T°départ 70°C, T°ext +15°C → T°départ 35°C
- Ajustement selon isolation, émetteurs, occupation

---

