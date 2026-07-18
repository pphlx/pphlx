# Guide de traduction de la version française d'Astro Docs

Ce guide est destiné aux personnes souhaitant participer à la traduction en français de la documentation d'Astro.

Son objectif est d'aider à maintenir une traduction homogène au sein de la documentation et de l'écosystème d'Astro (par exemple [Starlight](https://starlight.astro.build/)).

## 🧩 Participer à la traduction

N'hésitez-pas à rejoindre le [serveur Discord officiel d'Astro](https://pphlx.org/chat), et notamment le fil [`i18n-fr`](https://discord.com/channels/830184174198718474/971729943786561556), pour obtenir de l'aide, demander des conseils et vous coordonner avec les autres traducteurs.

## 📖 Glossaire

Le glossaire permet d'établir une liste des traductions convenues et des mots ne nécessitant pas de traduction parce que considérés comme propre à Astro ou globalement acceptés.

> [!WARNING]
> Ce glossaire est vivant : n'hésitez pas à l'enrichir quand le besoin se fait sentir !

### 🔄️ Mots ne nécessitant pas de traductions

Les mots se référant à des technologies ne devraient pas être traduits, qu'il s'agisse de langages, de bibliothèques ou de frameworks (p. ex. Markdown, Typescript, React, Svelte, Intellisense, etc.).

#### Concepts d'Astro

| Terme | Condition | Explications/Référence (facultatif) |
|-------|-----------|-------------------------------------|
| Fragment | Lorsqu'il fait référence aux balises `<Fragment> </Fragment>` ou `<> </>`. | La balise ne peut pas être traduite. |
| frontmatter | Tout le temps. | Peut être considéré comme l'en-ête du fichier (Astro ou Markdown), délimité par trois tirets (`---`). Il ne semble pas être traduit ailleurs. |
| props | Lorsqu'il fait référence aux propriétés/attributs d'un composant. | « props » est un terme courant dans la Jamstack, pour faciliter la migration nous avons choisi de ne pas le traduire (ref. [discussion sur Discord](https://discord.com/channels/830184174198718474/971729943786561556/1286078277810782289)). |
| slot | Lorsqu'il est employé en tant que nom et qu'il fait référence à la balise `<slot/>`. | La balise ne peut pas être traduite. |

#### Autres mots non traduits

Il existe d'autres mots ne possédant pas de traduction officielle ou dont le terme recommandé ne fait pas consensus et est peu utilisé ailleurs. Dans ces cas-là, il est préférable de garder le mot anglais qui sera sans doute plus compréhensible.

| Terme | Définition | Explications/Référence (facultatif) |
|-------|------------|-------------------------------------|
| framework | Fait référence à une infrastructure logicielle permettant la conception d'applications (p. ex. : `Angular`, `Vue`, `Svelte`, etc…) | Bien que « cadriciel » soit la traduction officielle, elle est rarement utilisée ailleurs. |
| front-end | Fait référence à l'interface utilisateur d'une application. | Bien que « frontal » soit une traduction possible, elle est rarement utilisée ailleurs. |
| back-end | Fait référence à la partie serveur d'une application. | Bien que « dorsal » soit une traduction possible, elle est rarement utilisée ailleurs. |
| middleware | Désigne un logiciel qui permet de coordonner le fonctionnement de plusieurs logiciels. | Bien que « logiciel médiateur » et « intergiciel » soient les traductions recommandées, elles sont rarement utilisées ailleurs. |
| hook | Désigne un point d'ancrage permettant à l'utilisateur d'exécuter du code à un endroit précis du code source. | « crochet » n'apporte pas suffisamment de sens et il est rarement traduit ailleurs. Précédente discussion : https://github.com/withastro/docs/pull/11593#discussion_r2074976678 |
| headless | Désigne une architecture où le front-end et le back-end sont proposés séparément (p ex. un CMS pour gérer les contenus et une application Astro pour l'afficher). | Il n'existe pas de traduction officielle et nous pensons que « sans-tête » n'est pas suffisamment porteur de sens. Précédente discussion : https://github.com/withastro/docs/pull/11593#discussion_r2074967636 |
| glob / globbing | Désigne une fonction ou processus permettant de faire correspondre des motifs à des noms de répertoire ou de fichier en utilisant des caractères géneriques (aussi connu sous le nom de « [appariement de formes](https://www.culture.fr/franceterme/terme/INFO82) ») | Nous n'avons pas de traduction officielle pour ce terme et utiliser « global » serait une erreur de traduction. Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112787624 |

###  🔤 Traductions courantes

#### Concepts et API d'Astro

| Anglais | Français | Explications/Référence (facultatif) |
|---------|----------|-------------------------------------|
| Astro Islands | Îlots Astro | Précédente discussion : https://github.com/withastro/docs/pull/11593#discussion_r2074910929 |
| Content Collections | Collections de contenu | |
| Content Layer | Couche de contenu | |
| Fonts | Polices ou Polices d'écriture | |
| Island architecture | Architecture en îlots | |
| Sessions | Sessions | |

#### Langage courant

Certains mots ont un équivalent français qui devrait être utilisé uniformément dans les différentes traductions.

| Anglais | Français | Explications/Référence (facultatif) |
|---------|----------|-------------------------------------|
| assets | ressources | « actif » vient du [monde de la finance](https://www.culture.fr/franceterme/Resultats-de-recherche?q=asset&domaine=0). Il est préférable d'utiliser « [ressource](https://fr.wiktionary.org/wiki/asset) ». |
| breaking changes | changements non rétrocompatibles / changements avec rupture de compatibilité | « changements majeurs » est acceptable pour éviter les répétitions tant qu'on garde la notion que quelque chose peut « casser » dans le titre/texte. Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112817635 |
| build | Suivant le contexte : création / construction / compilation. En référence à `astro build`, on privilégiera le terme compilation.  | |
| bundle | regroupement | |
| changelog | journal des modifications | |
| (the) CLI / (the) command line interface | (la) CLI / (l')interface en ligne de commande | |
| client-side | côté client | |
| code fences (`---`) | délimiteur de code | Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2114303905 |
| deprecation | dépréciation | Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2114666199 |
| docs | documentation | |
| endpoint | point de terminaison | |
| export | exportation | |
| feedback | retour / réaction / commentaire | |
| flag | option | |
| footer | pied de page | |
| header | en-tête | |
| heading | titre ou en-tête | Nous n'avons pas de consensus (« titre » dans Astro Docs et « en-tête » dans Starlight). Précédente discussion : https://github.com/withastro/starlight/pull/2884#discussion_r1957379942 |
| import | importation | |
| inline | au sein de / dans le corps / intégré à | La traduction dépend fortement du contexte, les propositions précédentes ne sont que des exemples possibles. Il est préférable de trouver une alternative à « en ligne » dans la majorité des cas. Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112802152 |
| issue | problème ou, en référence à la fonctionnalité de Github, ticket | [MDN](https://developer.mozilla.org/fr/docs/MDN/Community/Issues) utilise également « ticket » pour décrire la fonctionnalité de Github |
| layout | mise en page | |
| live | en ligne / opérationnel | « live » est généralement utilisé pour différencier développement/production donc « en direct » est rarement la bonne traduction. Il peut être préférable d'omettre ce mot si le contexte est clair. |
| on-demand rendering | rendu à la demande | |
| overlay | par superposition / (fenêtre) superposée | |
| package | paquet | |
| pattern | modèle / (en référence aux expressions régulières et à glob) motif ou formule | D'autres traductions peuvent être acceptables en fonction du contexte. Par exemple, « formule » semblait plus appropriée [dans le tutoriel](https://pphlx.org/docs/fr/tutorial/2-pages/3/#analyser-la-formule). Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112789565 |
| placeholder | mot + par défaut / réservé / substituable / fictif | La traduction dépend fortement du contexte, les propositions précédentes ne sont que des exemples possibles. Précédente discussion : https://github.com/withastro/docs/pull/11593#discussion_r2071856903 |
| plugin | module d'extension | |
| preset | préréglage | « préconfiguration » pourrait convenir mais dans certains cas la formulation devient lourde. Discuté dans https://github.com/withastro/starlight/pull/3126#discussion_r2046673972 |
| prop | propriété | |
| props | props ou propriétés | En référence à un composant privilégiez « props », autrement utilisez « propriétés » (voir [Concepts d'Astro](#concepts-dastro)) |
| output | sortie ou affichage | |
| re-render | rafraîchir / effectuer un nouveau rendu | |
| recipes | recettes | |
| renderer | moteur de rendu | |
| rendering | rendu / affichage / restitution | |
| repository | dépôt | |
| routing | routage | |
| runtime | Suivant le contexte : moteur d'exécution / environnement d'exécution / au moment de l'exécution | |
| scope | limité / délimité | |
| scoped | à portée limitée | |
| server-side | coté serveur | |
| template | modèle | |
| to build | Suivant le contexte : créer / construire / compiler | |
| to bundle | regrouper | |
| to commit | valider | Il s'agit de la traduction utilisée dans l'interface française de VS Code. |
| to deprecate | déprécier | Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2114666199 |
| to export | exporter | |
| to fetch | récupérer | |
| to import | importer | |
| to prefetch | précharger | |
| to prerender | prégénérer | |
| to render | afficher / générer / effectuer un rendu / restituer | |
| to slot | inclure / injecter / insérer | |
| to store | stocker / enregistrer / sauvegarder / conserver | |
| to style | mettre en forme / appliquer des styles | |
| to support (quelque chose) | prendre en charge / être compatible avec / fonctionner sous | Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112812611 |
| to support (une personne) | assister | Précédente discussion : https://github.com/withastro/docs/pull/11795#discussion_r2112812611 |
| to type | écrire / saisir / (en référence à Typescript) définir un type | |
| to wrap | englober / envelopper | |
| type safe | sûreté du typage | |
| UI | UI | |
| update | mise à jour | |
| upgrade | mise à niveau | |
| view transitions | transitions de vue | |

> [!NOTE]
> Ces traductions sont des recommandations. Il est parfois acceptable d'utiliser une autre traduction si elle est plus adaptée au contexte ou elle rend la lecture plus fluide en évitant des répétitions.

> [!TIP]
> Il est parfois utile pour faciliter la recherche ou la compréhension d'un concept de proposer à certains endroits, avec parcimonie, deux versions. Par exemple : `the component props` pourrait être traduit de la façon suivante `les props (ou « propriétés ») du composant`.

## 📝 Conseils stylistiques

### Les titres

Certains guides de style anglais préconisent l'utilisation de majuscules pour chaque mot d'un titre. Ces règles ne s'appliquent pas en français. Il est donc recommandé de n'utiliser une majuscule qu'en début de titre, à l'exception des noms propres et des composants.

Privilégiez également l'usage de l'infinitif à celui de l'impératif lors de la traduction des titres.

### Les guillemets

Bien que difficilement accessibles sur un clavier, privilégiez les guillements francophones (`« texte »`) aux guillemets doubles (`"texte"`).

Mémo Unicode :
* `«` : <kbd>U+00ab</kbd>
* `»` : <kbd>U+00bb</kbd>

### Code utilisé comme mot dans une phrase

Dans la version anglaise, vous verrez régulièrement des mots entre accents graves `` ` ``. Cette syntaxe n'a pas toujours de sens en français, il est donc parfois préférable de remplacer le code par un mot français et de rajouter la version code à côté.

**Par exemple :** ``The file `path`...`` devrait être traduit par ``Le chemin du fichier (`path`)...``.

### Liens

#### Internes

Tous les liens doivent rediriger vers la version française, pensez donc à remplacer `/en/` par `/fr/` même si la page de destination n'est pas encore traduite. Vous devez également pensez à traduire la partie après `#`, quand elle existe : elle correspond a l'id généré pour un titre sur la page. Vous pouvez utiliser le serveur de développement pour vérifier l'id généré.

#### Externes

* L'ancre (ou texte) du lien est toujours traduit.
* Le lien ne change pas sauf s'il existe une version française. Il est donc préférable de vérifier le lien au moment de la traduction.
* S'il n'existe pas de version française, il peut être utile de rajouter la langue après le lien (p. ex. `texte du lien (Anglais)`) pour informer le lecteur que la page de destination est dans une autre langue.

### Les exemples

Pour que l'exemple reste compréhensible pour les moins anglophones des lecteurs, il est important de traduire les commentaires. Il peut également être intéressant d'adapter les chemins et les noms des variables si vous jugez que ça facilite la compréhension.

Lorsque vous modifiez un exemple, pensez à vérifier les arguments passés au bloc de code pour également [mettre à jour la mise en évidence](https://expressive-code.com/key-features/text-markers/#usage-in-markdown--mdx) si besoin.

## 📚 Ressources

* En cas de doute pour une traduction, n'hésitez pas à consulter ce qui est fait dans d'autres documentations (MDN, React, Vue, etc.).
* [FranceTerme](https://www.culture.fr/franceterme)
* [Vitrine linguistique de l'Office québécois](https://vitrinelinguistique.oqlf.gouv.qc.ca/) : Bien que destiné aux québecois, les recommandations françaises sont souvent affichées. En utilisant les « termes privilégiés », vous devriez trouver une traduction pouvant convenir à tout le monde.
