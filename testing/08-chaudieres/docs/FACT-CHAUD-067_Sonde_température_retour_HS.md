## FACT-CHAUD-067: Sonde température retour HS

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-067 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Sonde température retour |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques (chaudières avec régulation avancée) |

**Symptômes :**
- Code erreur sonde retour (moins fréquent que sonde départ)
- Calcul ΔT impossible ou erroné
- Protection anti-condensation non fonctionnelle (chaudières standard)
- Modulation sous-optimale
- Régulation hydraulique dégradée

**Cause racine probable :**
Sonde CTN défectueuse, mauvais contact thermique, câblage défectueux, connecteur oxydé, absence pâte thermique.

**Étapes de résolution :**

1. **Vérification présence sonde**
   - Toutes les chaudières n'ont pas de sonde retour
   - Vérifier schéma technique de la chaudière
   - Localiser la sonde (généralement sur tube retour primaire)
   - Identifier le connecteur sur carte électronique

2. **Diagnostic différentiel**
   - Comparer température retour affichée avec départ
   - ΔT normal chauffage : 15-25°C (retour < départ)
   - Si retour > départ : incohérence, sonde défectueuse
   - Si retour = départ : sonde mal placée ou défectueuse

3. **Mesure résistance sonde**
   - Couper alimentation électrique
   - Déconnecter sonde au niveau carte
   - Mesurer résistance (mêmes valeurs que sonde départ)
   - Comparer avec courbe constructeur
   - Mesurer température réelle eau retour (thermomètre contact)

4. **Contrôle positionnement sonde**
   - Vérifier que sonde est sur retour (pas départ)
   - Contrôler immersion dans doigt de gant
   - Vérifier contact thermique (pâte thermique)
   - S'assurer que doigt de gant est dans flux eau
   - Vérifier absence air dans doigt de gant

5. **Test croisé avec sonde départ**
   - Échanger temporairement sondes départ/retour (si identiques)
   - Si défaut suit la sonde : sonde HS
   - Si défaut reste sur même entrée carte : défaut carte
   - Remettre en position d'origine

6. **Remplacement et installation**
   - Remplacer par sonde identique (référence constructeur)
   - Appliquer pâte thermique généreusement
   - Enfoncer complètement dans doigt de gant
   - Fixer solidement
   - Protéger connexion de l'humidité

7. **Vérification fonctionnelle**
   - Contrôler températures affichées départ et retour
   - Vérifier calcul ΔT cohérent (menu diagnostic)
   - Tester régulation (adaptation puissance selon ΔT)
   - Vérifier anti-condensation si applicable
   - Observer modulation sur cycle complet

**Prévention :**
- Vérification annuelle cohérence départ/retour
- Contrôle ΔT lors entretien (valeur typique 15-20°C)
- Inspection connectique
- Renouvellement pâte thermique

**Impact panne sonde retour :**
- Perte régulation optimale (modulation moins précise)
- Protection anti-condensation inopérante (chaudières standard)
- Calcul rendement impossible
- Diagnostic hydraulique difficile
- Consommation potentiellement accrue

**Fonction sonde retour selon type chaudière :**
- **Chaudière condensation :** optimisation condensation, modulation
- **Chaudière standard :** protection anti-condensation (T retour > 50°C)
- **Chaudière modulante :** calcul ΔT pour ajuster puissance
- **Système hydraulique complexe :** gestion circuits multiples

---

