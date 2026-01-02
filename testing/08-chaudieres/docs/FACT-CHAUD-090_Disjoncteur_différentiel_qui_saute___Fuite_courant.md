## FACT-CHAUD-090: Disjoncteur différentiel qui saute - Fuite courant

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-090 |
| **Catégorie** | Sécurité Gaz & Sécurités Générales |
| **Système** | Disjoncteur différentiel 30 mA |
| **Gravité** | **Élevée** |
| **Marques** | Multi-marques |

**Symptômes :**
- Disjoncteur différentiel saute au démarrage chaudière
- Ou déclenchement aléatoire en fonctionnement
- Voyant différentiel rouge (disjoncteur en position basse)
- Impossible réenclencher, redéclenche immédiatement
- Ou tient quelques secondes/minutes puis redéclenche
- Parfois au démarrage pompe, vanne gaz, ou allumage

**Cause racine probable :**
Fuite courant vers terre > 30 mA, défaut isolement composant (pompe, résistance, carte), humidité, câblage défectueux, différentiel trop sensible ou défectueux.

**Étapes de résolution :**

1. **Compréhension disjoncteur différentiel**
   - **Rôle** : protection personnes contre électrocution
   - **Principe** : compare courant phase et neutre
   - **Si différence > 30 mA** : fuite vers terre → déclenchement
   - **Obligatoire** : NFC 15-100 (toute installation)
   - **Types** :
     * Type AC : courants alternatifs (standard)
     * Type A : AC + courants continus pulsés (obligatoire plaques, lave-linge)
     * Type F : A + hautes fréquences (informatique)

2. **Localisation défaut (méthode élimination)**
   - **Étape 1** : couper TOUS les disjoncteurs divisionnaires
   - Réenclencher différentiel seul
   - Si tient : défaut sur un circuit
   - Si saute : différentiel défectueux OU défaut permanent installation
   - **Étape 2** : réenclencher circuits un par un
   - Identifier circuit qui fait sauter différentiel
   - Isoler ce circuit pour tests approfondis

3. **Isolation chaudière du réseau**
   - Couper disjoncteur chaudière
   - Débrancher chaudière (prise ou bornier)
   - Réenclencher différentiel
   - **Si tient** : défaut dans chaudière
   - **Si saute** : défaut câblage amont ou autre appareil

4. **Test composants chaudière un par un**
   - Ouvrir boîtier électrique chaudière
   - Débrancher TOUS composants (pompe, vannes, carte, etc.)
   - Laisser uniquement alimentation secteur
   - Réenclencher différentiel :
     * Si saute : défaut câblage ou carte
     * Si tient : reconnecter composants un par un
   - Identifier composant défectueux (fait sauter)

5. **Mesure courant fuite (pince ampèremétrique)**
   - Utiliser pince ampèremétrique AC/DC (sensibilité mA)
   - Entourer fil terre chaudière avec pince
   - Mesurer courant fuite terre
   - **Valeurs** :
     * Normal : < 5 mA
     * Limite : 15-20 mA
     * Déclenchement : > 30 mA
   - Identifier composant source fuite (débrancher un par un)

6. **Mesure isolement composants (mégohmmètre)**
   - Débrancher composant suspect
   - Mesurer isolement phase/terre (500V DC)
   - **Défaut si** : < 100 kΩ (fuite importante)
   - **Composants fréquemment défectueux** :
     * **Pompe circulation** : bobinage humidité/vieillesse
     * **Résistance électrique** : ECS, appoint (fuite gaine)
     * **Transformateur HT** : allumage (humidité)
     * **Électrovanne gaz** : bobine (humidité)
     * **Carte électronique** : condensation, humidité
     * **Câbles chauffants** : anti-gel (fissure isolant)

7. **Contrôle humidité (cause fréquente)**
   - Inspecter visuellement composants
   - Rechercher traces eau, condensation
   - Sources humidité :
     * Fuite eau chaudière (échangeur, joints)
     * Condensation (mauvaise ventilation)
     * Infiltration extérieure (chaudière extérieure)
     * Siphon condensats débordant
   - Sécher composants (soufflette, chaleur douce)
   - Éliminer source humidité

8. **Vérification terre et différentiel**
   - Mesurer résistance terre (< 100 Ω requis)
   - Vérifier connexion terre chaudière (serrage)
   - Tester différentiel (bouton Test) :
     * Doit déclencher instantanément
     * Si ne déclenche pas : différentiel HS (danger)
   - Vérifier calibre différentiel (30 mA pour habitation)

9. **Cas particulier : déclenchement intempestif**
   - Différentiel saute sans défaut réel
   - **Causes** :
     * Différentiel usé (> 15 ans), trop sensible
     * Accumulation micro-fuites (< 30 mA individuelles mais > 30 mA total)
     * Perturbations HF (variateurs, informatique)
     * Foudre, surtension
   - **Solutions** :
     * Remplacer différentiel par neuf
     * Installer différentiel type A ou F (moins sensible HF)
     * Séparer circuits (différentiels multiples)
     * Parafoudre si zone orageuse

10. **Remplacement composant défectueux**
    - Identifier composant (pompe, résistance, etc.)
    - Remplacer par pièce d'origine ou équivalente
    - Vérifier isolement nouveau composant avant montage
    - Reconnecter et tester

11. **Remise service et vérifications**
    - Reconnecter tous composants
    - Mesurer isolement global chaudière (> 500 kΩ)
    - Mesurer courant fuite terre (< 5 mA)
    - Réenclencher différentiel
    - Démarrer chaudière, surveiller 30 minutes
    - Tester tous modes (chauffage, ECS, allumage)
    - Vérifier différentiel ne saute pas

**Prévention :**
- Contrôle isolement composants à l'entretien annuel
- Protection contre humidité (ventilation, étanchéité)
- Remplacement préventif composants > 12-15 ans
- Test mensuel différentiel (bouton Test) par utilisateur
- Remplacement différentiel > 15 ans
- Séchage local chaufferie si humide

**Courants fuite typiques composants :**
| Composant | Fuite normale | Fuite défaut |
|-----------|---------------|--------------|
| Pompe circulation neuve | < 1 mA | > 10 mA |
| Pompe ancienne (10 ans) | 2-5 mA | > 15 mA |
| Carte électronique | < 1 mA | > 5 mA |
| Résistance électrique | < 0,5 mA | > 10 mA |
| Transformateur HT | < 0,5 mA | > 5 mA |
| Installation totale | < 10 mA | > 30 mA |

**Seuils déclenchement différentiel :**
- **Différentiel 30 mA** : usage domestique (protection personnes)
  * Déclenchement : 15-30 mA (norme)
  * Temps : < 300 ms à 30 mA
  * Immunité : 15 mA (ne déclenche pas)
- **Différentiel 300 mA** : usage tertiaire (protection incendie)
- **Différentiel 500 mA** : usage industriel

**Différents types différentiels :**
- **Type AC** : courants alternatifs sinusoïdaux
  * Usage : éclairage, prises basiques
  * Limite : ne détecte pas DC
- **Type A** : AC + courants pulsés continus
  * Usage : lave-linge, plaques induction, variateurs
  * Obligatoire : circuits spécialisés
- **Type F (ou HI/SI)** : Type A + immunisé HF
  * Usage : informatique, onduleurs, surgélateurs
  * Moins de déclenchements intempestifs

**Diagnostic rapide selon moment déclenchement :**
- **Au démarrage chaudière** : fuite permanente composant
- **À l'allumage brûleur** : transformateur HT ou électrode
- **Au démarrage pompe** : pompe défectueuse
- **Aléatoire fonctionnement** : humidité, composant intermittent
- **Après pluie** : infiltration eau, défaut étanchéité
- **Temps froid** : condensation, givre

**⚠️ SÉCURITÉ :**
- Différentiel 30 mA = PROTECTION VITALE anti-électrocution
- Ne JAMAIS shunter ou neutraliser différentiel
- Déclenchement répété = DÉFAUT RÉEL (recherche obligatoire)
- Différentiel qui ne déclenche pas au test = DANGER (remplacer)
- Absence différentiel = installation NON CONFORME (hors-la-loi)
- Risque mortel électrocution si défaut non réparé

**Obligations réglementaires (NFC 15-100) :**
- Différentiel 30 mA OBLIGATOIRE (toute installation domestique)
- Test fonctionnement mensuel (bouton Test)
- Vérification installation tous les 10 ans (diagnostic immobilier)
- Remplacement différentiel si défectueux
- Attestation conformité Consuel (installation neuve/rénovation)

**Outils diagnostic :**
- **Pince ampèremétrique mA** : mesure fuite terre
- **Mégohmmètre 500V** : mesure isolement
- **Multimètre** : continuité, résistance
- **Testeur différentiel** : simulation fuite contrôlée
- **Contrôleur installation** : mesure complète (terre, différentiel, isolement)

---

