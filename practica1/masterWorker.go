/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	06/10/2022
* Descripción de la estructura del código:
*	1. Importar paquetes
*	2. Funciones generales de servidor y estructuras de datos
*	3. Funciones particulares core del servidor
*	4. Funciones de comunicación del servidor con los clientes
*	5. Funciones de comunicación del servidor con los workers
*	6. Funciones empleadas por el servidor para el apagado y 
*	   encendido de los workers
*	7. Función de worker y funciones auxiliares
*	8. Main, desde el se lanza un proceso worker o un proceso masterWorker
*	   para lanzar un worker se tiene que usar el flag de incovación --worker
*	   en caso de no usar este flag se lanza un proceso masterWorker
*
* Notas de despligue:
*	- Modificar las ips de escucha del listener de workers y las ips
*	  de conexión de los workers, modificar listado de ips.txt y variable
*	  user del launchWorkers.sh
*	- Copiar el ejecutable masterWorker a la carpeta HOME de las máquinas
*	  que van a ser workers.
*
* Pendiente de implementar por falta de tiempo:
*	- Control de vida de los procesos worker
*	- Uso de threads en los workers para procesar varias tareas de forma
*	  simultanea
*/

//
// IMPORTAR PAQUETES
//

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

//
// FUNCIONES GENERALES DEL SERVIDOR Y ESTRUCTURAS DE DATOS
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

//
// FUNCIONES PARTICULARES CORE DEL SERVIDOR
//

func workersServer(listener net.Listener) {
	fmt.Println("Servidor masterWorker")
	tasksChan := make(chan Task)
	
	go listenClients(listener,tasksChan)
	serverCore(tasksChan)

}

//Crea la conexión con los workers y los lanza, devuelve las conexiones al core
func launchWorkers(workersChan chan net.Conn) {
	//Conexión
	CONN_TYPE := "tcp"
	CONN_HOST := "127.0.0.1"
	CONN_PORT := "30001"
	workerListener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	//Encendemos los workers
	go listenWorkers(workerListener,workersChan)
	go turnOnWorkers()
}

//FUNCIÓN CORE DEL SERVIDOR
//escucha los nuevos workers, las nuevas tareas y las respuestas de los workers a las tareas
//cada uno en su corespondiente canal, para cada respuesta lanza la función de control correspondiente
func serverCore(tasksChan chan Task) {
	//Canales para la recepción de información por parte de las funciones de escucha
	workersChan := make(chan net.Conn)
	replyChan := make(chan WorkerReply)
	//Colas internas para el control de los workers, de las conexiones de los clientes
	//y de las tareas pendientes
	workers := make([]net.Conn,0)
	clients := make([]net.Conn,1000)
	acummulatedTasks := make([]Task,0)
	readyToWork := make([]bool,0)
	//Variables de control del fin
	var readyToEnd bool = false
	end := []bool{false}
	launchWorkers(workersChan)
	for !end[0] {
		select{
		case conn :=<-workersChan:
			workers, readyToWork, acummulatedTasks = manageNewWorker(conn,workers,acummulatedTasks,readyToWork,clients,replyChan)
		case task :=<-tasksChan:
			//acummulated se devuelve porque sobre el se hace un append y dicho cambio solo afecta a la copia
			//interna del slice de la función
			acummulatedTasks = manageTask(task,workers,clients,readyToWork,acummulatedTasks, replyChan,readyToEnd, end)
		case reply:=<-replyChan:
			acummulatedTasks = manageReply(reply, clients, workers, readyToWork, acummulatedTasks, replyChan, readyToEnd, end)
		}
	}
	turnOffWorkers(workers)
}


//Recibe la conexión de un nuevo worker, en caso de que haya tareas acumuladas le asigna una y lo 
//marca como ocupado, en caso de que no haya tareas lo marca como libre
//en ambos casos añade su conexión al array workers
func manageNewWorker(conn net.Conn,workers []net.Conn, acummulatedTasks []Task,
readyToWork []bool,clients []net.Conn, replyChan chan WorkerReply) ([]net.Conn, []bool, []Task) {
	fmt.Println("Mensaje recibido en workersChan")
	workers = append(workers, conn)
	
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
	return workers, readyToWork, acummulatedTasks
}


//se lanza cuando el master recibe un, en nueva tarea por parte de un cliente
//comprueba si es la tarea final, en caso de ser esta marca el posible fin
//y comprueba si quedan workers trabajando o tareas pendientes
//en caso de que ambas respuestas sean no, finaliza el servicio
//si es una tarea normal trata de asignarla a un worker libre(y escuchar su respuesta)
//si no hay workers libres la añade a la cola de tareas pendientes
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

//Cuando un worker envia una respuesta al master
//esta función gestiona el envio de la respuesta al cliente
//y marca el worker como libre si no hay tareas pendientes
//si hay tareas pendientes le asigna una al worker
//si ya se ha recibido la señal de fin comprueba
//si quedan tareas pendientes y si quedan workers trabajando
//si la en ambos casos la respuesta es no, marca el fin del servidor
//devuelve la lista de tareas pendientes
func manageReply(reply WorkerReply, clients []net.Conn, workers []net.Conn, readyToWork []bool,
acummulatedTasks []Task, replyChan chan WorkerReply, readyToEnd bool, end []bool) ([]Task){
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

//
// FUNCIONES DEL SERVIDOR PARA LA COMUNICACIÓN CON LOS CLIENTES
//

//Escucha clientes y va mandando las tareas con sus respectivos clientes
//al core del servidor
func listenClients(listener net.Listener,tasksChan chan Task) {
	//Escucha clientes hasta que llega el último paquete con id -1
	var b bool = true
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
}

//Dada una respuesta la envia al cliente y cierra la conexión
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


//
// FUNCIONES DEL SERVIDOR PARA LA COMUNICACIÓN CON LOS WORKERS
//


//Dada una conexión a un worker le manda la tarea
func assignTask(conn net.Conn, wreq WReq) {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(wreq)
	checkError(err)
	fmt.Println("Tarea asignada")
}

//Dada una conexión a un worker espera su respuesta y la devuelve al core del server
func listenAnwsers(workerConn net.Conn, replyChan chan WorkerReply) {
	decoder := gob.NewDecoder(workerConn)
	var reply WorkerReply
	err := decoder.Decode(&reply)
	checkError(err)
	fmt.Println("Respuesta recibida")
	replyChan <- reply
}

//
// FUNCIONES DEL SERVIDOR PARA APAGAR Y
// ENCENDER LOS WORKERS
//

//Enciende los workers
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

//Crea las conexiones con los workers encendidos y se las pasa al core del server
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

//Apaga todos los workers
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

//
// FUNCIÓN DEL WORKER Y FUNCIONES AUXILIARES
//

//Creación de la conexión al server
func createConn() (net.Conn){
	fmt.Println("Worker lanzado")
	endpoint := "127.0.0.1:30001"
	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	return conn
}

//Procesa una workerRequest, en caso de ser necesario responde con 
//el resultado usando el encoder, devuelve siempre un booleano
//indicando si espera o no mas tareas en base a si ha recibido o no
//el paquete de fin
func processWorkerRequest(request WReq,encoder *gob.Encoder) (bool) {
	fmt.Println(request.Id,request.WorkerId)
	if(request.WorkerId == -1) {
		fmt.Println("FinWorker")
		return false
	} else {
		reply := WorkerReply{request.Id,FindPrimes(request.Interval),request.WorkerId}
		errRep := encoder.Encode(reply)
		checkError(errRep)
		fmt.Println("Tarea resuelta")
		return true
	}
}

//Protocolo a ejecutar por los workers
func workerProtocol() {
	conn := createConn()
	defer conn.Close()
	//Bucle principal, se recibe una tarea y se responde la respuesta
	var request WReq
	var moreTasks bool = true
	for moreTasks {
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		err := decoder.Decode(&request)
		fmt.Println("Tarea asignada")
		checkError(err)
		moreTasks = processWorkerRequest(request, encoder)
	}
}


//
// MAIN
//

func main() {
	worker := flag.Bool("worker",false,"This instance of the program is a worker in ther master worker architecture")
	flag.Parse()

	if (*worker) {
		//Es un proceso worker
		workerProtocol()
	} else {
		//Es un proceso servidor
		//Creamos listener
		CONN_TYPE := "tcp"
		CONN_HOST := "127.0.0.1"
		CONN_PORT := "30000"
		listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
		checkError(err)
		//Lanzamos función servidor
		workersServer(listener)
		fmt.Println("Fin del servicio")
	}
}