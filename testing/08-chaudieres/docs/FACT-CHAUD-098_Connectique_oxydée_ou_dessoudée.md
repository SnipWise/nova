## FACT-CHAUD-098: Connectique oxydée ou dessoudée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-098 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Connecteurs carte électronique |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Fonctionnement intermittent aléatoire
- Perte de fonction spécifique (pompe, ventilateur, sonde)
- Faux contacts au démarrage
- Besoin de "tapoter" la chaudière pour redémarrer
- Sonde température avec valeurs erratiques
- Erreurs aléatoires non reproductibles

**Cause racine probable :**
Oxydation contacts (humidité), vibrations desserrant connexions, échauffement excessif (mauvais contact), corrosion (condensation), soudures froides, vieillissement.

**Étapes de résolution :**

1. **Identification connecteurs problématiques**
   - Observer quelles fonctions sont intermittentes
   - Localiser connecteurs associés sur carte :
     - Sondes température (CTN/NTC)
     - Pompe, ventilateur, vanne gaz
     - Alimentation 230V
     - Thermostat d'ambiance
   - Inspecter visuellement

2. **Diagnostic visuel connecteurs**
   - Retirer carte électronique
   - Inspecter chaque connecteur :
     - **Oxydation** : dépôt vert/blanc, contacts ternis
     - **Surchauffe** : plastique noirci, déformé
     - **Dessoudure** : soudure fissurée, composant mobile
     - **Corrosion** : dépôts brunâtres, poudre verte (cuivre oxydé)

3. **Nettoyage connecteurs oxydés**
   - **Connecteurs enfichables** (cosses Faston, connecteurs à clip) :
     - Débrancher connecteur
     - Nettoyer contacts mâles et femelles :
       - Bombe contact électrique (spray)
       - Ou alcool isopropylique + brosse douce
       - Ou gomme douce type Caig DeoxIT
     - Sécher complètement
     - Rebrancher fermement
   - **Bornier à vis** :
     - Desserrer vis
     - Nettoyer cosse fil (papier abrasif fin si oxydée)
     - Nettoyer borne (brosse laiton)
     - Revisser fermement

4. **Traitement anti-oxydation**
   - Après nettoyage, appliquer :
     - Graisse conductrice (vaseline technique)
     - Ou spray contact (CRC Contact, WD40 Contact Cleaner)
   - Protège contre oxydation future
   - Améliore conductivité

5. **Réparation soudures froides**
   - Identifier soudures fissurées (loupe) :
     - Aspect terne, granuleux
     - Fissure visible autour pin
     - Composant qui bouge légèrement
   - Refaire soudure :
     - Chauffer soudure existante + ajouter flux
     - Ajouter un peu étain neuf
     - Soudure doit être brillante, conique
   - Zones critiques :
     - Connecteurs puissance (230V, pompe, ventilateur)
     - Transformateur (composant lourd)
     - Relais (composant soumis vibrations)

6. **Remplacement connecteur défectueux**
   - Si connecteur plastique cassé/fondu :
     - Dessouder ancien connecteur
     - Identifier référence (pas, nombre broches)
     - Souder nouveau connecteur identique
   - Types courants :
     - Borniers à vis KF2EDGK (pas 5.08mm)
     - Connecteurs XH, JST (pas 2.54mm)
     - Cosses Faston 6.3mm

7. **Renforcement mécanique**
   - Si dessoudure due vibrations :
     - Ajouter colle thermofusible (hot glue) pour fixation mécanique
     - Ou mousse isolante entre composant et capot
   - Évite contraintes mécaniques répétées

8. **Contrôle continuité fils**
   - Mesurer continuité fil de bout en bout
   - Si coupure interne fil :
     - Couper fil défectueux
     - Dénuder
     - Raccorder avec domino ou soudure + gaine thermo
   - Vérifier absence faux contact dans gaine

9. **Protection contre humidité**
   - Si oxydation due humidité :
     - Vérifier étanchéité capot chaudière
     - Améliorer ventilation local
     - Appliquer vernis tropicalisation sur carte (précaution)
   - Installer absorbeur humidité dans capot si nécessaire

10. **Vérification serrage**
    - Resserrer TOUS les connecteurs à vis
    - Vérifier que cosses enfichables sont bien clipsées
    - Noter couples de serrage si spécifié (borniers puissance)

**Prévention :**
- Vérification annuelle serrage connecteurs
- Nettoyage préventif contacts tous les 2-3 ans
- Protection contre humidité locale chaufferie
- Application spray contact après chaque intervention

**Outils et produits :**
- Bombe contact électrique (CRC, WD40)
- Alcool isopropylique 99%
- Graisse conductrice (Electrolube, Caig DeoxIT)
- Fer à souder température contrôlée
- Flux décapant (pas acide)
- Multimètre continuité

**Zones sensibles :**
- Sondes température (courant faible, sensibles oxydation)
- Connecteurs puissance (échauffement si mauvais contact)
- Connecteurs thermostat (basse tension, oxydation)

**Symptômes typiques par connecteur :**
- **Sonde température** : valeurs erratiques, erreur sonde
- **Pompe** : démarrage aléatoire, arrêts intempestifs
- **Ventilateur** : F33 intermittent (Saunier Duval)
- **Thermostat** : chaudière ne suit pas demande

---

