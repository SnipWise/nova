## FACT-CHAUD-100: Mise à jour firmware nécessaire

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-100 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Logiciel embarqué (firmware) |
| **Gravité** | Faible à Moyenne |
| **Marques** | Multi-marques (surtout récentes) |

**Symptômes :**
- Bugs répétés, comportements illogiques
- Erreurs non justifiées (composants OK)
- Incompatibilité avec nouveau thermostat/régulation
- Fonctionnalités manquantes documentées notice
- Code erreur "Software Error" ou similaire
- Problèmes résolus par mise à jour selon constructeur

**Cause racine probable :**
Firmware obsolète, bugs connus corrigés version ultérieure, évolution normes (communication, sécurité), nouvelles fonctionnalités ajoutées, incompatibilité matérielle.

**Étapes de résolution :**

1. **Identification version firmware actuelle**
   - Accéder menu diagnostic/service
   - Chercher information version :
     - "Software version", "Firmware", "SW Ver"
     - Exemple : "v2.15", "FW 03.04.02"
   - Noter version exacte
   - Noter aussi : modèle chaudière, référence carte

2. **Vérification disponibilité mise à jour**
   - Consulter site web constructeur (espace pro SAV)
   - Rechercher firmware pour modèle exact
   - Vérifier notes de version (changelog) :
     - Corrections de bugs
     - Nouvelles fonctionnalités
     - Améliorations performances
   - Comparer version installée vs disponible

3. **Types de mise à jour selon marque**
   - **Mise à jour USB** (Vaillant, Saunier Duval, De Dietrich récents) :
     - Télécharger fichier firmware (.bin, .hex, .upd)
     - Copier sur clé USB (FAT32, racine)
     - Insérer clé USB dans port chaudière
     - Suivre procédure affichée ou menu service
   - **Mise à jour via outil SAV** (Frisquet Vision, Atlantic) :
     - Nécessite outil/logiciel propriétaire constructeur
     - Connexion PC ↔ chaudière (câble spécifique)
     - Lancer logiciel SAV
     - Suivre procédure guidée
   - **Remplacement EPROM** (anciennes chaudières) :
     - Composant mémoire (puce) à dessouder/remplacer
     - Commander EPROM avec nouveau firmware
     - Dessouder ancienne, souder nouvelle
     - Ou support EPROM (socket) si présent

4. **Préparation mise à jour**
   - **Sauvegarder paramètres** :
     - Photographier tous écrans paramètres
     - Noter sur papier : puissance, type gaz, courbe chauffe, etc.
     - Certaines MAJ effacent paramètres
   - **Vérifier alimentation stable** :
     - Chaudière connectée secteur stable
     - Pas de risque coupure pendant MAJ (risque brick)
   - **Charge batterie** (si applicable) :
     - Chaudière chaude, batterie auxiliaire chargée

5. **Procédure mise à jour USB (type Vaillant)**
   - Télécharger firmware exact modèle chaudière
   - Décompresser si fichier .zip
   - Formater clé USB en FAT32
   - Copier fichier firmware à la racine (pas dans dossier)
   - Éteindre chaudière
   - Insérer clé USB dans port (généralement sous capot, sur carte)
   - Allumer chaudière
   - Menu affiche "Update available" ou similaire
   - Valider mise à jour (bouton OK ou séquence boutons)
   - Patienter (5-15 minutes, ne PAS couper alimentation)
   - Redémarrage automatique après MAJ
   - Retirer clé USB

6. **Procédure mise à jour logiciel SAV**
   - Installer logiciel SAV sur PC (fourni constructeur)
   - Connecter câble PC ↔ chaudière :
     - USB vers connecteur carte
     - Ou adaptateur série/USB
   - Lancer logiciel
   - Détecter chaudière (communication établie)
   - Lire version actuelle firmware
   - Charger fichier firmware (.hex, .bin)
   - Lancer mise à jour (bouton "Update", "Flash")
   - Suivre progression (barre %)
   - Attendre fin (ne pas déconnecter)
   - Redémarrage chaudière

7. **Vérification après mise à jour**
   - Vérifier nouvelle version firmware (menu diagnostic)
   - Contrôler paramètres :
     - Si effacés : reprogrammer (voir notes sauvegarde)
     - Type chaudière, puissance, type gaz, etc.
   - Lancer auto-diagnostic si disponible
   - Tester fonctionnement :
     - Demande chauffage : OK
     - Demande ECS : OK
     - Modulation : OK
     - Sondes : valeurs cohérentes

8. **Résolution problèmes MAJ**
   - **MAJ échoue / erreur** :
     - Vérifier fichier firmware correspond exactement au modèle
     - Reformater clé USB en FAT32
     - Essayer autre clé USB (compatibilité)
     - Vérifier connexion câble (logiciel SAV)
   - **Chaudière brickée** (ne démarre plus) :
     - Tenter re-flash avec ancien firmware
     - Ou mode recovery (selon marque, séquence boutons spéciale)
     - Contacter SAV constructeur (intervention technicien)
   - **Paramètres effacés** :
     - Reprogrammer selon notes sauvegarde
     - Lancer assistant première mise en service

9. **Cas particuliers par marque**
   - **Vaillant ecoTEC** :
     - MAJ via clé USB ou logiciel DIALOGplus
     - Port USB sous capot avant
   - **Saunier Duval** :
     - MAJ via clé USB (modèles récents)
     - Fichier .upd sur clé FAT32
   - **Frisquet Eco Radio Vision** :
     - MAJ via logiciel Frisquet SAV
     - Connexion USB → module Vision
   - **De Dietrich** :
     - MAJ via logiciel Diematic SAV
     - Certains modèles : carte SD
   - **Elm Leblanc** :
     - MAJ rare, souvent remplacement carte
   - **Atlantic** :
     - Logiciel Cozytouch Pro (modèles connectés)

10. **Quand est-ce nécessaire ?**
    - Bugs répétés non résolus par interventions matérielles
    - Code erreur "Update required" affiché
    - Installation nouveau thermostat incompatible avec firmware ancien
    - Constructeur publie MAJ corrective pour problème connu
    - Évolution normes (ex: ErP, éco-conception)

**Prévention :**
- Vérifier firmware à chaque entretien annuel
- S'abonner newsletter SAV constructeur (alertes MAJ)
- Conserver clé USB + fichiers firmware (stock préventif)
- Documenter versions firmware installations

**Risques :**
- **Brick** (carte inutilisable) si :
  - Coupure alimentation pendant MAJ
  - Mauvais fichier firmware
  - Erreur procédure
- Toujours sauvegarder paramètres avant MAJ
- Ne jamais interrompre mise à jour en cours

**Coût :**
- Mise à jour firmware : gratuite (fichier)
- Logiciel SAV : gratuit (constructeurs) ou payant (selon marque)
- Câble interface : 20-100€ (si non fourni)
- Intervention technicien SAV : 100-200€ (si nécessaire assistance)

**Outils nécessaires :**
- Clé USB (FAT32, 1-4 Go suffisant)
- Ou PC + logiciel SAV + câble interface
- Documentation procédure (notice SAV)
- Notes paramètres chaudière

**Ressources :**
- Sites constructeurs (espace pro/SAV)
- Forums chauffagistes (partage expériences MAJ)
- SAV constructeur (assistance téléphonique)

---

