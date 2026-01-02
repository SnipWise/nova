## FACT-CHAUD-076: Anticipation surchauffe

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-076 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Régulation température |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Température départ dépasse largement consigne
- Température intérieure atteint consigne puis continue à monter
- Chaudière ne s'arrête pas à temps
- Surchauffe de 2-5°C après arrêt chaudière
- Inconfort thermique (trop chaud)

**Cause racine probable :**
Inertie système non compensée, régulation trop lente, paramètres PID inadaptés, absence d'anticipation, sur-dimensionnement chaudière, mauvaise régulation hydraulique.

**Étapes de résolution :**

1. **Compréhension phénomène**
   - **Inertie thermique :** système continue à chauffer après arrêt
   - Sources inertie :
     - Masse eau circuit (volume important)
     - Masse émetteurs (radiateurs fonte)
     - Masse bâtiment (murs, dalles)
   - Chaleur stockée est restituée → dépassement consigne

2. **Mesure dépassement**
   - Noter température consigne (ex: 20°C)
   - Mesurer température max atteinte (ex: 23°C)
   - Calculer dépassement (ici: +3°C)
   - Observer temps retour à consigne après arrêt
   - Identifier si problème ponctuel ou systématique

3. **Contrôle type régulation**
   - **Régulation TOR simple :** pas d'anticipation
     - Marche/arrêt basique
     - Dépassement normal (1-2°C acceptable)
     - Si > 2°C : problème autre
   - **Régulation PID :** anticipation possible
     - Paramètres P, I, D à ajuster
     - Peut limiter dépassement à 0,5°C
   - **Régulation climatique :** anticipation météo
     - Devrait éviter dépassements importants

4. **Optimisation régulation TOR**
   - **Différentiel/hystérésis :**
     - Écart entre seuil marche et arrêt
     - Exemple : arrêt à 20°C, redémarrage à 19°C (différentiel 1°C)
     - Si dépassement : réduire différentiel (attention cycles courts)
   - **Anticipation arrêt :**
     - Certains thermostats : fonction anticipation
     - Arrêt avant consigne (ex: -0,5°C) pour compenser inertie
     - Activer/ajuster si disponible

5. **Ajustement régulation PID**
   - **Paramètre P (Proportionnel) :**
     - Réduit puissance à l'approche consigne
     - Augmenter P si dépassement (action plus forte)
   - **Paramètre I (Intégral) :**
     - Corrige erreur résiduelle
     - Augmenter I si oscillations
   - **Paramètre D (Dérivé) :**
     - Anticipe évolution
     - Augmenter D pour limiter dépassement
   - **Attention :** ajustements par petits pas, observer 24-48h entre modifications

6. **Contrôle dimensionnement chaudière**
   - Chaudière surdimensionnée : montée rapide température
   - Vérifier puissance chaudière vs besoins
   - Si > 150% besoins : risque surchauffe
   - Solutions :
     - Réduire puissance max chaudière (menu)
     - Améliorer régulation (modulante)
     - Installer ballon tampon (stockage, amortissement)

7. **Optimisation hydraulique**
   - **Débit circulation :**
     - Débit excessif : transport rapide chaleur, dépassement
     - Réduire vitesse pompe (si réglable)
     - Viser ΔT départ-retour 15-20°C
   - **Bypass :**
     - Vérifier réglage bypass
     - Bypass trop ouvert : débit réduit émetteurs, accumulation chaleur chaudière

8. **Solutions émetteurs**
   - **Vannes thermostatiques radiateurs :**
     - Installer si absentes
     - Fermeture progressive à l'approche consigne
     - Limite dépassement local
   - **Têtes thermostatiques anticipation :**
     - Détectent montée température
     - Ferment avant atteinte consigne
   - **Régulation pièce par pièce :**
     - Évite surchauffe globale

9. **Inertie bâtiment**
   - **Bâtiment lourd (béton, pierre) :**
     - Forte inertie : dépassement normal
     - Augmenter anticipation régulation
     - Programmer arrêt plus tôt
   - **Bâtiment léger (ossature bois) :**
     - Faible inertie : réaction rapide
     - Dépassement anormal si se produit

10. **Fonction optimisation**
    - Certains thermostats : apprentissage automatique
    - Analysent comportement thermique
    - Ajustent automatiquement anticipation
    - Activer si disponible (ex: Nest, Netatmo)

**Prévention :**
- Dimensionnement correct chaudière dès installation
- Régulation adaptée (PID préférable à TOR)
- Vannes thermostatiques radiateurs
- Ajustement paramètres après travaux isolation (inertie modifiée)

**Dépassement acceptable selon système :**
| Système | Dépassement acceptable |
|---------|------------------------|
| TOR simple | 1-2°C |
| PID bien réglé | 0,5-1°C |
| Régulation climatique | < 0,5°C |
| Plancher chauffant | 0,5°C (forte inertie mais régulation lente) |
| Radiateurs fonte | 1-2°C (inertie émetteurs) |
| Radiateurs alu | 0,5-1°C (faible inertie) |

**Impact surchauffe :**
- Inconfort occupants (transpiration, ouverture fenêtres)
- Surconsommation énergétique (gaspillage)
- Usure installation (cycles thermiques)
- Air intérieur sec (inconfort, santé)

**Solutions court terme :**
- Réduire consigne -1°C (compenser dépassement)
- Programmer arrêt anticipé
- Régler têtes thermostatiques plus bas

**Solutions long terme :**
- Régulation performante (PID, OpenTherm)
- Chaudière modulante correctement dimensionnée
- Ballon tampon si nécessaire
- Régulation pièce par pièce

---

