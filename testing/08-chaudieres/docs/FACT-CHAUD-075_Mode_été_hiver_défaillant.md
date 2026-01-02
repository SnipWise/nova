## FACT-CHAUD-075: Mode été/hiver défaillant

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-075 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Commutation été/hiver |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières mixtes chauffage+ECS) |

**Symptômes :**
- Chauffage actif en été (hors demande)
- Pas de production ECS en mode hiver
- Commutation automatique ne fonctionne pas
- Mode été/hiver bloqué sur une position
- Chaudière ne suit pas sélection mode

**Cause racine probable :**
Vanne 3 voies bloquée, paramètre mode mal configuré, sonde extérieure défaillante (mode auto), servomoteur vanne défectueux, carte électronique.

**Étapes de résolution :**

1. **Compréhension modes**
   - **Mode Hiver :** chauffage + ECS actifs
   - **Mode Été :** ECS seule, chauffage désactivé
   - **Mode Auto :** commutation selon T°ext (si sonde extérieure)
   - Objectif été : économiser en n'activant pas chauffage

2. **Vérification sélection mode**
   - Contrôler mode sélectionné (afficheur chaudière/thermostat)
   - Tester changement manuel été ↔ hiver
   - Observer réaction chaudière (affichage, vanne 3 voies)
   - Vérifier cohérence saison/mode

3. **Contrôle vanne 3 voies (si présente)**
   - Localiser vanne 3 voies (généralement sortie chaudière)
   - Vérifier servomoteur (bruit rotation lors commutation)
   - Contrôler position vanne (indicateur mécanique)
   - Tester manuellement rotation (levier secours sur certains modèles)
   - Positions typiques :
     - Chauffage seul : flux vers circuit chauffage
     - ECS seule : flux vers échangeur ballon
     - Possible position intermédiaire (mixte)

4. **Test servomoteur vanne**
   - Forcer mode chauffage : servomoteur doit tourner
   - Forcer mode ECS : servomoteur doit tourner sens inverse
   - Écouter bruit moteur (ronronnement)
   - Si pas de mouvement : servomoteur HS ou câblage
   - Vérifier alimentation électrique servomoteur (230V ou 24V)

5. **Contrôle câblage servomoteur**
   - Vérifier connectique servomoteur (bien enfichée)
   - Mesurer tensions commande (multimètre)
   - Contrôler continuité câbles
   - Vérifier commutation tension lors changement mode

6. **Diagnostic vanne mécanique**
   - Si servomoteur OK mais pas d'effet : vanne grippée
   - Causes : entartrage, vieillissement joints, corrosion
   - Tenter rotation manuelle (avec précaution)
   - Démonter et nettoyer si accessible
   - Remplacer si bloquée définitivement

7. **Contrôle mode automatique**
   - Si mode auto été/hiver selon T°ext :
     - Vérifier sonde extérieure fonctionnelle (voir FACT-CHAUD-068)
     - Contrôler seuil commutation (menu installateur)
     - Seuil typique : 15-18°C extérieur
     - Si T°ext > seuil : mode été
     - Si T°ext < seuil : mode hiver
   - Ajuster seuil selon climat local

8. **Vérification priorité ECS**
   - En mode hiver : priorité ECS normale
   - Demande ECS → arrêt chauffage temporaire
   - Vanne 3 voies bascule vers ballon
   - Après satisfaction ECS → retour chauffage
   - Vérifier ce cycle fonctionne correctement

9. **Contrôle carte électronique**
   - Vérifier menu diagnostic (position vanne, mode actif)
   - Tester commande manuelle vanne si disponible
   - Contrôler relais carte (écouter clic commutation)
   - Remplacer carte si commande absente

10. **Paramétrage chaudière**
    - Vérifier activation fonction été/hiver (menu installateur)
    - Contrôler paramètres :
      - Type installation (chauffage seul, mixte, ECS seule)
      - Présence ballon ECS (oui/non)
      - Type vanne (3 voies, 2 vannes séparées)
    - Corriger si mal configuré

**Prévention :**
- Exercice vanne 3 voies hors saison (test été/hiver)
- Contrôle servomoteur lors entretien
- Vérification sonde extérieure (mode auto)
- Détartrage installation selon dureté eau

**Solutions selon diagnostic :**
- Servomoteur HS : remplacement (100-200€)
- Vanne grippée : nettoyage ou remplacement (200-400€)
- Carte électronique : remplacement module ou carte (150-500€)
- Paramétrage : reconfiguration (gratuit)

**Alternatives si panne :**
- Mode été : désactiver manuellement chauffage (commutateur/menu)
- Mode hiver : forcer chauffage actif en permanence
- Installer vanne manuelle temporaire (by-pass)

**Systèmes sans vanne 3 voies :**
- Chaudière chauffage seul + ballon électrique ECS : pas de commutation
- Chaudière ECS instantanée : pas de vanne, gestion électronique
- Systèmes séparés : 2 circulateurs (chauffage + ECS)

**Optimisation été :**
- Arrêt complet chauffage (économies)
- Réduction température ECS possible (55°C mini légionelle)
- Certaines chaudières : arrêt ventilateur été (économie électrique)
- Entretien chaudière profiter de l'été (plus disponible)

**Consommation mode été :**
- Chaudière gaz ECS : ~30-50 m³/mois pour 4 personnes
- Préchauffage chaudière avant puisage : pertes 10-15%
- Isolation ballon importante (pertes statiques)

---

