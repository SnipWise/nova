## FACT-CHAUD-092: Fusible carte grillé

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-092 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Protection électrique carte |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Chaudière totalement inerte, aucun affichage
- Pas de LED, pas de réaction aux boutons
- Ou : fonction spécifique HS (ventilateur, pompe, etc.)
- Fusible noirci ou filament rompu visible

**Cause racine probable :**
Court-circuit sur composant protégé, surtension secteur, composant défaillant en aval, humidité, vieillissement fusible, erreur câblage.

**Étapes de résolution :**

1. **Identification fusible défectueux**
   - Couper alimentation 230V
   - Localiser les fusibles sur la carte (cylindres verre ou CMS)
   - Types courants : 5x20mm verre, fusibles CMS, porte-fusible
   - Tester continuité au multimètre (0 Ω si bon)
   - Déposer et inspecter visuellement

2. **Identification type fusible**
   - Noter valeur marquée : ex. T2A (temporisé 2 ampères)
   - F = rapide (Fast), T = temporisé (Time-lag)
   - Tensions : 250V ou 125V marquées
   - Exemples courants chaudières :
     - Alimentation générale : T3.15A ou T5A (250V)
     - Ventilateur : T2A (250V)
     - Pompe : T2A ou T3.15A
     - Circuits basse tension : T500mA ou T1A (125V)

3. **Recherche cause du défaut**
   - **IMPORTANT** : Ne jamais remplacer sans chercher la cause
   - Fusible = protection, pas pièce d'usure
   - Débrancher la charge protégée (pompe, ventilateur, vanne)
   - Tester isolement charge (résistance > 1 MΩ vers masse)
   - Mesurer résistance/bobine (doit correspondre valeurs nominales)
   - Vérifier absence court-circuit évident

4. **Contrôle composants aval**
   - Si fusible pompe : tester pompe isolée en direct
   - Si fusible ventilateur : tester ventilateur
   - Si fusible vanne gaz : tester bobines vanne
   - Si fusible alimentation : chercher court-circuit carte
   - Utiliser multimètre mode résistance

5. **Vérification absence humidité**
   - Inspecter carte : traces d'humidité, condensation
   - Sécher carte complètement si humide (air sec tiède)
   - Vérifier joint carter étanche
   - Contrôler absence infiltration eau

6. **Remplacement fusible**
   - Utiliser fusible **exactement identique** (calibre, type, tension)
   - JAMAIS de valeur supérieure (risque incendie)
   - JAMAIS de fil ou "bricolage" à la place
   - Enficher fermement dans porte-fusible
   - Sur fusible CMS : soudure avec fer température contrôlée

7. **Test remise en service**
   - Reconnecter les charges une par une
   - Remettre tension et tester
   - Observer comportement : fusible doit tenir
   - Si re-grille immédiatement : défaut persistant
   - Mesurer consommation si possible (pince ampèremétrique)

8. **Cas fusible re-grille immédiatement**
   - Court-circuit franc : chercher composant HS sur carte
   - Vérifier condensateurs gonflés
   - Tester diodes, transistors, relais
   - Débrancher charges suspectes une par une
   - Si pas de solution : remplacement carte nécessaire

**Prévention :**
- Contrôle annuel serrage connecteurs (mauvais contact = échauffement)
- Protection parafoudre secteur
- Éviter humidité locale chaufferie
- Vérification fonctionnement charges (pompe, ventilateur)

**Stock préventif recommandé :**
- Fusibles T3.15A 250V (x2)
- Fusibles T2A 250V (x2)
- Conserver dans boîte étanche

**Références fusibles courants :**
- Standard 5x20mm temporisé trouvable commerce électronique
- Marques fiables : Littelfuse, Schurter, Eaton Bussmann
- Éviter fusibles "premier prix" sous-calibrés

---

