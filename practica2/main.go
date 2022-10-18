package main

import (
	//"fmt"
	"log"
	"lector"
	"escritor"
)


func main() {
	numEscritores := 10
	numLectores := 10
	nProc := numEscritores + numLectores
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

	for i := 0 + numLectores; i < nProc; i++ { // m escritores
		go escritor.Start(i + 1, nProc, empezar, espera, terminado, barreraFin)
		log.Printf("lanzo escritor con PID %d.", i)
	}

	for i := 0; i < nProc; i++ {
		<-espera
	}

	for i := 0; i < nProc; i++ {
		empezar <- true
	}
	log.Printf("Barrera alcanzada")
	for i := 0;i < nProc; i++ {
		<-terminado
	}
	

	log.Println("End")
}