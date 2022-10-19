package main

import (
	"log"
	"os"
	"lector"
	"escritor"
	"ms"
)

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

func main() {
	numEscritores := 10
	numLectores := 10
	nProc := numEscritores + numLectores
	messageTypes := []ms.Message{ms.SyncSignal{}, ms.SyncWait{}}
    msgs := ms.New(nProc + 1, "users.txt", messageTypes)
	barrier := make(chan bool)
	log.Println("START")
	go barrierSyncro(msgs,nProc,barrier)
	for i := 0; i < nProc; i++ {
		if(i < numLectores) {
			go lector.Start(i + 1, nProc)
		}else {go escritor.Start(i + 1, nProc)}
		log.Printf("Process launched with PID %d.", i)
	}

	<- barrier
	log.Println("READY TO END")
	//meter en hilo o problemas añadir canal
	go barrierSyncro(msgs,nProc,barrier)
	<- barrier
	log.Println("END")
}