## FACT-CHAUD-064: Brûleur ne démarre pas sur demande ECS

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-064 |
| **Catégorie** | Eau Chaude Sanitaire (ECS) |
| **Système** | Démarrage ECS |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Ouverture robinet ECS : aucune réaction chaudière
- Brûleur ne s'allume pas en demande sanitaire
- Mode chauffage fonctionne normalement
- Pas de bruit, pas de ventilateur
- Afficheur ne passe pas en mode ECS

**Cause racine probable :**
Débitmètre ne détecte pas débit, carte électronique ne reçoit pas signal, priorité ECS désactivée, programmation bloquée, capteur débit défectueux, pression eau insuffisante.

**Étapes de résolution :**

1. **Vérification basique**
   - Ouvrir robinet ECS en grand (débit maximum)
   - Maintenir ouvert 30 secondes
   - Observer afficheur chaudière :
     - Symbole robinet/goutte doit apparaître
     - Ou passage mode ECS
   - Écouter brûleur (tentative démarrage)

2. **Contrôle débit et pression**
   - Mesurer débit robinet (litres/minute)
   - Débit minimum déclenchement : 2-4 L/min (selon modèle)
   - Si débit faible : problème hydraulique amont
   - Mesurer pression eau froide (manomètre)
   - Pression minimum : 1,5-2 bars
   - Si pression insuffisante : pas de détection débit

3. **Contrôle débitmètre**
   - Voir FACT-CHAUD-059 (détails complets)
   - Vérifier rotation turbine débitmètre
   - Nettoyer si encrassé
   - Tester signal électrique
   - Remplacer si défectueux

4. **Contrôle paramètres chaudière**
   - **Mode chaudière :**
     - Vérifier mode : Été, Hiver, ECS seule
     - Mode "Hiver" ou "Auto" : chauffage + ECS
     - Mode "Été" : ECS seule
     - Mode "Chauffage seul" : pas d'ECS
   - **Priorité ECS :**
     - Paramètre priorité ECS activé
     - Si désactivé : ECS ne démarre pas
   - **Programmation :**
     - Vérifier plages horaires ECS autorisées
     - Mode absence/vacances : ECS parfois bloquée
     - Débloquer programmation

5. **Contrôle carte électronique**
   - Vérifier réception signal débitmètre :
     - Menu diagnostic : affichage débit ECS
     - Ouvrir robinet : débit doit s'afficher
     - Si pas d'affichage : câblage ou carte
   - Contrôler sortie commande :
     - Vanne 3 voies doit commuter
     - Brûleur doit recevoir ordre démarrage
   - Tester mode forcé ECS si disponible

6. **Contrôle vanne 3 voies**
   - Voir FACT-CHAUD-058 (détails complets)
   - Vanne doit commuter en position ECS
   - Écouter bruit moteur vanne
   - Vérifier position vanne (manuelle si possible)
   - Débloquer ou remplacer si nécessaire

7. **Contrôle sécurités**
   - Vérifier absence blocage sécurité :
     - Surchauffe
     - Pression chauffage trop basse
     - Défaut combustion
   - Consulter codes erreur historique
   - Réarmer sécurités si nécessaire

8. **Contrôle contacteur priorité (si présent)**
   - Certaines installations : contacteur externe priorité ECS
   - Contacteur ballons tampons, bouilleurs
   - Vérifier fonctionnement contacteur
   - Tester continuité (multimètre)

9. **Test diagnostic carte**
   - Accéder menu diagnostic/installateur
   - Forcer demande ECS (mode test)
   - Observer comportement :
     - Vanne 3 voies commute ?
     - Ventilateur démarre ?
     - Brûleur s'allume ?
   - Identifier étape défaillante

**Prévention :**
- Vérification annuelle fonctionnement ECS (entretien)
- Test débitmètre (nettoyage)
- Contrôle paramètres après coupure courant
- Vérification vanne 3 voies
- Maintien pression eau suffisante

**Diagnostic par étapes :**
1. Débitmètre détecte débit ? → Si non : nettoyer/remplacer
2. Carte reçoit signal ? → Si non : câblage
3. Vanne 3 voies commute ? → Si non : vanne HS
4. Ventilateur démarre ? → Si non : sécurité/ventilateur
5. Brûleur s'allume ? → Si non : voir combustion

**Seuil déclenchement ECS (exemples) :**
- Saunier Duval : 2,5 L/min
- Vaillant : 2,8 L/min
- Elm Leblanc : 2,5 L/min
- Réglable sur certains modèles (menu installateur)

---

