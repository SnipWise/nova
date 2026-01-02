## FACT-CHAUD-093: Défaut alimentation 230V

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-093 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Alimentation électrique |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Chaudière complètement morte, pas d'affichage
- Disjoncteur qui saute au démarrage
- Affichage faible ou clignotant
- Réinitialisation horloge fréquente
- Fonctionnement intermittent

**Cause racine probable :**
Disjoncteur défectueux, câblage défaillant, mauvaise connexion bornier, tension secteur anormale, absence de neutre, inversion phase/neutre, défaut terre.

**Étapes de résolution :**

1. **Vérification tableau électrique**
   - Vérifier disjoncteur chaudière en position ON
   - Tester disjoncteur : déclencher et réenclencher
   - Vérifier calibre disjoncteur (généralement 10A ou 16A courbe C)
   - Contrôler état disjoncteur : pas de traces brûlure
   - Si vieux disjoncteur : tester par substitution

2. **Mesure tension secteur**
   - Mesurer tension au tableau : 230V ± 10% (207-253V)
   - Mesurer tension au bornier chaudière (peut différer)
   - Vérifier phase-neutre : ~230V
   - Vérifier phase-terre : ~230V
   - Vérifier neutre-terre : < 5V (sinon problème neutre)

3. **Contrôle câblage alimentation**
   - Vérifier section câble : minimum 1.5mm² (conseillé 2.5mm²)
   - Contrôler état câble (gaine non dégradée)
   - Vérifier longueur : si > 20m, chute tension possible
   - Mesurer résistance câble (doit être ~0 Ω)
   - Contrôler absence de câble sectionné, pincé

4. **Contrôle bornier chaudière**
   - Couper alimentation tableau
   - Ouvrir bornier électrique chaudière
   - Vérifier serrage vis : doivent être fermes
   - Contrôler état connexions : pas d'oxydation, noircissement
   - Vérifier identification : L (phase), N (neutre), T (terre)
   - Resserrer toutes les connexions

5. **Vérification phase/neutre/terre**
   - Identifier phase (fil marron ou noir, ou rouge)
   - Identifier neutre (fil bleu)
   - Identifier terre (fil vert/jaune)
   - ATTENTION inversion phase/neutre : certains appareils ne fonctionnent pas
   - Vérifier avec testeur de polarité ou tournevis testeur

6. **Test continuité terre**
   - Mesurer résistance terre : < 100 Ω (norme < 100 Ω)
   - Si > 100 Ω : problème installation terre
   - Vérifier continuité terre tableau → chaudière
   - Contrôler raccordement terre sur chaudière (cosse serrée)

7. **Diagnostic disjoncteur qui saute**
   - **Immédiatement au démarrage** : court-circuit franc
     - Débrancher charges une par une (pompe, ventilateur, vanne)
     - Identifier composant en court-circuit
   - **Après quelques secondes** : surconsommation
     - Mesurer courant total chaudière (< 10A normalement)
     - Vérifier pas de charge anormale
   - **Aléatoirement** : défaut terre (différentiel) ou disjoncteur défectueux
     - Tester avec autre disjoncteur
     - Vérifier isolement terre tous composants

8. **Contrôle transformateur alimentation**
   - Sur carte : localiser transformateur (composant gros rectangulaire)
   - Mesurer tensions secondaires (selon schéma : 12V, 24V)
   - Si tensions absentes : transformateur HS ou fusible primaire grillé
   - Remplacer transformateur si défectueux (selon modèle : soudé ou enfichable)

9. **Cas micro-coupures réseau**
   - Symptôme : réinitialisation horloge, pertes paramètres
   - Vérifier qualité alimentation secteur (voltmètre enregistreur)
   - Installer onduleur si micro-coupures fréquentes
   - Ou système batterie backup pour maintien mémoire

**Prévention :**
- Vérification annuelle serrage bornier
- Contrôle calibre et état disjoncteur
- Test terre annuel
- Resserrage connexions tableau électrique
- Protection parafoudre + différentiel 30mA obligatoire

**Conformité électrique :**
- Disjoncteur dédié chaudière obligatoire
- Différentiel 30mA type A obligatoire
- Section câble : 1.5mm² mini (2.5mm² conseillé)
- Couleurs normalisées respectées
- Mise à terre < 100 Ω obligatoire

**Outils nécessaires :**
- Multimètre (tension AC, résistance)
- Tournevis testeur phase
- Pince ampèremétrique (diagnostic)
- Testeur différentiel (si disponible)

---

