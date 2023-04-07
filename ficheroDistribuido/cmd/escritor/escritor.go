/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	20/10/2022
* Descripción del fichero:
*	Contiene todo el código principal de un proceso escritor
*	Crea la estructura de Ricart-Agrawala
*	Se sincroniza con los demás procesos al empezar y acabar
*	Itera entrando en SC usando el algoritmo de Ricart-Agrawala generalizado
*	en cada iteración escribe nuevo contenido en el fichero
* Descripción de la estructura del código:
*	1. Importar paquetes 
*	2. Funciones auxiliares del código principal de cada escritor
*	3. Función START, la cual ejecuta un proceso escritor
*/
package escritor

//====================================================================
//	IMPORTS
//====================================================================

import (
	"fmt"
	"os"
	"ra"
	"log"
	"math/rand"
	"strconv"
	"time"
)

//====================================================================
//	FUNCIONES AUXILIARES
//====================================================================

//Añade en el fichero el texto deseado
func WriteF(file string, text string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file.")
		os.Exit(1)
	}
	_, err = f.WriteString(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file.")
		os.Exit(1)
	}
	f.Close()
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
	text:="Hi, i am writer " + strconv.Itoa(pid)+"\n"
	go ownRa.ReceiveMsg()

	sincronize(ownRa)
	
	for i := 0; i < 10; i++ {
		r := rand.Intn(500)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(true)
		
		log.Printf("I am writer %d,writing on file.", pid)
		WriteF(strconv.Itoa(pid)+".txt", text)
		ownRa.UpdateFile(text)
		ownRa.WaitConfirms()
		ownRa.PostProtocol()
	}

	sincronize(ownRa)
			
	log.Printf("Proceso %d end.", pid)
	//ownRa.Stop() //bug
	
}

