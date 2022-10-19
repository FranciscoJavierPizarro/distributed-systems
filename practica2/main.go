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
	
	log.Println("START")
	for i := 0; i < nProc; i++ {
		if(i < numLectores) {
			go lector.Start(i + 1, nProc)
		}else {go escritor.Start(i + 1, nProc)}
		log.Printf("Process launched with PID %d.", i)
	}

	//meter en hilo o problemas añadir canal
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


	// readyToEnd[nProc - 1] <- true
	// msgs.Send(nProc, ms.EndSignal{})
	log.Println("Ready to end")
	// <-end
	// switch msgs.Receive().(type) {
	// 	case ms.EndSignal:
	// 		log.Println("END")
	// 	default:
	// 		log.Println(os.Stderr, "Error wrong comms to end.")
	// 		os.Exit(1)
	// }


	//meter en hilo o problemas añadir canal
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
	log.Println("END")
}