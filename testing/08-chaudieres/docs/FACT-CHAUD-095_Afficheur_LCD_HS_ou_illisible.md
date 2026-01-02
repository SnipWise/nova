## FACT-CHAUD-095: Afficheur LCD HS ou illisible

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-095 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Interface utilisateur |
| **Gravité** | Faible |
| **Marques** | Multi-marques |

**Symptômes :**
- Écran complètement noir ou blanc
- Affichage partiel (segments manquants)
- Affichage fantôme (tous segments allumés)
- Contraste trop faible ou trop fort
- Affichage inversé (négatif)
- LCD cassé, fissuré

**Cause racine probable :**
Nappe LCD dessoudée/déconnectée, LCD cassé (choc), vieillissement cristaux liquides, température extrême, défaut contraste, défaut rétroéclairage, carte électronique défaillante.

**Étapes de résolution :**

1. **Diagnostic type affichage**
   - **Écran complètement noir** :
     - Vérifier rétroéclairage (visible dans le noir)
     - Si rétroéclairage OK mais pas de caractères : problème contraste ou LCD
   - **Écran blanc** :
     - Problème contraste (réglage ou composant)
     - Nappe déconnectée
   - **Segments manquants** :
     - Nappe partiellement déconnectée
     - Pistes coupées sur nappe
     - LCD défectueux
   - **Tous segments allumés** :
     - Court-circuit commande LCD
     - Carte électronique défaillante

2. **Vérification nappe LCD**
   - Couper alimentation
   - Localiser connexion LCD ↔ carte électronique
   - Types connexion :
     - Nappe élastomère (zébra) : bande caoutchouc conductrice
     - Nappe flexible soudée
     - Connecteur enfichable
   - Vérifier état nappe : pas de coupure, pliure excessive
   - Nettoyer contacts (alcool isopropylique + coton-tige)
   - Remettre en place en appuyant fermement (zébra)

3. **Réglage contraste LCD**
   - Certains LCD ont potentiomètre contraste sur carte
   - Localiser (petit composant bleu à vis cruciforme)
   - Ajuster délicatement (1/4 tour max à la fois)
   - Tester affichage après chaque ajustement
   - Position optimale : affichage net, noir/blanc contrasté

4. **Test rétroéclairage**
   - Regarder LCD dans le noir
   - Si éclairage visible mais pas de caractères : problème LCD ou nappe
   - Si pas d'éclairage :
     - Vérifier LED rétroéclairage (arrière LCD)
     - Tester alimentation LED (généralement 12V ou 3.3V)
     - Remplacer LED si grillée (soudure CMS)

5. **Contrôle température**
   - LCD sensibles températures extrêmes :
     - < 0°C : cristaux liquides figés, affichage lent/absent
     - > 60°C : dégradation cristaux, affichage fantôme
   - Vérifier température locale chaufferie
   - Isoler/ventiler si nécessaire
   - LCD peut récupérer après retour température normale

6. **Test LCD avec autre carte**
   - Si possible : tester LCD sur carte identique fonctionnelle
   - Permet identifier : problème LCD ou problème carte
   - Attention compatibilité : certains LCD spécifiques

7. **Remplacement LCD**
   - **LCD zébra** (bande élastomère) :
     - Commander LCD + bande élastomère
     - Déposer ancien LCD (clips ou vis)
     - Positionner bande zébra (alignement précis)
     - Clip nouveau LCD
   - **LCD nappe soudée** :
     - Dessouder nappe (fer fine pointe, température contrôlée)
     - Nettoyer pads
     - Souder nouveau LCD (flux, soudure fine)
     - Attention : facile d'endommager pistes
   - **LCD enfichable** :
     - Retirer connecteur
     - Enficher nouveau LCD

8. **Solution temporaire**
   - Si affichage défectueux mais chaudière fonctionne :
     - Utiliser codes LED (selon marque) pour diagnostic
     - Installer thermostat d'ambiance programmable (compense)
     - Planifier remplacement LCD ou carte

9. **Alternative : remplacement carte complète**
   - Si LCD intégré carte (soudé CMS complexe) :
     - Remplacement LCD seul difficile
     - Préférer remplacement carte complète
   - Comparer coûts : LCD seul vs carte complète

**Prévention :**
- Protection contre températures extrêmes
- Éviter chocs sur afficheur
- Nettoyage doux (chiffon microfibre légèrement humide)
- Ne jamais appuyer fortement sur LCD

**Compatibilité :**
- LCD souvent spécifiques à chaque modèle chaudière
- Noter références exactes avant commande
- Vérifier nombre de caractères, segments
- Photos utiles pour identification

**Fonctionnement sans afficheur :**
- Chaudière peut fonctionner sans affichage
- Utiliser codes LED ou thermostat externe
- Diagnostic limité mais fonctionnel

---

