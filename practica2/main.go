package main

import (
	//"fmt"
	"log"
	"lector"
	"escritor"
)


func main() {
	nProc := 20
	numEscritores := 10
	numLectores := 10
	espera := make(chan bool, nProc)
	empezar := make(chan bool, nProc)
	terminado := make(chan bool, nProc)
	barreraFin := make([]chan bool, nProc)
	
	for i := range barreraFin {
		barreraFin[i] = make(chan bool)
	}

	
	for i := 0; i < numLectores; i++ {
		go lector.Start(i + 1, nProc, empezar, espera, terminado, barreraFin)
		log.Printf("lanzo lector con PID %d.", i)
	}

	for i := 0; i <= numEscritores; i++ { // m escritores
		go escritor.Start(i + 1, nProc, empezar, espera, terminado, barreraFin)
		log.Printf("lanzo escritor con PID %d.", i)
	}
	log.Println("llegue1")
	for i := 0; i < nProc; i++ {
		<-espera
	}
	log.Printf("llegue2")
	for i := 0; i < nProc; i++ {
		empezar <- true
	}
	log.Printf("llegue")
	for i := 0;i < nProc; i++ {
		<-terminado
	}
	

	log.Println("End")
}