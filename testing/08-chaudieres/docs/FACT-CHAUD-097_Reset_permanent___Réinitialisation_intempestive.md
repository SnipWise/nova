## FACT-CHAUD-097: Reset permanent / Réinitialisation intempestive

| Champ | Valeur |
|-------|--------|
| **ID** | FACT-CHAUD-097 |
| **Catégorie** | Électronique & Cartes de Commande |
| **Système** | Alimentation / Mémoire carte |
| **Gravité** | Moyenne |
| **Marques** | Multi-marques |

**Symptômes :**
- Chaudière redémarre aléatoirement
- Horloge se réinitialise fréquemment (00:00)
- Paramètres perdus après chaque coupure
- Écran affiche "Initialisation" de façon répétée
- Erreurs mémoire affichées
- Code erreur watchdog / reset

**Cause racine probable :**
Micro-coupures secteur, tension instable, condensateur mémoire défaillant, pile CMOS/RTC déchargée, défaut alimentation carte, watchdog intempestif, défaut firmware.

**Étapes de résolution :**

1. **Analyse fréquence réinitialisations**
   - Noter quand elles surviennent :
     - À chaque coupure secteur → problème mémoire/pile
     - Aléatoire en fonctionnement → problème alimentation ou watchdog
     - À chaque démarrage pompe/ventilateur → chute tension
   - Observer codes erreur associés

2. **Contrôle alimentation secteur**
   - Mesurer tension secteur : 230V stable
   - Installer enregistreur tension (si disponible)
   - Vérifier absence micro-coupures
   - Contrôler qualité neutre (neutre-terre < 5V)
   - Tester avec autre circuit électrique (isolation)

3. **Vérification pile RTC (Real-Time Clock)**
   - Localiser pile sur carte :
     - Type CR2032 (lithium 3V) souvent
     - Ou pile rechargeable ML2032
     - Parfois super-condensateur
   - Mesurer tension pile : doit être > 2.7V
   - Si < 2.5V : pile déchargée, remplacement nécessaire
   - Pile durée vie : 3-5 ans

4. **Remplacement pile RTC**
   - Couper alimentation chaudière
   - Retirer pile (clip ou porte-pile)
   - Remplacer par pile identique (CR2032 3V)
   - **Attention** : certaines piles sont soudées (ML2032)
     - Dessouder ancienne
     - Souder nouvelle (rapidement, chaleur limitée)
   - Remettre alimentation
   - Régler horloge et paramètres

5. **Contrôle condensateurs mémoire**
   - Localiser condensateurs alimentation (gros cylindres)
   - Inspecter visuellement :
     - Pas de bombement dessus
     - Pas de fuite électrolyte (liquide)
     - Pas de traces brunâtres
   - Mesurer capacité si testeur disponible
   - Remplacer si défectueux (ESR élevé ou capacité faible)

6. **Diagnostic watchdog**
   - Watchdog = circuit surveillance qui reset si blocage détecté
   - Reset watchdog intempestif → problème firmware ou carte
   - Codes erreur possibles : "Watchdog", "WD Reset", "Internal Error"
   - Solutions :
     - Mise à jour firmware
     - Réinitialisation complète paramètres usine
     - Remplacement carte si persiste

7. **Contrôle chutes tension au démarrage charges**
   - Symptôme : reset quand pompe ou ventilateur démarre
   - Cause : chute tension excessive (démarrage moteur)
   - Mesurer tension 230V au démarrage pompe (oscilloscope ou voltmètre rapide)
   - Solutions :
     - Vérifier câblage alimentation (section suffisante)
     - Installer condensateur stabilisation
     - Vérifier qualité transformateur alimentation carte

8. **Réinitialisation complète paramètres**
   - Accéder menu service/installateur
   - Réinitialisation factory reset
   - Reprogrammer tous paramètres :
     - Type chaudière, puissance
     - Type gaz (G20/G25)
     - Paramètres spécifiques installation
   - Tester stabilité après réinit

9. **Mise à jour firmware**
   - Vérifier version firmware actuelle (menu diagnostic)
   - Consulter site constructeur : firmware plus récent disponible ?
   - Mise à jour via :
     - Interface USB (selon marque)
     - Remplacement EPROM (ancien modèles)
     - Outil SAV constructeur
   - Attention : mise à jour ratée peut bricker la carte

10. **Protection contre micro-coupures**
    - Installer onduleur (UPS) :
      - Puissance 300-500VA suffisante (carte électronique seule)
      - Maintient alimentation lors coupures brèves
    - Ou stabilisateur/régulateur tension
    - Protège aussi contre surtensions

**Prévention :**
- Remplacement pile RTC préventif tous les 5 ans
- Protection secteur (parafoudre + onduleur)
- Vérification qualité alimentation électrique
- Mise à jour firmware régulière

**Coût solutions :**
- Pile CR2032 : 1-3€
- Onduleur 500VA : 50-100€
- Mise à jour firmware : gratuite (si matériel compatible)
- Remplacement carte : 150-600€

**Paramètres à sauvegarder avant réinit :**
- Photographier tous écrans paramètres
- Noter sur papier : puissance, type gaz, courbe chauffe, etc.
- Permet reprogrammation rapide après reset

---

