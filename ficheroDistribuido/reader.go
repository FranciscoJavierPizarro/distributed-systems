/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	20/10/2022
* Descripción del fichero:
*	Programa empleado para lanzar un proceso lector
*	Se usa para ejecutar lectores de forma remota
* Descripción de la estructura del código:
*	1. Importar paquetes
*	2. Main, lanza proceso lector
*/
package main

//====================================================================
//	IMPORTS
//====================================================================

import (
	"lector"
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
	lector.Start(pid, nProc)
}