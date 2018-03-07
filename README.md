# Postgres-MPD (French)

## Préparation de l'environnement de build

Pour compiler il faut d'abord installer un environnement GoLang, pour cela se rendre sur le site officiel [*https://golang.org/dl/*](https://golang.org/dl/) afin de le télécharger.

La compilation de l'utilitaire demande d'avoir la librairie PostgreSQL installée pour cela se placer dans un répertoire, par exemple votre HOME et récupérer la librairie.

   -   cd ~/
   -   mkdir GoTools && cd GoTools/
   -   mkdir bin pkg src
   -   export GOPATH=`pwd`
   -   go get -u github.com/lib/pq

## Compilation

La compilation va se faire dans le répertoire ou vous allez cloner le projet, par exemple si c'est dans votre HOME

   -   cd ~/
   -   git clone .... (reprendre l'URL indiquée par github)
   -   cd Postgres-MPD/GoWork
   -   export GOPATH=~/GoTools/:`pwd`
   -   go build postgres-mpd

## Installation

Il suffit de copier le binaire dans le répertoire Postgres-MPD/Example/Manitou/

## Utilisation

Extract and draw tables relations graph from existing Postgres database 

Ok
