/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes a la práctica 1
*/
package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	//"io"
	//"com"
	"practica1/com"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

//Recibe una peticion y la maneja, tira la conexión en caso necesario
func ManageRequest(decoder *gob.Decoder,encoder *gob.Encoder,conn net.Conn)(bool){
	var request com.Request
	err := decoder.Decode(&request)
	checkError(err)
	if(request.Id == -1) {
		conn.Close()
		return false
	} else {
		Reply := com.Reply{request.Id,FindPrimes(request.Interval)}
		errRep := encoder.Encode(Reply)
		checkError(errRep)
		conn.Close()
		return true
	}
}

func secuentialServer(listener net.Listener) {
	var b bool = true
	fmt.Println("Servidor secuencial")
	for b{
		conn, err := listener.Accept()
		checkError(err)
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		b = ManageRequest(decoder,encoder,conn)
	}
}

//Apaño para multithread
func ManageRequest_TH(decoder *gob.Decoder,encoder *gob.Encoder,conn net.Conn,done chan bool){
	defer conn.Close()
	var request com.Request
	err := decoder.Decode(&request)
	checkError(err)
	if(request.Id == -1) {
		done <- true
	} else {
		done <- false
		Reply := com.Reply{request.Id,FindPrimes(request.Interval)}
		errRep := encoder.Encode(Reply)
		checkError(errRep)
	}
}

func threadedServer(listener net.Listener) {
	fmt.Println("Servidor multithread")
	done := make(chan bool)
	var first bool = true

	for {
		//Apaño para que no lea de un chan vacio
		if (first) {
			first = false
		} else {
			c := <- done
			if (c){
				break
			}
		}

		conn, err := listener.Accept()
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		checkError(err)
		go ManageRequest_TH(decoder,encoder,conn,done)
	}
}

func handleComms(ch chan net.Conn) {
	for {
		conn := <-ch
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		ManageRequest(decoder,encoder,conn)
	}
}

func poolThreadedServer(listener net.Listener) {
	fmt.Println("Servidor pool threads")
	ch := make(chan net.Conn)
	NTHREAD := 10
	for i:=0;i<NTHREAD;i++ {
		go handleComms(ch)
	}
	
	var b bool = true
	for b{
		conn, err := listener.Accept()
		checkError(err)
		ch <- conn
	}
}

func main() {
	var option int
	fmt.Print("Opciones disponibles: \n\t0 servidor secuencial\n\t1 servidor multithread\n\t2 servidor con pool de threads\n\t3 servidor con workers(ssh)\n\n")
	fmt.Scan(&option)
	fmt.Println("") 
	CONN_TYPE := "tcp"
	CONN_HOST := "127.0.0.1"
	CONN_PORT := "30000"
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	switch option {
	case 0:
		secuentialServer(listener)
	case 1:
		threadedServer(listener)
	case 2:
		poolThreadedServer(listener)
	case 3:
		//añadir llamada a ./masterWorker
	default:
		panic("Error opción incorrecta")
	}	
	fmt.Println("Fin del servicio")

}
		
		
