## FACT-CHAUD-085: Limiteur de température de sécurité

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-085 |
| **Catégorie** | Sécurité Gaz & Sécurités Générales |
| **Système** | Limitation température |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Température départ plafonnée à 80-85°C malgré demande supérieure
- Brûleur s'arrête avant atteinte consigne
- Code erreur limitation température
- Puissance bridée automatiquement
- Cycles marche/arrêt fréquents

**Cause racine probable :**
Limiteur de température activé (normal ou défaut), sonde température défectueuse, régulation mal paramétrée, débit eau insuffisant, dimensionnement inadapté.

**Étapes de résolution :**

1. **Comprendre le limiteur de température**
   - **Rôle** : empêcher température excessive (sécurité)
   - **Types** :
     * Limiteur logiciel (carte électronique) : 80-85°C
     * Limiteur physique (thermostat) : 90-95°C
     * Limiteur sécurité (klixon) : 100-110°C
   - **Fonctionnement** : coupe ou réduit puissance si T > seuil
   - **Réarmement** : automatique (logiciel) ou manuel (klixon)

2. **Diagnostic comportement**
   - Afficher température départ (menu diagnostic)
   - Noter température limitation constatée
   - Comparer avec consigne demandée
   - Vérifier si limitation permanente ou temporaire
   - Consulter codes erreur ou historique

3. **Vérification consigne et paramètres**
   - Consigne chauffage raisonnable :
     * Radiateurs haute température : 70-80°C
     * Radiateurs basse température : 50-60°C
     * Plancher chauffant : 35-45°C
   - Vérifier paramètre température maximale carte
   - Ajuster si nécessaire selon type émetteurs
   - Respecter limitations constructeur

4. **Contrôle sonde température départ**
   - Mesurer résistance sonde (CTN/NTC)
   - Comparer courbe température/résistance constructeur
   - Exemple typique :
     * 25°C : 10 kΩ
     * 50°C : 3,3 kΩ
     * 80°C : 1,2 kΩ
   - Remplacer sonde si valeur erronée (limitation intempestive)

5. **Contrôle circulation et débit**
   - Débit insuffisant = température excessive locale
   - Vérifier pompe : débit adapté (réglage vitesse)
   - Contrôler vannes : toutes ouvertes
   - Purger air circuit
   - Vérifier filtre pas colmaté
   - Mesurer ΔT départ-retour :
     * Normal : 15-20°C
     * Si > 25°C : débit insuffisant

6. **Vérification dimensionnement**
   - Chaudière surdimensionnée :
     * Cycles courts
     * Température monte trop vite
     * Limitation fréquente
   - Solutions :
     * Réduire puissance maxi (paramètre)
     * Améliorer inertie (ballon tampon)
     * Installer régulation climatique

7. **Contrôle limiteur physique (thermostat)**
   - Localiser thermostat limiteur (95-100°C)
   - Vérifier pas de déclenchement intempestif
   - Contrôler position (bon contact thermique)
   - Tester continuité électrique
   - Remplacer si défectueux

8. **Cas particulier ECS (Eau Chaude Sanitaire)**
   - Limitation ECS : 60°C (anti-légionellose mais anti-brûlure)
   - Cycle anti-légionellose : 1×/semaine à 65-70°C
   - Vérifier paramètre température ECS
   - Contrôler sonde ballon ECS
   - Ajuster selon besoins et réglementation

**Prévention :**
- Réglage consignes adaptées aux émetteurs
- Vérification sondes température annuelle
- Contrôle circulation et débit
- Paramétrage correct puissance chaudière
- Dimensionnement adapté dès l'installation

**Températures maximales réglementaires :**
- **Chauffage collectif** : 90°C max (DTU 65.11)
- **Chauffage individuel** : 110°C max (limiteur sécurité)
- **ECS distribution** : 60°C max (anti-brûlure)
- **ECS stockage/cycle** : 65°C (anti-légionellose)
- **Plancher chauffant** : 50°C absolu max (déformation)

**Avantages limitation température :**
- Protection échangeur (durée de vie)
- Sécurité utilisateurs (anti-brûlure)
- Économies énergie (rendement)
- Confort (régulation stable)
- Conformité réglementaire

**Réglage optimal selon émetteurs :**
| Type émetteur | T° départ | T° limite |
|---------------|-----------|-----------|
| Plancher chauffant | 35-45°C | 50°C |
| Radiateurs BT | 50-60°C | 70°C |
| Radiateurs HT | 70-80°C | 90°C |
| ECS | 50-60°C | 60°C |

**⚠️ IMPORTANT :**
- Ne jamais désactiver limiteur de température
- Limitation fréquente = signal dysfonctionnement
- Consigne excessive = surconsommation inutile
- Température trop haute = inconfort (surchauffe)
- Respecter toujours les valeurs constructeur

---

