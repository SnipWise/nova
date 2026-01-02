## FACT-CHAUD-072: Thermostat d'ambiance sans effet

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-072 |
| **Catégorie** | Régulation & Sondes |
| **Système** | Thermostat d'ambiance |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Modification consigne thermostat sans effet sur chaudière
- Chaudière fonctionne en continu malgré température atteinte
- Ou chaudière ne démarre pas malgré demande thermostat
- Température intérieure ne correspond pas à consigne
- Thermostat affiche correct mais régulation inopérante

**Cause racine probable :**
Panne thermostat (piles, capteur), problème câblage, mauvais paramétrage chaudière (influence thermostat), thermostat mal positionné, conflit régulation climatique/thermostat.

**Étapes de résolution :**

1. **Vérification type thermostat**
   - **Thermostat filaire TOR (Tout Ou Rien) :**
     - Contact sec marche/arrêt
     - 2 fils vers chaudière
   - **Thermostat modulant (OpenTherm, eBUS) :**
     - Communication bidirectionnelle
     - 2 fils bus de communication
   - **Thermostat radio/WiFi :**
     - Récepteur sur/dans chaudière
     - Communication sans fil

2. **Test thermostat TOR filaire**
   - Retirer thermostat du socle mural
   - Court-circuiter les 2 fils au niveau chaudière
   - Si chaudière démarre : thermostat défectueux
   - Si chaudière ne démarre pas : problème chaudière ou câblage
   - Mesurer continuité contact thermostat (multimètre)
   - Contact fermé = demande chauffe (0 Ω)
   - Contact ouvert = pas de demande (infini)

3. **Contrôle câblage**
   - Vérifier continuité câble sur toute longueur
   - Contrôler serrage bornes (thermostat et chaudière)
   - Vérifier polarité si thermostat modulant
   - Contrôler absence court-circuit
   - Tester avec câble provisoire si doute

4. **Vérification piles thermostat**
   - Remplacer piles (même si affichage OK)
   - Piles faibles : affichage OK mais relais ne commute pas
   - Utiliser piles alcalines qualité (durée 1-2 ans)
   - Certains thermostats signalent piles faibles

5. **Contrôle positionnement thermostat**
   - **Position correcte :**
     - Pièce de vie principale (séjour)
     - Mur intérieur (pas façade)
     - Hauteur 1,50 m
     - Loin sources chaleur (radiateur, soleil, TV, lampe)
     - Loin courants d'air (porte, fenêtre)
     - Circulation air normale (pas derrière rideau/meuble)
   - **Mauvais positionnement :**
     - → Mesure température non représentative
     - → Régulation inadaptée

6. **Paramétrage chaudière**
   - Vérifier activation entrée thermostat (menu installateur)
   - Contrôler type thermostat déclaré (TOR/modulant)
   - Vérifier influence thermostat sur régulation
   - Paramètres selon marque :
     - Vaillant : paramètre 700 (fonction thermostat)
     - Saunier Duval : activation contact TA
     - Frisquet : déclaration Eco-Radio ou contact sec

7. **Conflit régulation climatique**
   - Si régulation climatique + thermostat :
     - Régler pondération (priorité climatique ou thermostat)
     - Mode cascade : climatique principal, thermostat limiteur
     - Éviter consignes contradictoires
   - Vérifier température départ (ne doit pas être trop basse)
   - Autoriser chaudière à chauffer selon météo

8. **Test thermostat modulant**
   - Vérifier communication (LED sur récepteur)
   - Contrôler messages erreur thermostat
   - Tester association thermostat/récepteur
   - Réinitialiser et ré-appairer si nécessaire
   - Vérifier compatibilité (OpenTherm, eBUS selon marque)

9. **Diagnostic avancé**
   - Consulter menu diagnostic chaudière
   - Vérifier état entrée thermostat (ouvert/fermé)
   - Observer changement état lors modification consigne
   - Si état change mais pas d'effet : paramétrage chaudière
   - Si état ne change pas : thermostat ou câblage

**Prévention :**
- Remplacement piles annuel (préventif)
- Vérification câblage lors entretien
- Test fonctionnement début saison chauffe
- Dépoussiérage thermostat
- Vérification positionnement (pas de nouveau meuble, rideau)

**Solutions selon diagnostic :**
- Thermostat HS : remplacement
- Câblage coupé : réparation ou remplacement câble
- Mauvais paramétrage : reconfiguration
- Mauvais positionnement : déplacement

**Choix thermostat :**
- **Simple TOR :** économique, fiable, basique
- **Programmable :** confort + économies (plages horaires)
- **Modulant OpenTherm :** optimisation rendement chaudière
- **Connecté WiFi :** pilotage distance, statistiques
- **Multi-zones :** confort pièce par pièce

**Économies thermostat programmable :**
- Réduction nuit : 15-20% économies
- Réduction absences : 10-15% économies
- Total potentiel : 20-30% vs thermostat fixe

---

