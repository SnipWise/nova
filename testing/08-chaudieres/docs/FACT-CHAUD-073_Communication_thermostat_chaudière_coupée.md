## FACT-CHAUD-073: Communication thermostat/chaudière coupée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-073 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Communication bus (OpenTherm, eBUS, radio) |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (systèmes modulants) |

**Symptômes :**
- Message erreur communication sur thermostat ou chaudière
- Thermostat connecté affiche "perte connexion"
- Chaudière fonctionne en mode dégradé (température fixe)
- Pas de retour info chaudière sur thermostat (température, pression)
- Pictogramme communication absent/barré

**Cause racine probable :**
Problème câblage bus, mauvaise polarité, perturbation électromagnétique, défaut alimentation, incompatibilité protocole, défaut module communication.

**Étapes de résolution :**

1. **Identification type communication**
   - **OpenTherm :** standard européen, 2 fils non polarisés
   - **eBUS :** Vaillant/Saunier Duval, 2 fils polarisés
   - **Propriétaire :** Frisquet Eco-Radio, Atlantic, etc.
   - **Radio 868 MHz :** sans fil, récepteur chaudière
   - **WiFi :** connecté internet, box/passerelle

2. **Diagnostic radio/WiFi**
   - **Thermostat radio :**
     - Vérifier piles émetteur (remplacer)
     - Contrôler LED récepteur (doit clignoter réception)
     - Vérifier distance émetteur/récepteur (< 30m général)
     - Contrôler obstacles (murs béton, métal)
     - Ré-appairer émetteur et récepteur (procédure constructeur)
   - **Thermostat WiFi :**
     - Vérifier connexion WiFi thermostat (SSID, mot de passe)
     - Contrôler signal WiFi emplacement thermostat
     - Vérifier connexion internet box
     - Redémarrer box internet et thermostat
     - Reconnecter application mobile

3. **Diagnostic bus filaire (OpenTherm/eBUS)**
   - Vérifier continuité 2 fils bus (multimètre)
   - Contrôler absence court-circuit fils entre eux ou masse
   - Vérifier serrage bornes thermostat et chaudière
   - Mesurer tension bus (doit varier : OpenTherm 0-40V, eBUS ~15-24V)
   - Longueur max câble : généralement 50m (vérifier notice)

4. **Contrôle polarité (eBUS)**
   - eBUS est polarisé : + et - à respecter
   - Inverser polarité si pas de communication
   - Repérer fils : généralement rouge (+) et noir/bleu (-)
   - Vérifier schéma notice chaudière et thermostat

5. **Perturbations électromagnétiques**
   - Câble bus éloigné câbles puissance (> 10 cm)
   - Pas de câble bus dans même gaine que 230V
   - Éloigner transformateurs, variateurs lumière
   - Utiliser câble blindé si environnement perturbé
   - Raccorder blindage à masse unique (éviter boucles)

6. **Contrôle alimentation**
   - Vérifier alimentation chaudière (secteur OK)
   - Contrôler alimentation piles thermostat
   - Sur thermostat alimenté par bus : vérifier tension fournie
   - Certains thermostats nécessitent alimentation séparée

7. **Compatibilité protocole**
   - Vérifier compatibilité thermostat/chaudière (notice)
   - OpenTherm : standard mais versions différentes (v2.2, v3.0)
   - eBUS : spécifique Vaillant/Saunier Duval
   - Certains thermostats multi-protocoles (réglage à effectuer)
   - Consulter liste compatibilité constructeur

8. **Réinitialisation et association**
   - **Réinitialiser thermostat :** procédure constructeur (reset usine)
   - **Réinitialiser module communication chaudière**
   - **Ré-appairer :**
     - Thermostat radio : mode appairage simultané
     - Thermostat WiFi : reconfiguration réseau
     - Bus filaire : généralement auto-détection
   - Patienter 2-5 minutes après association (initialisation)

9. **Diagnostic module communication**
   - Vérifier LED module récepteur chaudière
   - LED verte fixe : OK
   - LED rouge ou clignotante : erreur
   - Tester avec autre thermostat compatible si disponible
   - Remplacer module si défectueux

10. **Mode dégradé**
    - En attendant réparation : chaudière en mode standalone
    - Paramétrer température départ fixe (menu chaudière)
    - Ou installer thermostat TOR filaire temporaire

**Prévention :**
- Vérification annuelle communication (test)
- Remplacement préventif piles
- Contrôle câblage lors travaux
- Mise à jour firmware si disponible
- Protection câble bus (gaine, chemin protégé)

**Dépannage rapide selon symptôme :**

| Symptôme | Cause probable | Action |
|----------|----------------|--------|
| Perte liaison radio intermittente | Piles faibles | Remplacer piles |
| Perte liaison radio permanente | Désappairage | Ré-appairer |
| Pas de liaison bus filaire | Câble coupé ou polarité | Vérifier continuité et polarité |
| Communication OK puis perdue | Perturbation EMI | Éloigner câbles puissance |
| Liaison intermittente bus | Mauvais contact | Resserrer bornes |

**Avantages communication modulante (vs TOR) :**
- Régulation précise (modulation puissance)
- Retour infos chaudière (température, pression, erreurs)
- Optimisation rendement
- Programmation avancée
- Statistiques consommation

---

