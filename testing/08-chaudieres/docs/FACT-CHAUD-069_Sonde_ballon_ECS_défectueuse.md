## FACT-CHAUD-069: Sonde ballon ECS défectueuse

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-069 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Production ECS accumulée |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières avec ballon ECS) |

**Symptômes :**
- Température ECS affichée incohérente
- Eau tiède ou froide alors que chaudière chauffe
- Eau brûlante (surchauffe ballon)
- Cycles de chauffe incessants ou absents
- Code erreur sonde ballon (selon marque)

**Cause racine probable :**
Sonde CTN défectueuse, mauvais contact thermique avec ballon, sonde mal positionnée (hors eau), connectique oxydée, câble endommagé.

**Étapes de résolution :**

1. **Diagnostic symptômes**
   - Relever température ECS affichée
   - Mesurer température réelle eau (thermomètre robinet)
   - Noter fréquence cycles chauffe ballon
   - Vérifier consigne température ECS (menu)
   - Observer déclenchements chauffage ballon

2. **Localisation sonde ballon**
   - Sonde généralement dans doigt de gant sur ballon
   - Position idéale : tiers supérieur ballon
   - Parfois sonde sur départ ECS ou bouclage
   - Identifier câble et connecteur carte chaudière

3. **Mesure résistance sonde**
   - Couper alimentation électrique
   - Déconnecter sonde au niveau carte
   - Mesurer résistance avec multimètre
   - Comparer avec température réelle eau ballon
   - Valeurs attendues CTN : 2 kΩ à 60°C, 1,5 kΩ à 70°C
   - Incohérence : sonde défectueuse

4. **Contrôle positionnement sonde**
   - Vérifier immersion complète dans doigt de gant
   - S'assurer que doigt de gant baigne dans l'eau
   - Vérifier pâte thermique (contact thermique)
   - Contrôler absence air dans doigt de gant
   - Vérifier que doigt de gant n'est pas obstrué (tartre)

5. **Test comportement**
   - Lancer cycle chauffe ECS manuellement
   - Observer évolution température affichée
   - Température doit monter progressivement
   - Arrêt chauffe doit se produire à consigne
   - Si pas d'arrêt : sonde ne détecte pas montée température

6. **Contrôle installation**
   - Sur ballon indirect : vérifier circulation primaire (vanne 3 voies)
   - Contrôler pompe ballon si présente
   - Vérifier absence entartrage échangeur ballon
   - S'assurer stratification correcte ballon

7. **Remplacement sonde**
   - Vidanger partiellement ballon si nécessaire
   - Déposer ancienne sonde (dévisser doigt de gant)
   - Nettoyer doigt de gant (détartrage si besoin)
   - Appliquer pâte thermique
   - Installer sonde neuve, enfoncer complètement
   - Resserrer doigt de gant (joint neuf)
   - Reconnecter électriquement

8. **Paramétrage et test**
   - Régler consigne ECS (généralement 55-60°C)
   - Lancer cycle chauffe complet
   - Vérifier arrêt à consigne
   - Tester puisage : température cohérente
   - Vérifier absence relance intempestive
   - Ajuster consigne selon confort et légionelle

**Prévention :**
- Vérification annuelle température ballon (affichée vs réelle)
- Contrôle cycles chauffe ECS
- Détartrage ballon selon dureté eau (tous les 3-5 ans)
- Contrôle connectique sonde

**Risques sonde ballon défectueuse :**
- **Sonde indiquant trop chaud :** eau tiède, inconfort, risque légionelle
- **Sonde indiquant trop froid :** surchauffe, ébouillantage, surconsommation, entartrage accéléré
- **Sonde HS :** pas de production ECS ou chauffe permanente

**Réglage température ECS :**
- **Minimum :** 55°C (prévention légionelle)
- **Recommandé :** 55-60°C (confort + sécurité sanitaire)
- **Maximum :** 65°C (limite entartrage, sécurité)
- **Avec mitigeur thermostatique :** possible 60-65°C (distribution 50°C)

**Cycle anti-légionelle :**
- Certaines chaudières proposent montée 65-70°C hebdomadaire
- Désinfection thermique ballon
- Vérifier fonction active et opérationnelle

---

