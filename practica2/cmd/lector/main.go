package lector

import (
	"fmt"
	"io/ioutil"
	"os"
	"ra"
	"log"
	"math/rand"
	"time"
)

//Función auxiliar de lectura de fichero
func ReadF(file string) string {
	buffer, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file.")
		os.Exit(1)
	}
	return string(buffer)
}

func Start(pid int, nProc int,
	end chan bool, readyToEnd  []chan bool) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	ownRa.SendSignal()
	<-ownRa.Syncronized
	for i := 0; i < 15; i++ {
		r := rand.Intn(500)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(false)

		log.Printf("Hi i am reader %d reading \n" +
		ReadF("pachanga.txt"), pid)

		ownRa.PostProtocol()
	}


	<-readyToEnd[pid-1]
	if (pid != 1) {readyToEnd[pid-2] <- true
	}else {end <- true}
			
	log.Printf("Process %d end.", pid)

	//ownRa.Stop() //bug
}
