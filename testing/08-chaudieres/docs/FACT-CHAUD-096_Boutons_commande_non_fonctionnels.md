## FACT-CHAUD-096: Boutons commande non fonctionnels

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-096 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Interface boutons poussoirs |
| **Gravité** | Faible à Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Boutons ne répondent pas à la pression
- Bouton enfoncé en permanence (collé)
- Réponse intermittente (il faut appuyer plusieurs fois)
- Mauvais bouton activé (crosstalk)
- Impression tactile altérée (pas de "clic")

**Cause racine probable :**
Bouton mécanique usé, dôme conducteur cassé, oxydation contacts, encrassement, humidité, nappe boutons déconnectée, piste carte coupée.

**Étapes de résolution :**

1. **Identification type boutons**
   - **Boutons poussoirs mécaniques** (switches) :
     - Composant monté sur carte
     - Tactile "clic" mécanique
   - **Boutons à dôme** (dome switch) :
     - Dôme métallique sous membrane
     - Contact sur pistes PCB
   - **Boutons capacitifs** (tactile) :
     - Pas de mouvement mécanique
     - Détection capacitive doigt

2. **Test boutons individuels**
   - Accéder mode diagnostic/menu si possible
   - Tester chaque bouton séquentiellement
   - Noter quels boutons répondent/ne répondent pas
   - Vérifier absence de bouton bloqué enfoncé

3. **Nettoyage boutons à dôme**
   - Démonter façade chaudière
   - Retirer membrane silicone/plastique
   - Nettoyer dômes métalliques :
     - Alcool isopropylique + coton-tige
     - Enlever oxydation, saleté
   - Nettoyer pistes PCB sous dômes :
     - Alcool isopropylique
     - Gomme douce si oxydation importante
   - Nettoyer membrane par l'intérieur
   - Sécher complètement avant remontage

4. **Contrôle dômes métalliques**
   - Vérifier état dômes :
     - Bombement correct (pas écrasés)
     - Pas de déchirure, déformation
     - Élasticité conservée (ressort)
   - Remplacer dômes défectueux :
     - Disponibles commerce électronique
     - Tailles standards : 8mm, 12mm diamètre
     - Hauteur et force variable

5. **Test boutons poussoirs mécaniques**
   - Couper alimentation
   - Mesurer continuité bouton au repos : ouvert (∞)
   - Appuyer sur bouton : fermé (0 Ω)
   - Si pas de commutation : bouton HS
   - Dessouder et remplacer bouton

6. **Contrôle nappe boutons**
   - Si boutons déportés (nappe vers carte) :
     - Vérifier connexion nappe ↔ carte
     - Nettoyer connecteur
     - Tester continuité pistes nappe
     - Remplacer nappe si coupée

7. **Vérification pistes PCB**
   - Inspecter pistes sous boutons (loupe)
   - Vérifier absence coupure piste
   - Tester continuité bouton → composant carte
   - Réparer piste si coupée (fil émaillé fin + soudure)

8. **Remplacement boutons poussoirs**
   - Noter référence bouton (hauteur, force, type)
   - Types courants :
     - Tactile switch 6x6mm hauteur 4.3mm à 13mm
     - Force : 160gf, 260gf, 520gf
   - Dessouder ancien bouton (pompe à dessouder)
   - Souder nouveau bouton (alignement correct)

9. **Solution temporaire**
   - Si bouton unique défectueux :
     - Utiliser thermostat d'ambiance pour contrôle
     - Ou télécommande filaire (si compatible)
   - Planifier réparation

10. **Boutons capacitifs défectueux**
    - Plus complexe à diagnostiquer
    - Vérifier circuits capacitifs (condensateurs, IC)
    - Nettoyer surface tactile (traces de doigt)
    - Recalibration possible sur certains modèles
    - Sinon : remplacement carte ou module tactile

**Prévention :**
- Éviter humidité (mains mouillées)
- Nettoyage régulier façade
- Ne pas appuyer avec objets durs (risque casse)
- Protection contre vapeur cuisine (si chaudière en cuisine)

**Pièces de rechange :**
- Dômes conducteurs : lot 100 pièces ~5-10€
- Boutons tactiles 6x6mm : lot 50 pièces ~3-5€
- Disponibles sur sites électronique (AliExpress, Amazon, Mouser)

**Fonctionnement sans boutons :**
- Installer thermostat d'ambiance filaire
- Utilise uniquement consigne thermostat
- Chaudière suit demande externe
- Perd accès réglages avancés

---

