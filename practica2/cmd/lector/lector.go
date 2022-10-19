package lector

import (
	"fmt"
	"io/ioutil"
	"os"
	"ra"
	"log"
	"math/rand"
	"time"
	"strconv"
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

func sincronize(ownRa *ra.RASharedDB) {
	ownRa.SendSignal()
	ownRa.ReceiveSync()
}

func Start(pid int, nProc int) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()

	sincronize(ownRa)

	for i := 0; i < 10; i++ {
		r := rand.Intn(500)
		time.Sleep(time.Duration(r) * time.Millisecond)
		ownRa.PreProtocol(false)

		log.Printf("Hi i am reader %d reading \n" +
		ReadF(strconv.Itoa(pid)+".txt"), pid)

		ownRa.PostProtocol()
	}

	sincronize(ownRa)

	log.Printf("Process %d end.", pid)

	// ownRa.Stop() //bug
}
