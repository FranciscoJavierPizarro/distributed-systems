/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	06/10/2022
* Descripción de la estructura del código:
*	1. Importar paquetes
*	2. Funciones generales de todos los servidor
*	3. Servidor secuencial + funcion auxiliar
*	4. Servidor threaded + funciones auxiliares threade y pool treaded
*	5. Servidor pool threaded + funcion auxiliar
*	6. Main, desde el se lanza una de las 3 funciones de servidor aquí definidas
*/

//
//	IMPORTS
//

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

//
//	FUNCIONES GENERALES
//

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

//
//	SERVIDOR SECUENCIAL + FUNCION AUXILIAR
//

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

//
//	SERVIDOR MULTITHREAD + FUNCIÓN AUXILIAR MULTITHREAD Y POOLTHREAD
//

//función auxiliar para multithread y poolthread
//recibe una petición y la procesa y responde
//tiene añadidos el canal done y el booleano pool para la concurrencia
func ManageRequest_TH(decoder *gob.Decoder,encoder *gob.Encoder,
	conn net.Conn,done chan bool,pool bool){
	defer conn.Close()
	var request com.Request
	err := decoder.Decode(&request)
	checkError(err)
	if(request.Id == -1) {
		done <- true
	} else {
		if(!pool){//para el multithread se envia el canal done, para el pool no
		done <- false
		}
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
		go ManageRequest_TH(decoder,encoder,conn,done,first)
	}
}

//
//	SERVIDOR POOL THREAD + FUNCION AUXILIAR
//

//función que va acepando las conexiones y maneja la concurrencia
//utilizando el canal next y el canal done
func handleComms(start chan bool,listener net.Listener,done chan bool,next chan bool) {
	var b bool = true
	for {
		<-start
		conn, err := listener.Accept()
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		checkError(err)
		ManageRequest_TH(decoder,encoder,conn,done,b)
		next <- b
	}
}

func poolThreadedServer(listener net.Listener) {
	fmt.Println("Servidor pool threads")
	start := make(chan bool)
	next := make (chan bool)
	done :=make(chan bool)
	NTHREAD := 10
	var b bool = true
	for i:=0;i<NTHREAD;i++ {
		go handleComms(start,listener,done,next)
		start <- b
	}
	
	for {
		select{
		case <- next:
			start <- b
		case <- done:
			return
		}
	}
}

//
// MAIN DEL PROGRAMA
//

func main() {
	//Menu
	var option int
	fmt.Print("Opciones disponibles: \n\t0 servidor secuencial\n\t1 servidor multithread\n\t2 servidor con pool de threads\n\n")
	fmt.Scan(&option)
	fmt.Println("") 
	//Crear listener
	CONN_TYPE := "tcp"
	CONN_HOST := "127.0.0.1"
	CONN_PORT := "30000"
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	//Lanzar servidor
	switch option {
	case 0:
		secuentialServer(listener)
	case 1:
		threadedServer(listener)
	case 2:
		poolThreadedServer(listener)
	default:
		panic("Error opción incorrecta")
	}	
	fmt.Println("Fin del servicio")

}
		
		
