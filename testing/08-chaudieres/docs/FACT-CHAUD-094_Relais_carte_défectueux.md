## FACT-CHAUD-094: Relais carte défectueux

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-094 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Relais commande charges |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Pompe ne démarre pas ou tourne en permanence
- Ventilateur ne s'active pas ou reste actif
- Vanne gaz ne s'ouvre pas
- Bruit de claquement répété sans effet
- Brûleur ne démarre pas malgré demande

**Cause racine probable :**
Relais collé (contacts soudés), bobine relais grillée, contacts usés/oxydés, surtension ayant endommagé relais, vieillissement mécanique.

**Étapes de résolution :**

1. **Identification relais défectueux**
   - Localiser les relais sur carte (composants rectangulaires transparents/bleus)
   - Identification marquage : K1, K2, K3, RLY1, etc.
   - Fonctions courantes :
     - Relais pompe (marche/arrêt circulateur)
     - Relais ventilateur
     - Relais vanne gaz (1 ou 2 relais)
     - Relais brûleur/allumage
   - Observer LED associée si présente

2. **Test auditif relais**
   - Remettre chaudière sous tension
   - Lancer demande chauffage
   - Écouter "clic" relais à chaque activation
   - Absence de "clic" : bobine relais HS ou commande absente
   - "Clic" présent mais pas d'effet : contacts défectueux

3. **Test visuel relais**
   - Couper alimentation
   - Déposer carte pour accès relais
   - Vérifier état visuel relais :
     - Traces de brûlure, noircissement
     - Déformation plastique
     - Traces d'arc électrique
   - Si doute : dessouder et tester hors carte

4. **Test électrique relais**
   - **Test bobine** (relais déposé ou in-situ) :
     - Mesurer résistance bobine : généralement 50-500 Ω
     - Si infini (∞) : bobine coupée, relais HS
     - Si 0 Ω : bobine court-circuitée, relais HS
   - **Test contacts** (relais déposé) :
     - Identifier bornes : commun (C), NO (normalement ouvert), NC (normalement fermé)
     - Mesurer continuité au repos : C-NC fermé, C-NO ouvert
     - Alimenter bobine (12V ou 24V selon specs) : inversion
     - Si pas d'inversion : contacts grippés/soudés

5. **Diagnostic contacts collés**
   - Symptôme : charge fonctionne en permanence
   - Cause : arc électrique a soudé les contacts
   - Test : couper alimentation, mesurer continuité contacts NO (doit être ouvert)
   - Si fermé au repos : relais collé, remplacement obligatoire

6. **Diagnostic bobine grillée**
   - Symptôme : pas de "clic", charge ne démarre jamais
   - Mesurer tension commande bobine (carte sous tension, demande active)
   - Si tension présente (12V ou 24V) mais pas de "clic" : bobine HS
   - Si pas de tension : problème carte électronique (transistor commande)

7. **Remplacement relais**
   - Noter références exactes relais :
     - Tension bobine : 12VDC, 24VDC, etc.
     - Courant contacts : 10A, 16A, etc.
     - Configuration : SPDT (1RT), DPDT (2RT)
     - Exemple : OMRON G5LE, FINDER 40.52, SONGLE SRD
   - Dessouder ancien relais (pompe à dessouder ou tresse)
   - Souder nouveau relais (orientation correcte !)
   - Nettoyer flux soudure

8. **Alternative réparation temporaire**
   - **Contacts oxydés/sales** (si relais accessible) :
     - Démonter capot relais (clipsé)
     - Nettoyer contacts avec papier abrasif très fin ou contact cleaner
     - Remonter et tester
     - Solution temporaire : prévoir remplacement
   - **Relais collé** :
     - Tapoter légèrement sur relais (peut décrocher contacts)
     - Solution très temporaire, remplacement urgent

9. **Test après remplacement**
   - Remonter carte
   - Reconnecter charges
   - Mettre sous tension
   - Tester activation relais : écouter "clic"
   - Vérifier fonctionnement charge (pompe, ventilateur)
   - Mesurer tension sortie relais

**Prévention :**
- Remplacement préventif relais après 8-10 ans (selon sollicitation)
- Éviter surtensions (parafoudre)
- Vérifier charges (pompe, ventilateur) pour éviter surcourant
- Protéger relais avec varistance ou RC snubber si charges inductives

**Relais courants chaudières :**
- Relais 12VDC 10A SPDT : pompe, ventilateur
- Relais 24VDC 16A SPDT : charges puissantes
- Montage sur circuit imprimé (PCB)
- Disponibles en commerce électronique

**Compétences requises :**
- Soudure composants traversants
- Lecture schéma électrique
- Mesures électriques multimètre
- Si pas compétent : remplacement carte complète

---

