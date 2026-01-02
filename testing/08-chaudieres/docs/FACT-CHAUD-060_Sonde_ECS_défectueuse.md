## FACT-CHAUD-060: Sonde ECS défectueuse

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-060 |
| **Catégorie** | Eau Chaude Sanitaire (ECS) |
| **Système** | Régulation température ECS |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Température ECS incohérente (affichage vs réalité)
- Code erreur sonde ECS
- Température ECS instable ou excessive
- Chaudière ne régule pas température ECS
- Affichage température aberrante (ex: 120°C)

**Cause racine probable :**
Sonde CTN/NTC défectueuse (dérive, coupure), mauvais contact thermique sonde, câblage coupé ou court-circuit, connecteur oxydé, sonde mal positionnée (hors flux).

**Étapes de résolution :**

1. **Identification symptômes**
   - Code erreur sonde affiché (ex: Vaillant F75)
   - Température affichée aberrante
   - Température affichée fixe (ne varie pas)
   - ECS trop chaude ou trop froide

2. **Localisation sonde ECS**
   - Sonde généralement sur sortie échangeur sanitaire
   - Montage :
     - Doigt de gant (immergé dans flux)
     - Collier sur tuyauterie (contact externe)
     - Intégré échangeur (selon modèle)
   - Vérifier repérage (sonde ECS vs sonde chauffage)

3. **Test électrique sonde**
   - Déconnecter sonde (connecteur chaudière)
   - Mesurer résistance sonde (multimètre Ω)
   - **Sonde CTN/NTC typique :**
     - 10 kΩ à 25°C
     - 3,3 kΩ à 50°C
     - 1,5 kΩ à 70°C
   - Comparer à courbe constructeur (notice technique)
   - Valeurs aberrantes :
     - Infini (∞) : sonde coupée
     - 0 Ω : court-circuit
     - Valeur fixe qui ne varie pas : sonde HS

4. **Test variation résistance**
   - Chauffer sonde (eau chaude, sèche-cheveux)
   - Résistance doit diminuer avec température (CTN)
   - Variation normale : divisée par 2-3 tous les 20°C
   - Si pas de variation : sonde défectueuse

5. **Contrôle contact thermique**
   - Sonde doigt de gant : vérifier présence pâte thermique
   - Doigt de gant vide : mauvais contact → mesure erronée
   - Appliquer pâte thermique (conductivité thermique)
   - Sonde collier : vérifier serrage correct (contact métal/métal)

6. **Contrôle câblage**
   - Vérifier continuité câble sonde → carte
   - Contrôler isolation (résistance câble/masse > 1 MΩ)
   - Inspecter câble (coupure, dénudage, brûlure)
   - Vérifier connecteur (oxydation, humidité)
   - Nettoyer contacts (bombe contact électronique)

7. **Test avec sonde de rechange**
   - Si doute : tester avec sonde neuve
   - Ou permuter temporairement sonde chauffage/ECS (test)
   - Attention : sondes parfois différentes (courbes)

8. **Remplacement sonde**
   - Commander sonde référence constructeur
   - Ou sonde universelle (vérifier courbe compatible)
   - Vidanger circuit si nécessaire (doigt de gant)
   - Installer sonde neuve :
     - Doigt de gant : pâte thermique obligatoire
     - Collier : serrage correct, contact tuyau
   - Reconnecter câblage (respecter polarité si marquée)

9. **Vérification après remplacement**
   - Vérifier température affichée cohérente
   - Lancer production ECS
   - Observer régulation température
   - Vérifier température robinet (thermomètre)
   - Écart affiché/réel doit être < 5°C

**Prévention :**
- Vérification annuelle valeur sonde (menu diagnostic)
- Contrôle contact thermique
- Protection câblage (pas de frottement, chaleur excessive)
- Nettoyage connecteurs
- Remplacement préventif (durée vie ~15 ans)

**Codes erreur sonde ECS par marque :**
- Vaillant : F75 (défaut sonde ECS)
- Saunier Duval : F73 (sonde ECS court-circuit), F74 (sonde ECS coupée)
- Elm Leblanc : E03 (sonde ECS)
- De Dietrich : E30 (sonde sanitaire)
- Frisquet : 203 (sonde sanitaire)

**Courbe typique CTN 10kΩ (exemple) :**
- 0°C : 32,7 kΩ
- 25°C : 10 kΩ
- 50°C : 3,3 kΩ
- 80°C : 1,1 kΩ

---

