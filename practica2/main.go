package main

import (
	"log"
	"lector"
	"escritor"
)


func main() {
	numEscritores := 10
	numLectores := 10
	nProc := numEscritores + numLectores
	wait := make(chan bool, nProc)
	ready := make(chan bool, nProc)
	end := make(chan bool, nProc)
	readyToEnd := make([]chan bool, nProc)
	
	log.Println("START")
	for i := 0; i < nProc; i++ {
		readyToEnd[i] = make(chan bool)
		if(i < numLectores) {
			go lector.Start(i + 1, nProc, ready, wait, end, readyToEnd)
		}else {go escritor.Start(i + 1, nProc, ready, wait, end, readyToEnd)}
		log.Printf("Process launched with PID %d.", i)
	}

	for i := 0; i < nProc; i++ {
		<-wait
	}

	for i := 0; i < nProc; i++ {
		ready <- true
	}
	readyToEnd[nProc - 1] <- true
	log.Println("Ready to end")
	<-end

	log.Println("END")
}