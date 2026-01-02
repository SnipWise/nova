## FACT-CHAUD-082: Vanne gaz de sécurité fermée - Chaudière bloquée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-082 |
| **Catégorie** | Sécurité Gaz & Sécurités Générales |
| **Système** | Vanne gaz sécurité |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Chaudière bloquée, pas de tentative d'allumage
- Code erreur "défaut gaz" ou "vanne gaz"
- Étincelle présente mais aucune flamme
- Bruit caractéristique d'ouverture vanne absent
- Vanne reste fermée malgré appel chaleur

**Cause racine probable :**
Coupure électrique vanne, bobine vanne défectueuse, vanne bloquée mécaniquement, carte électronique défaillante, sécurité amont déclenchée, pressostat gaz ouvert.

**Étapes de résolution :**

1. **Diagnostic initial**
   - Vérifier que le robinet gaz manuel est bien ouvert
   - Contrôler alimentation électrique chaudière (disjoncteur)
   - Écouter lors démarrage (clic ouverture vanne)
   - Consulter code erreur affiché
   - Vérifier pression gaz au manomètre

2. **Identification type vanne gaz**
   - **Vanne simple** : 1 bobine, tout ou rien
   - **Vanne double** : 2 bobines (sécurité redondante)
   - **Vanne modulante** : servomoteur + bobines sécurité
   - Localiser le modèle (plaque signalétique)
   - Identifier les connexions électriques

3. **Contrôle alimentation électrique vanne**
   - Mesurer tension aux bornes bobine 1 (phase démarrage)
   - Tension attendue : 230V AC
   - Mesurer tension bobine 2 si présente
   - Vérifier la connectique (cosses bien enfichées)
   - Contrôler câblage (pas de coupure, brûlure)

4. **Test bobines vanne gaz**
   - Couper alimentation électrique
   - Débrancher connecteurs bobines
   - Mesurer résistance chaque bobine (multimètre Ω)
   - Valeur attendue : 3-5 kΩ selon modèle (voir notice)
   - Si résistance infinie : bobine coupée (HS)
   - Si résistance nulle : bobine court-circuitée (HS)

5. **Contrôle mécanique vanne**
   - Fermer robinet gaz en amont
   - Déposer la vanne gaz (précautions gaz)
   - Vérifier absence blocage mécanique (corps étranger)
   - Contrôler état membrane/clapet (pas de déformation)
   - Vérifier ressort rappel (pas cassé)
   - Nettoyer filtres intégrés si présents

6. **Contrôle préssostat gaz (si équipé)**
   - Localiser pressostat gaz (sur vanne ou rampe)
   - Vérifier continuité électrique (contact fermé au repos)
   - Mesurer pression gaz : doit dépasser seuil pressostat
   - Seuil typique : 15-18 mbar pour commutation
   - Ajuster ou remplacer si défectueux

7. **Contrôle carte électronique**
   - Vérifier que carte commande l'ouverture vanne
   - Mesurer tension sortie carte (230V en phase allumage)
   - Consulter menu diagnostic (état vanne)
   - Vérifier fusible protection vanne sur carte
   - Réinitialiser carte électronique

8. **Remplacement vanne gaz**
   - Identifier référence exacte (plaque vanne)
   - Exemple marques : Honeywell, SIT, Dungs
   - Fermer robinet gaz, déposer ancienne vanne
   - Installer vanne neuve avec joints neufs
   - Respecter sens de montage (flèche)
   - Serrage modéré (pas de déformation)
   - Contrôler étanchéité (eau savonneuse)

9. **Remise en service**
   - Rouvrir gaz progressivement
   - Purger l'air du circuit
   - Démarrer chaudière
   - Vérifier ouverture vanne (bruit clic)
   - Contrôler allumage correct
   - Mesurer pression gaz brûleur (selon notice)

**Prévention :**
- Test fonctionnement vanne à chaque entretien (écoute)
- Vérification pression gaz annuelle
- Contrôle connectiques électriques
- Remplacement préventif vanne > 15 ans
- Nettoyage filtres gaz

**Durée de vie vanne gaz :**
- Durée moyenne : 10-15 ans
- Selon nombre de cycles : 100 000-500 000 cycles
- Signes usure : bruits anormaux, hésitations ouverture
- Remplacement préventif recommandé à 12 ans

**Références vannes courantes :**
- **Honeywell VK4105** : vanne modulante
- **SIT 845 Sigma** : vanne double sécurité
- **Dungs MB-DLE** : vanne modulante professionnelle
- Toujours utiliser pièce d'origine constructeur

**⚠️ SÉCURITÉ :**
- Ne jamais forcer mécaniquement une vanne gaz
- Ne jamais shunter les bobines (risque explosion)
- Contrôler étanchéité après toute intervention
- Intervention qualifiée PGN/PGP obligatoire
- Vanne gaz = organe de sécurité critique

---

