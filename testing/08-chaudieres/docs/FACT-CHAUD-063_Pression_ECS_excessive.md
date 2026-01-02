## FACT-CHAUD-063: Pression ECS excessive

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-063 |
| **Catégorie** | Eau Chaude Sanitaire (ECS) |
| **Système** | Pression sanitaire |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Pression ECS trop élevée (> 5 bars)
- Coups de bélier au robinet
- Soupape sécurité sanitaire qui fuit
- Bruit fort à la fermeture robinet
- Groupe sécurité goutte en continu

**Cause racine probable :**
Réducteur pression défectueux (trop haute pression réseau), absence réducteur pression, dilatation eau chaude (pas de vase expansion sanitaire), clapet anti-retour bloqué, groupe sécurité sous-dimensionné.

**Étapes de résolution :**

1. **Mesure pression ECS**
   - Installer manomètre sur robinet purge
   - Mesurer pression eau froide statique
   - Mesurer pression eau chaude (robinets fermés)
   - **Pressions normales :**
     - Réseau eau froide : 2-4 bars (ville)
     - Maximum recommandé : 5 bars
     - Maximum admissible : 6-7 bars (risque)

2. **Contrôle réducteur pression**
   - Localiser réducteur pression (entrée installation)
   - Mesurer pression aval réducteur
   - Réglage réducteur :
     - Pression cible : 3 bars
     - Ajuster vis réglage (rotation)
     - Tester après ajustement
   - Si réglage impossible : réducteur HS (membrane, ressort)
   - Remplacer réducteur si défectueux

3. **Contrôle dilatation eau chaude**
   - **Problème :**
     - Eau chaude se dilate (+3% de 20 à 60°C)
     - Si circuit fermé (clapet anti-retour) : surpression
     - Pression peut monter à 10-15 bars (danger)
   - **Solution 1 : Vase expansion sanitaire**
     - Vase 2-4 litres sur circuit ECS
     - Absorbe dilatation
     - Prégonflage : 80% pression eau froide
   - **Solution 2 : Groupe sécurité**
     - Groupe sécurité ECS (soupape + disconnecteur)
     - Évacue surpression par goutte-à-goutte
     - Raccordement évacuation obligatoire

4. **Contrôle groupe sécurité**
   - Groupe sécurité = soupape taré 7 bars
   - Fonctionnement normal : goutte occasionnelle (dilatation)
   - Fonctionnement anormal : fuite continue
   - **Causes fuite continue :**
     - Soupape entartrée (reste ouverte)
     - Pression excessive (> 7 bars)
     - Soupape HS (ressort fatigué)
   - **Test groupe sécurité :**
     - Actionner manette vidange
     - Eau doit s'écouler franchement
     - Relâcher : écoulement doit s'arrêter
     - Si fuite persiste : détartrer ou remplacer

5. **Détartrage groupe sécurité**
   - Démonter groupe sécurité
   - Tremper dans vinaigre blanc 12h
   - Actionner mécanisme plusieurs fois
   - Rincer abondamment
   - Si fuite persiste après nettoyage : remplacement

6. **Installation vase expansion sanitaire**
   - **Si absent :**
     - Vase 2-4L selon volume ECS
     - Montage sur départ eau chaude
     - Prégonflage côté air : 2,5 bars (80% pression eau froide)
   - **Avantages :**
     - Absorbe dilatation
     - Réduit cycles groupe sécurité
     - Protège installation
     - Réduit consommation eau (moins pertes)

7. **Contrôle clapet anti-retour**
   - Clapet anti-retour : empêche retour ECS vers eau froide
   - Si bloqué fermé : circuit fermé → surpression
   - Démonter, nettoyer ou remplacer

8. **Cas pression réseau excessive**
   - Si pression réseau > 5 bars :
     - Installation réducteur pression OBLIGATOIRE
     - Protection installation (robinets, flexibles)
     - Confort (éviter coups de bélier)
   - Réducteur à membrane : plus fiable
   - Réglage : 3 bars recommandé

**Prévention :**
- Installation réducteur pression (pression réseau > 4 bars)
- Vase expansion sanitaire (circuit ECS fermé)
- Groupe sécurité vérifié annuellement
- Détartrage groupe sécurité si eau dure
- Contrôle pression annuel

**Réglementation :**
- Groupe sécurité obligatoire (chauffe-eau, ballon)
- Pression max admissible : 7 bars (appareils domestiques)
- Évacuation groupe sécurité : visible, raccordée évacuation

**Dimensionnement vase expansion sanitaire :**
- Formule : V = (Ve × Ce × ΔT) / (1 - (P1/P2))
  - Ve : volume eau ECS (litres)
  - Ce : coefficient expansion eau (0,04 pour ΔT 50°C)
  - ΔT : variation température
  - P1 : pression gonflage vase (bars absolus)
  - P2 : pression tarage soupape (bars absolus)
- Exemple : ballon 200L → vase 3-4L

---

