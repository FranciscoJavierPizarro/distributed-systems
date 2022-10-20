/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	20/10/2022
* Descripción del fichero:
*	Programa empleado para lanzar un proceso escritor
*	Se usa para ejecutar escritores de forma remota
* Descripción de la estructura del código:
*	1. Importar paquetes
*	2. Main, lanza proceso escritor
*/
package main

//====================================================================
//	IMPORTS
//====================================================================

import (
	"escritor"
	"os"
	"log"
	"strconv"
)

//====================================================================
//	MAIN
//====================================================================

func main() {
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Uso: ./main pid nProc")
	}
	nProc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Uso: ./main pid nProc")
	}
	escritor.Start(pid, nProc)
}