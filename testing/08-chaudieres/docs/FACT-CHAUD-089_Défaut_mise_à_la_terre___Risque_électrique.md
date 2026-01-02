## FACT-CHAUD-089: Défaut mise à la terre - Risque électrique

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-089 |
| **Catégorie** | Sécurité Gaz & Sécurités Générales |
| **Système** | Mise à la terre / Protection électrique |
| **Gravité** | **Critique** |
| **Marques** | Multi-marques |

**Symptômes :**
- Picotements au toucher chaudière (fuite courant)
- Disjoncteur différentiel saute au démarrage
- Code erreur électrique sur afficheur
- Chaudière ne démarre pas
- Tension mesurée entre chaudière et terre
- Corrosion accélérée échangeur (électrolyse)

**Cause racine probable :**
Absence terre, terre déconnectée, terre insuffisante (résistance élevée), défaut isolement composant (fuite masse), câblage défectueux, humidité carte électronique.

**Étapes de résolution :**

1. **SÉCURITÉ IMMÉDIATE**
   - ⚠️ DANGER : risque électrocution mortelle
   - Si picotements au toucher : NE PAS toucher chaudière
   - COUPER disjoncteur général tableau électrique
   - CONDAMNER disjoncteur (cadenas, affichage)
   - Utiliser EPI isolants si intervention (gants, tapis)
   - Intervention par électricien qualifié si doute

2. **Vérification présence mise à la terre**
   - Ouvrir boîtier électrique chaudière
   - Identifier fil terre : vert/jaune (bicolore)
   - Vérifier connexion terre sur borne dédiée
   - Contrôler serrage connexion (bien serré)
   - Vérifier continuité fil terre jusqu'à prise

3. **Mesure résistance terre (testeur terre)**
   - Utiliser testeur résistance terre (ohmmètre spécifique)
   - Méthode 3 points (piquets)
   - **Valeurs résistance terre** :
     * Excellent : < 10 Ω
     * Bon : 10-30 Ω
     * Acceptable : 30-100 Ω
     * Insuffisant : > 100 Ω (à améliorer)
     * Limite réglementaire : < 100 Ω
   - Si > 100 Ω : amélioration terre nécessaire

4. **Vérification liaison équipotentielle**
   - Relier toutes masses métalliques entre elles
   - Chaudière reliée à barrette terre
   - Tuyauteries métalliques reliées (si conductrices)
   - Vérifier continuité électrique (multimètre)
   - Résistance < 2 Ω entre masses

5. **Contrôle isolement composants (mégohmmètre)**
   - ⚠️ Couper alimentation, débrancher connecteurs
   - Mesurer isolement phase/terre (mégohmmètre 500V)
   - **Valeurs isolement** :
     * Neuf : > 1 MΩ (1 000 000 Ω)
     * Bon : > 500 kΩ
     * Limite : 100 kΩ
     * Défaut : < 100 kΩ (fuite masse)
   - Identifier composant défectueux (résistance, pompe, vanne, transformateur)

6. **Recherche défaut isolement méthodique**
   - Débrancher tous composants un par un
   - Mesurer isolement phase/terre après chaque débranchement
   - Identifier composant en défaut
   - **Composants fréquemment défectueux** :
     * Résistance électrique (ECS, appoint)
     * Pompe circulation (bobinage HS)
     * Transformateur allumage (humidité)
     * Électrovanne (bobine)
     * Carte électronique (humidité, condensation)

7. **Contrôle humidité carte électronique**
   - Humidité = principale cause fuite masse
   - Inspecter visuellement carte (traces humidité, corrosion)
   - Rechercher condensation (température, ventilation)
   - Sécher carte (soufflette air sec, chaleur douce)
   - Protéger carte (vernis tropicalisation si récurrent)
   - Améliorer ventilation local (éviter condensation)

8. **Vérification câblage électrique**
   - Contrôler gaine câbles (pas d'écrasement, coupure)
   - Vérifier isolant fils (pas de craquelure, brûlure)
   - Contrôler connexions (serrage, oxydation)
   - Vérifier passage câbles (pas de frottement, arête vive)
   - Remplacer câbles défectueux

9. **Amélioration prise terre si insuffisante**
   - **Solutions** :
     * Ajouter piquet(s) terre supplémentaire(s)
     * Installer boucle fond fouille (fil cuivre nu 25 mm²)
     * Améliorer humidification sol (bentonite)
     * Relier plusieurs piquets (résistances en parallèle)
   - Objectif : résistance < 30 Ω (idéal < 10 Ω)
   - Faire appel électricien qualifié

10. **Contrôle différentiel 30 mA**
    - Vérifier présence différentiel 30 mA (obligatoire)
    - Tester fonctionnement (bouton Test)
    - Doit déclencher < 30 mA fuite
    - Remplacer si ne déclenche pas au test
    - Type AC ou A selon installation

11. **Remise en service et tests**
    - Reconnecter tous composants
    - Mesurer isolement global (> 500 kΩ)
    - Réenclencher disjoncteur différentiel
    - Démarrer chaudière sous surveillance
    - Mesurer tension chaudière/terre (0 V attendu)
    - Vérifier différentiel ne saute pas

**Prévention :**
- Vérification terre annuelle (continuité, serrage)
- Mesure résistance terre tous les 5 ans
- Contrôle isolement composants à l'entretien
- Protection contre humidité (ventilation)
- Remplacement préventif composants > 15 ans
- Détection fuite masse (contrôleur permanent)

**Réglementation électrique (NFC 15-100) :**
- Mise à la terre OBLIGATOIRE (tous appareils classe I)
- Résistance terre : < 100 Ω
- Différentiel 30 mA OBLIGATOIRE (protection personnes)
- Section fil terre : ≥ phase (min 6 mm² cuivre)
- Liaison équipotentielle salle eau OBLIGATOIRE
- Couleur terre : vert/jaune EXCLUSIVEMENT

**Dangers absence/défaut terre :**
- **Électrocution** : tension dangereuse sur carcasse (230V)
- **Incendie** : arc électrique, échauffement
- **Électrolyse** : corrosion accélérée échangeur (courant parasite)
- **Dysfonctionnement** : carte électronique perturbée
- **Non-déclenchement différentiel** : défaut non détecté

**Électrolyse échangeur (courant parasite) :**
- Cause : différence potentiel eau/échangeur
- Conséquence : corrosion rapide (perforation < 1 an possible)
- Signes : taches corrosion localisées, perforation
- Prévention : terre correcte + anode sacrificielle (si prévu)

**Classes protection électrique :**
- **Classe I** : mise terre obligatoire (chaudières murales)
- **Classe II** : double isolation, terre non requise (rare)
- **Classe III** : très basse tension (< 50V)

**⚠️ DANGER MORTEL :**
- Courant > 30 mA : fibrillation cardiaque, décès
- Tension 230V : mortelle si contact prolongé
- Chaudière non reliée terre : carcasse peut être sous tension
- Eau + électricité : conductivité maximale (salle bain)
- JAMAIS intervenir sous tension sans EPI et compétences

**EPI électricien :**
- Gants isolants classe 0 (500V) ou 00 (1000V)
- Tapis isolant
- Chaussures sécurité isolantes
- Outillage isolé 1000V
- VAT (vérificateur absence tension)

---

