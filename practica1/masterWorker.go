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
	"os/exec"
	//"io"
	//"com"
	"practica1/com"
	"flag"
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

type Task struct {
    Id int
	Conn net.Conn
	Interval com.TPInterval
}

type Anwser struct {
    Id int
	Conn net.Conn
	Primes []int
}

func workersServer(listener net.Listener) {
	fmt.Println("Servidor masterWorker")
	tasksChan := make(chan Task)
	
	go listenClients(listener,tasksChan)
	manageWorkers(tasksChan)

}

func listenClients(listener net.Listener,tasksChan chan Task) {
	var b bool = true
	//Escucha clientes hasta que llega el último paquete
	for b{
		conn, err := listener.Accept()
		checkError(err)
		decoder := gob.NewDecoder(conn)
		var request com.Request
		errd := decoder.Decode(&request)
		checkError(errd)
		if (request.Id == -1) {
			b = false
		}
		task := Task{request.Id,conn,request.Interval}
		tasksChan <- task
		fmt.Println("Tarea añadida por cliente")
	}
	// cuando resultsChan devuelve Anwser{-1,-1,[0]) se acaba
}


func anwserClient(anwser Anwser) {
	//Extrae los resultados y se encarga de enviarlos y de cerrar las conexiones
	id := anwser.Id
	conn := anwser.Conn
	defer conn.Close()
	primes := anwser.Primes
	reply := com.Reply{id,primes}
	encoder := gob.NewEncoder(conn)
	errRep := encoder.Encode(reply)
	checkError(errRep)
	fmt.Println("Resultado enviado")
}

type WReq struct {
    Id int
    Interval com.TPInterval
	WorkerId int
}

type WorkerReply struct {
    Id int
    Primes []int
	WorkerId int
}

func manageWorkers(tasksChan chan Task) {
	//Recibe tareas por tasksChan, las asigna a los workers disponibles o las guarda
	//hasta tener workers disponibles, devuelve los resultados por resultsChan
	//cuando un worker responde con una anwser de id -1 apaga a todos y finaliza.
	
	//Constantes y conexión
	// N_WORKERS:= 1
	CONN_TYPE := "tcp"
	CONN_HOST := "127.0.0.1"
	CONN_PORT := "30001"
	workerListener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)

	//Listas dinámicas para almacenar workers, comunicarse con ellos, almacenar las tareas
	//y el estado de cada worker
	workersChan := make(chan net.Conn)
	replyChan := make(chan WorkerReply)
	workers := make([]net.Conn,0)
	clients := make([]net.Conn,1000)
	acummulatedTasks := make([]Task,0)
	readyToWork := make([]bool,0)
	var readyToEnd bool = false
	end := []bool{false}
	//Encendemos los workers
	go listenWorkers(workerListener,workersChan)
	go turnOnWorkers()
	
	for !end[0] {
		select{
		//Recibe los distintos workers y apunta sus canales y los marca como dispibles
		case conn :=<-workersChan:
			fmt.Println("Mensaje recibido en workersChan")
			workers = append(workers, conn)
			
			// CONSIDERAR WORKER ENTRA Y YA HAY TAREAS PENTIENTES
			if (len(acummulatedTasks) == 0) {
				readyToWork = append(readyToWork, true)
			}else {
				readyToWork = append(readyToWork, false)
				task := acummulatedTasks[0]
				acummulatedTasks = acummulatedTasks[1:]
				clients[task.Id] = task.Conn
				request := WReq{task.Id, task.Interval, len(readyToWork) - 1}
				assignTask(conn,request)
				go listenAnwsers(conn,replyChan)
				
			}
		//Recibe una tarea, se la asigna a un worker libre y lo marca como ocupado
		//Si no hya workers libres, la marca como tarea pendiente
		case task :=<-tasksChan:
			//acummulated se devuelve porque sobre el se hace un append y dicho cambio solo afecta a la copia
			//interna del slice de la función
			acummulatedTasks = manageTask(task,workers,clients,readyToWork,acummulatedTasks, replyChan,readyToEnd, end)

		//Recibe una respuesta y la procesa, en caso de que el id de la tarea -1 ya
		// haya aparecido y todos los workers esten libres los apaga
		case reply:=<-replyChan:
			acummulatedTasks = manageReply(reply, clients, workers, readyToWork, acummulatedTasks, replyChan, readyToEnd, end)
		}
			
	}
	turnOffWorkers(workers)
}

func manageTask(task Task, workers []net.Conn, clients []net.Conn, readyToWork []bool, acummulatedTasks []Task, replyChan chan WorkerReply, readyToEnd bool, end []bool) ([]Task){
	fmt.Println("Mensaje recibido en tasksChan")
	fmt.Println(task.Id)
	var i int = 0
	if (task.Id == -1) {
		readyToEnd = true
		if (len(acummulatedTasks) == 0) {
			var tmp bool = true
			for i:=0;i<len(workers);i++ {
				tmp = tmp && readyToWork[i]
			}
			end[0] = tmp
			if (end[0]) {
				return acummulatedTasks
			}
		}
	} else {
		for (i < len(workers)) {
			if (readyToWork[i] == true) {
				readyToWork[i] = false
				worker := workers[i]
				clients[task.Id] = task.Conn
				request := WReq{task.Id, task.Interval, i}
				assignTask(worker,request)
				go listenAnwsers(worker,replyChan)
				break
			}
			i++
		}
		if (i == len(workers)) {
			acummulatedTasks = append(acummulatedTasks, task)
			fmt.Println("Tarea encolada", len(acummulatedTasks))
		}
	}
	return acummulatedTasks
}

func manageReply(reply WorkerReply, clients []net.Conn, workers []net.Conn, readyToWork []bool, acummulatedTasks []Task,
replyChan chan WorkerReply, readyToEnd bool, end []bool) ([]Task){
	fmt.Println("Mensaje recibido en replyChan")
	anwser := Anwser{reply.Id,clients[reply.Id],reply.Primes}
	anwserClient(anwser)
	if (len(acummulatedTasks) == 0) {
		readyToWork[reply.WorkerId] = true
		if (readyToEnd) {
			tmp := true
			for i:=0;i<len(workers);i++ {
				tmp = tmp && readyToWork[i]
			}
			end[0] = tmp
		}
	} else {
		worker := workers[reply.WorkerId]
		task := acummulatedTasks[0]
		acummulatedTasks = acummulatedTasks[1:]
		clients[task.Id] = task.Conn
		request := WReq{task.Id, task.Interval, reply.WorkerId}
		assignTask(worker,request)
		go listenAnwsers(worker,replyChan)
		
	}
	return acummulatedTasks
}

func turnOnWorkers() {
	//LLama a ./launchWorkers.sh
	cmd := exec.Command("./launchWorkers.sh")
    err := cmd.Run()
    if err != nil {
		panic(err)
    }
	fmt.Println("Workers activados")
	return
}

func listenWorkers(listener net.Listener, workersChan chan net.Conn) {
	//Crea la comunicación con todos los workers
	for {
		conn, err := listener.Accept()
		checkError(err)
		fmt.Println("Nuevo worker")
		workersChan <- conn
	}
	return
}

func assignTask(conn net.Conn, wreq WReq) {
	//Dada una conexión a un worker le manda la tarea
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(wreq)
	checkError(err)
	fmt.Println("Tarea asignada")
}

func listenAnwsers(workerConn net.Conn, replyChan chan WorkerReply) {
	//Dada una conexión a un worker espera su respuesta y la devuelve por anwserChan
	decoder := gob.NewDecoder(workerConn)
	var reply WorkerReply
	err := decoder.Decode(&reply)
	checkError(err)
	fmt.Println("Respuesta recibida")
	replyChan <- reply
}

func turnOffWorkers(conns []net.Conn) {
	//Cierra la comunicación con todos los workers
	wreq := WReq{-1,com.TPInterval{0,0},-1}
	for i := 0;i < len(conns);i++ {
		conn := conns[i]
		assignTask(conn,wreq)
		conn.Close()
	}
	fmt.Println("Workers apagados")
	return
}

func workerProtocol() {
	//Creación de la conexión al server
	fmt.Println("Worker lanzado")
	endpoint := "127.0.0.1:30001"
	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	defer conn.Close()
	
	//Bucle principal, se recibe una tarea y se responde la respuesta
	
	var request WReq
	var moreTasks bool = true
	for moreTasks {
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		fmt.Println("Tarea asignada")
		err := decoder.Decode(&request)
		checkError(err)
		fmt.Println(request.Id,request.WorkerId)
		if(request.WorkerId == -1) {
			moreTasks = false
			fmt.Println("FinWorker")
		} else {
			reply := WorkerReply{request.Id,FindPrimes(request.Interval),request.WorkerId}
			errRep := encoder.Encode(reply)
			checkError(errRep)
			fmt.Println("Tarea resuelta")
		}
	}
}


func main() {
	worker := flag.Bool("worker",false,"This instance of the program is a worker in ther master worker architecture")
	flag.Parse()

	if (*worker) {
		//Es un proceso worker
		workerProtocol()
	} else {
		//Es un proceso servidor
		CONN_TYPE := "tcp"
		CONN_HOST := "127.0.0.1"
		CONN_PORT := "30000"
		listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		checkError(err)
		workersServer(listener)
		fmt.Println("Fin del servicio")
	}
}