package lector

import (
	"fmt"
	"io/ioutil"
	"os"
	"ra"
	// "log"
	// "math/rand"
	// "strconv"
	// "time"
)

//Función auxiliar de lectura de fichero
func ReadF(file string) string {
	buffer, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al leer el fichero.")
		os.Exit(1)
	}
	return string(buffer)
}

func Start(pid int, nProc int,
	run chan bool, readyToRun chan bool,
	end chan bool, endBarrier  []chan bool) {
	ownRa := ra.New(pid, "users.txt", nProc)
	go ownRa.ReceiveMsg()
	// // Barrera de inicialización
	readyToRun <- true
	<-run
	// for i := 0; i < 1000; i++ {
	// 	r := rand.Intn(1000)
	// 	time.Sleep(time.Duration(r) * time.Millisecond)
	// 	ownRa.PreProtocol(false)

	// 	log.Println("PID:" + strconv.Itoa(pid%5) + ",OP:READ \n" +
	// 	ReadF("pachanga.txt"))

	// 	ownRa.PostProtocol()
	// }

	// // Barrera de fin
	// if (pid != nProc) {<-endBarrier[pid-1]}
	// if (pid != 1) {endBarrier[pid-2] <- true}

	// //ownRa.Stop() //bug
	// end <- true
}
