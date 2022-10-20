/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	20/10/2022
* Descripción del fichero:
*	Archivo principal de código, este se encarga de definir
*	el número de escritores y de lectores, así como punto de
*	control en las barreras de sincronización
*	Se emplea para lanzar o bien localmente o bien remotamente
*	todos los procesos necesarios
*	Para lanzar en remoto emplear flag --remote
* Descripción de la estructura del código:
*	1. Importar paquetes
*	2. Funciones auxiliares del programa principal
*	3. Programa principal
*/

package main

//====================================================================
//	IMPORTS
//====================================================================

import (
	"log"
	"os"
	"lector"
	"escritor"
	"ms"
	"os/exec"
	"flag"
	"strconv"
)

//====================================================================
//	FUNCIONES AUXILIARES
//====================================================================

//Crea una barrera de sincronización, espera a que todos los procesos
//lleguen, entonces avisa a todos con el paquete SyncWait de que pueden
//seguir, avisa al programa principal por el canal barrier
func barrierSyncro(msgs ms.MessageSystem, nProc int, barrier chan bool) {
	for i := 0; i < nProc; i++ {
		switch msgs.Receive().(type) {
			case ms.SyncSignal:
				log.Println("Process ready to barrier.")
			default:
				log.Println(os.Stderr, "Error wrong comms to barrier.")
				os.Exit(1)
		}
	}

	for i := 0; i < nProc; i++ {
		msgs.Send(i + 1, ms.SyncWait{})
	}

	barrier <- true
}


//LLama al script de lanzamiento remoto con los parámetros necesarios
//para crear un escritor/lector remoto
func turnOnRemote(typeOfProc string, pid int, nProc int) {
	log.Printf("Launching process with PID %d.", pid)
	comm:= "./launchProcess.sh"
	cmd := exec.Command(comm,typeOfProc,strconv.Itoa(pid),strconv.Itoa(nProc))
    err := cmd.Run()
    if err != nil {
		panic(err)
    }
	return
}

//Lanza todos los procesos, dependiendo de remote los lanza en local
//o en remoto
func iniciarProcesos(remote bool,nProc int, numLectores int) {
	if (remote) {
		for i := 0; i < nProc; i++ {
			if(i < numLectores) {
				go turnOnRemote("reader",i + 1, nProc)
			}else {go turnOnRemote("writer",i + 1, nProc)}
			
		}
	} else {
		for i := 0; i < nProc; i++ {
			if(i < numLectores) {
				go lector.Start(i + 1, nProc)
			}else {go escritor.Start(i + 1, nProc)}
			log.Printf("Process launched with PID %d.", i)
		}
	}
}


//====================================================================
//	PROGRAMA PRINCIPAL
//====================================================================

func main() {
	//setup de valores configurables
	numEscritores := 3
	numLectores := 3
	//setup de flags, constantes y creación del ms
	remote := flag.Bool("remote",false,"This instance of the program is using remote protocol")
	flag.Parse()
	nProc := numEscritores + numLectores
	messageTypes := []ms.Message{ms.SyncSignal{}, ms.SyncWait{}}
    msgs := ms.New(nProc + 1, "users.txt", messageTypes)
	barrier := make(chan bool)
	
	//lanzar procesos en remoto o local y tener lista la escucha de la barrera
	//para que antes de lanzar ninguno ya este disponible
	go barrierSyncro(msgs,nProc,barrier)
	log.Println("START")
	if (*remote) {
		iniciarProcesos(true,nProc,numLectores)
	} else {
		iniciarProcesos(false,nProc,numLectores)
	}
	//sincronización de todos los procesos para empezar
	<- barrier
	log.Println("ALL PROCESS RUNNING")
	//sincronización de todos los procesos para finalizar
	go barrierSyncro(msgs,nProc,barrier)
	<- barrier
	log.Println("END")
}