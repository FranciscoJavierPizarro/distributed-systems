/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	20/10/2022
* Descripción del fichero:
*	Contiene todo el código principal de un proceso lector
*	Crea la estructura de Ricart-Agrawala
*	Se sincroniza con los demás procesos al empezar y acabar
*	Itera entrando en SC usando el algoritmo de Ricart-Agrawala generalizado
*	en cada iteración lee el contenido del fichero
* Descripción de la estructura del código:
*	1. Importar paquetes 
*	2. Funciones auxiliares del código principal de cada lector
*	3. Función START, la cual ejecuta un proceso lector
*/
package lector

//====================================================================
//	IMPORTS
//====================================================================

import (
	"fmt"
	"io/ioutil"
	"os"
	"ra"
	"log"
	"math/rand"
	"time"
	"strconv"
)

//====================================================================
//	FUNCIONES AUXILIARES
//====================================================================

//Función auxiliar de lectura de fichero
func ReadF(file string) string {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file.")
		os.Exit(1)
	}
	return string(content)
}

func sincronize(ownRa *ra.RASharedDB) {
	ownRa.SendSignal()
	ownRa.ReceiveSync()
}

//====================================================================
//	FUNCIÓN PRINCIPAL
//====================================================================

func Start(pid int, nProc int) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	sincronize(ownRa)

	for i := 0; i < 10; i++ {
		r := rand.Intn(500)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(false)

		log.Printf("Hi i am reader %d reading \n" +
		ReadF(strconv.Itoa(pid)+".txt"), pid)

		ownRa.PostProtocol()
	}

	sincronize(ownRa)

	log.Printf("Process %d end.", pid)

	//ownRa.Stop() //bug
}
