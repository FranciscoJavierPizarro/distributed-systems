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
		fmt.Fprintf(os.Stderr, "Error opening file.")
		os.Exit(1)
	}
	_, err = f.WriteString(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file.")
		os.Exit(1)
	}
	f.Close()
}

func sincronize(ownRa *ra.RASharedDB) {
	ownRa.SendSignal()
	<-ownRa.Syncronized
}

func Start(pid int, nProc int) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	sincronize(ownRa)

	for i := 0; i < 15; i++ {
		r := rand.Intn(500)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(true)
		
		log.Printf("I am writer %d,writing on file.", pid)
		WriteF("pachanga.txt", "Hi, i am writer " + strconv.Itoa(pid)+"\n")

		ownRa.PostProtocol()
	}

	sincronize(ownRa)
			
	log.Printf("Proceso %d end.", pid)
	ownRa.Stop() //bug
	
}

