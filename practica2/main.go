package main

import (
	"log"
	"os"
	"lector"
	"escritor"
	"ms"
)

func main() {
	numEscritores := 10
	numLectores := 10
	nProc := numEscritores + numLectores
	messageTypes := []ms.Message{ms.SyncSignal{}, ms.SyncWait{}}
    msgs := ms.New(nProc + 1, "users.txt", messageTypes)
	

	end := make(chan bool, nProc)
	readyToEnd := make([]chan bool, nProc)
	
	log.Println("START")
	for i := 0; i < nProc; i++ {
		readyToEnd[i] = make(chan bool)
		if(i < numLectores) {
			go lector.Start(i + 1, nProc, end, readyToEnd)
		}else {go escritor.Start(i + 1, nProc, end, readyToEnd)}
		log.Printf("Process launched with PID %d.", i)
	}

	for i := 0; i < nProc; i++ {
		switch msgs.Receive().(type) {
			case ms.SyncSignal:
				log.Println("Process ready to start.")
			default:
				log.Println(os.Stderr, "Error wrong comms to start.")
				os.Exit(1)
		}
	}

	for i := 0; i < nProc; i++ {
		msgs.Send(i + 1, ms.SyncWait{})
	}
	readyToEnd[nProc - 1] <- true
	log.Println("Ready to end")
	<-end

	log.Println("END")
}