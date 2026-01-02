## FACT-CHAUD-070: Régulation climatique mal paramétrée

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-070 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Régulation climatique |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Température intérieure instable ou inconfortable
- Surchauffe locaux (trop chaud malgré thermostat)
- Sous-chauffe locaux (pas assez chaud)
- Température départ inadaptée aux conditions
- Surconsommation énergétique

**Cause racine probable :**
Courbe de chauffe inadaptée (pente incorrecte), décalage parallèle mal réglé, sonde extérieure mal positionnée, absence optimisation lors mise en service, modification isolation non prise en compte.

**Étapes de résolution :**

1. **Compréhension régulation climatique**
   - Principe : température départ varie selon température extérieure
   - Plus il fait froid dehors, plus eau départ est chaude
   - Loi d'eau = courbe de chauffe (graphique T°ext vs T°départ)
   - Objectif : température intérieure stable quelle que soit météo

2. **Diagnostic situation actuelle**
   - Relever paramètres courbe chauffe actuelle (menu installateur)
   - Noter pente et décalage parallèle
   - Mesurer température départ en fonction T°ext
   - Interroger occupants sur confort (trop chaud/froid, quand)
   - Relever température intérieure réelle (plusieurs pièces)

3. **Vérification sonde extérieure**
   - S'assurer température extérieure affichée cohérente
   - Vérifier positionnement correct sonde (voir FACT-CHAUD-068)
   - Pas de soleil direct (fausserait mesure)
   - Comparer avec météo locale

4. **Analyse courbe de chauffe**
   - **Pente (inclinaison courbe) :**
     - Pente faible (0,3-0,8) : plancher chauffant, forte isolation
     - Pente moyenne (1,0-1,5) : radiateurs BT, isolation moyenne
     - Pente forte (1,5-2,5) : radiateurs HT, faible isolation
     - Pente excessive : radiateurs anciens, pas d'isolation
   - **Décalage parallèle :**
     - Décale toute la courbe vers haut ou bas
     - Ajustement fin du confort
     - Plage typique : -10 à +10°C

5. **Ajustement courbe - Méthode progressive**
   - **Si trop froid en permanence :**
     - Augmenter décalage parallèle (+2 à +5°C)
     - Attendre 24-48h, évaluer
     - Si insuffisant : augmenter pente (0,1 à 0,2)
   - **Si trop chaud en permanence :**
     - Réduire décalage parallèle (-2 à -5°C)
     - Attendre 24-48h, évaluer
     - Si insuffisant : réduire pente (0,1 à 0,2)
   - **Si trop froid seulement quand très froid dehors :**
     - Augmenter pente uniquement
   - **Si trop chaud seulement en mi-saison :**
     - Réduire décalage parallèle

6. **Paramétrage type émetteurs**
   - **Plancher chauffant :**
     - Pente : 0,3 à 0,6
     - T°départ max : 35-45°C
     - Régulation lente (inertie importante)
   - **Radiateurs basse température :**
     - Pente : 0,8 à 1,2
     - T°départ max : 45-55°C
   - **Radiateurs moyenne température :**
     - Pente : 1,2 à 1,8
     - T°départ max : 55-65°C
   - **Radiateurs haute température :**
     - Pente : 1,8 à 2,5
     - T°départ max : 70-80°C

7. **Optimisations complémentaires**
   - Activer optimisation démarrage (anticipation)
   - Paramétrer température réduit nuit (économies)
   - Régler temporisations (éviter cycles courts)
   - Configurer influence thermostat ambiance (pondération)

8. **Test et affinage**
   - Laisser fonctionner 48-72h entre chaque modification
   - Interroger occupants régulièrement
   - Mesurer températures intérieures (plusieurs pièces)
   - Noter consommation avant/après
   - Affiner par petites touches (patience nécessaire)

**Prévention :**
- Révision paramètres après travaux isolation
- Ajustement saisonnier (automne/hiver)
- Réévaluation annuelle confort occupants
- Formation utilisateurs (ne pas modifier n'importe comment)

**Erreurs fréquentes :**
- Pente trop forte : surchauffe mi-saison, cycles courts
- Pente trop faible : sous-chauffe grand froid
- Modifications trop fréquentes (ne pas laisser temps stabilisation)
- Ignorer l'inertie du bâtiment
- Ne pas tenir compte des apports gratuits (soleil, occupation)

**Gains régulation climatique bien réglée :**
- Économies : 10-25% selon installation
- Confort : température stable ±0,5°C
- Longévité chaudière : moins de cycles
- Réduction empreinte carbone

**Exemple courbe type (radiateurs MT, T°int consigne 20°C) :**
| T°ext | T°départ (pente 1,4) |
|-------|----------------------|
| -10°C | 70°C |
| -5°C  | 63°C |
| 0°C   | 56°C |
| 5°C   | 49°C |
| 10°C  | 42°C |
| 15°C  | 35°C |
| 20°C  | Arrêt chauffage |

---

