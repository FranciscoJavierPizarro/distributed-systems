package escritor

import (
	"fmt"
	"os"
	"ra"
	"log"
	"math/rand"
	"strconv"
	"time"
)


// Escribe en el fichero indicado un fragmento de texto al final del mismo
func WriteF(file string, text string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al abrir el fichero.")
		os.Exit(1)
	}
	_, err = f.WriteString(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al escribir en el fichero.")
		os.Exit(1)
	}
	f.Close()
}

func Start(pid int, nProc int,
	run chan bool, readyToRun chan bool,
	end chan bool, endBarrier  []chan bool) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	// // Barrera de inicialización
	readyToRun <- true
	<-run
	for i := 0; i < 15; i++ {
		r := rand.Intn(1000)
		time.Sleep(time.Duration(r) * time.Millisecond)
		log.Printf("soy el escritor %d, mando peticiones.", pid)
		ownRa.PreProtocol(true)
		
		WriteF("pachanga.txt", "Hola, soy el proceso " + strconv.Itoa(pid)+"\n")

		ownRa.PostProtocol()
		log.Printf("soy el escritor %d, y respondo peticiones.", pid)
	}

	// Barrera de fin
	if (pid != nProc) {
		log.Printf("soy el proceso %d, y me bloqueo en la barrera.", pid)
		<-endBarrier[pid-1]}
	log.Printf("soy el proceso %d, y he recibido la barrera.", pid)
	if (pid != 1) {endBarrier[pid-2] <- true}

	//ownRa.Stop() //bug
	end <- true
}

