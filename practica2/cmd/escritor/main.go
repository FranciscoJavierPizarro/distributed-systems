package main

import (
	"fmt"
	"os"
)


// Escribe en el fichero indicado un fragmento de texto al final del mismo
func WriteF(file string, text string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0666)
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

func start(pid int, nProc int,
	run chan bool, readyToRun chan bool,
	end chan bool, endBarrier  []chan bool) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	// Barrera de inicialización
	readyToRun <- true
	<-run
	for i := 0; i < 1000; i++ {
		r := rand.Intn(1000)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(true)

		log.Printf("soy el escritor %d, escribo en el fichero
		 y mando peticiones.", pid%5)
		WriteF("pachanga.txt"+  ".txt", "Hola, soy el proceso " + strconv.Itoa(pid)+"\n")

		ownRa.PostProtocol()
	}

	// Barrera de fin
	if (pid != nProc) <-endBarrier[pid-1]
	if (pid != 1) endBarrier[pid-2] <- true

	//ownRa.Stop() //bug
	end <- true
}


func main() {

}
