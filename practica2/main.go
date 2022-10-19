package main

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

func barrierSyncro2(msgs ms.MessageSystem, nProc int, barrier chan bool) {
	for i := 0; i < nProc; i++ {
		switch msgs.Receive().(type) {
			case ms.SyncSignal:
				log.Println("Process ENDBARRIER.")
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


//Enciende los workers
func turnOnRemote(typeOfProc string, pid int, nProc int) {
	//LLama a ./launchWorkers.sh
	comm:= "./launchProcess.sh " + typeOfProc + " " +  strconv.Itoa(pid) + " " +  strconv.Itoa(nProc)
	cmd := exec.Command(comm)
    err := cmd.Run()
    if err != nil {
		panic(err)
    }
	log.Printf("Process launched with PID %d.", pid)
	return
}

func main() {
	remote := flag.Bool("remote",false,"This instance of the program is using remote protocol")
	flag.Parse()

	
	numEscritores := 1
	numLectores := 1
	nProc := numEscritores + numLectores
	messageTypes := []ms.Message{ms.SyncSignal{}, ms.SyncWait{}}
    msgs := ms.New(nProc + 1, "users.txt", messageTypes)
	barrier := make(chan bool)
	log.Println("START")
	go barrierSyncro(msgs,nProc,barrier)
	
	if (*remote) {
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

	<- barrier
	log.Println("READY TO END")
	//meter en hilo o problemas añadir canal
	go barrierSyncro2(msgs,nProc,barrier)
	<- barrier
	log.Println("END")
}