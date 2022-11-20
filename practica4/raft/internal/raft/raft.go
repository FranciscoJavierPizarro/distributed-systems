/*
* Autores:
*	Jorge Solán Morote NIA: 816259
*	Francisco Javier Pizarro NIA:  821259
* Fecha de última revisión:
*	09/11/2022
* Descripción del fichero:
*	Archivo principal de código, este se encarga de definir
*	el funcionamiento de un nodo raft, así
*	como las llamadas externas RPC que se pueden realizar
*	sobre el
* Descripción de la estructura del código:
*	1.Imports
*	2.Constantes generales
*	3.Estructuras de datos internas
*	4.Creación nodo raft
*	5.Funciones internas que llamadas por las RCP de API de los test
*	6.Estructuras de datos llamadas RPC API
*	7.Protocolos RPC llamadas API
*	8.Estructuras de datos llamadas RPC internos
*	9.Protocolos RPC llamada interna
*	10.Llamadas internas para invocar RPC interno
*	11.Flujo principal del nodo
*	12.Funciones auxiliares nodo candidato
*	13.Funciones auxiliares nodo lider
*	14.Llamadas RPC para test y control externo
*	15.Funciones auxilares generales
 */

package raft

//
// API
// ===
// nodoRaft = NuevoNodo(...)
//   Crear un nuevo servidor del grupo de elección.
//
// nodoRaft.Para()
//   Solicitar la parado de un servidor
//
// nodo.ObtenerEstado() (yo, mandato, esLider)
//   Solicitar a un nodo de elección por "yo", su mandato en curso,
//   y si piensa que es el msmo el lider
//
// nodoRaft.SometerOperacion(operacion interface()) (indice, mandato, esLider)

//====================================================================
//	IMPORTS
//====================================================================

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// "crypto/rand"
	"sync"
	"time"
	//"net/rpc"
	"math"
	"math/rand"
	"raft/internal/comun/rpctimeout"
)

//====================================================================
//	CONSTANTES
//====================================================================

const errorTime = heartTime * 4//ms
const electionTime = 2.5//seg
const reelectionTime = heartTime * 6//ms
const heartTime = 50//ms
const (
	// Constante para fijar valor entero no inicializado
	IntNOINICIALIZADO = -1

	//  false deshabilita por completo los logs de depuracion
	// Aseguraros de poner kEnableDebugLogs a false antes de la entrega
	kEnableDebugLogs = true

	// Poner a true para logear a stdout en lugar de a fichero
	kLogToStdout = false

	// Cambiar esto para salida de logs en un directorio diferente
	kLogOutputDir = "./logs_raft/"
)

//====================================================================
//	ESTRUCTURAS DE DATOS INTERNAS
//====================================================================

type TipoOperacion struct {
	Operacion string // La operaciones posibles son "leer" y "escribir"
	Clave     string
	Valor     string // en el caso de la lectura Valor = ""
}

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion TipoOperacion
}

type Log struct {
	Term      int
	Operacion TipoOperacion
}

type State struct {
	CurrentTerm int
	VotedFor    int
	Logs        []Log

	CommitIndex int
	LastApplied int

	NextIndex  []int
	MatchIndex []int

	Rol string
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
type NodoRaft struct {
	Mux sync.Mutex // Mutex para proteger acceso a estado compartido

	// Host:Port de todos los nodos (réplicas) Raft, en mismo orden
	Nodos        []rpctimeout.HostPort
	Yo           int // indice de este nodos en campo array "nodos"
	IdLider      int
	Logger       *log.Logger
	CurrentState State

	StillAlive   chan bool
	VoteRecivied chan bool
	Comprommised []chan bool

	Mapa 		 map[string]string
}

//====================================================================
//	CREACIÓN NODO RAFT
//====================================================================

// Creacion de un nuevo nodo de eleccion
//
// Tabla de <Direccion IP:puerto> de cada nodo incluido a si mismo.
//
// <Direccion IP:puerto> de este nodo esta en nodos[yo]
//
// Todos los arrays nodos[] de los nodos tienen el mismo orden

// canalAplicar es un canal donde, en la practica 5, se recogerán las
// operaciones a aplicar a la máquina de estados. Se puede asumir que
// este canal se consumira de forma continúa.
//
// NuevoNodo() debe devolver resultado rápido, por lo que se deberían
// poner en marcha Gorutinas para trabajos de larga duracion
func NuevoNodo(nodos []rpctimeout.HostPort, yo int,
	canalAplicarOperacion chan AplicaOperacion) *NodoRaft {
	nr := &NodoRaft{}
	nr.Nodos = nodos
	nr.Yo = yo
	nr.IdLider = -1
	nr.crearLogger(nodos, yo)
	nr.StillAlive = make(chan bool)
	nr.VoteRecivied = make(chan bool)
	nr.Comprommised = []chan bool{make(chan bool)}
	nr.Mapa = make(map[string]string)
	nr.createInitialState()

	go nr.run(canalAplicarOperacion)

	return nr
}

//Crea el logger del nodo
func (nr *NodoRaft) crearLogger(nodos []rpctimeout.HostPort, yo int) {
	if kEnableDebugLogs {
		nombreNodo := nodos[yo].Host() + "_" + nodos[yo].Port()
		logPrefix := fmt.Sprintf("%s", nombreNodo)

		fmt.Println("LogPrefix: ", logPrefix)

		if kLogToStdout {
			nr.Logger = log.New(os.Stdout, nombreNodo+" -->> ",
				log.Lmicroseconds|log.Lshortfile)
		} else {
			err := os.MkdirAll(kLogOutputDir, os.ModePerm)
			if err != nil {
				panic(err.Error())
			}
			logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt",
				kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err.Error())
			}
			nr.Logger = log.New(logOutputFile,
				logPrefix+" -> ", log.Lmicroseconds|log.Lshortfile)
		}
		nr.Logger.Println("logger initialized")
	} else {
		nr.Logger = log.New(ioutil.Discard, "", 0)
	}
}

//Crea el estado inicial del nodo
func (nr *NodoRaft) createInitialState() {
	nr.CurrentState = State{}

	nr.CurrentState.CurrentTerm = 0
	nr.CurrentState.VotedFor = -1
	nr.CurrentState.Logs = make([]Log, 0)
	nr.CurrentState.CommitIndex = 0
	nr.CurrentState.LastApplied = 0
	nr.CurrentState.MatchIndex = make([]int, len(nr.Nodos))
	nr.CurrentState.NextIndex = make([]int, len(nr.Nodos))
	for i := 0; i < len(nr.Nodos); i++ {
		nr.CurrentState.NextIndex[i] = 1
		nr.CurrentState.MatchIndex[i] = 0
	}

	nr.CurrentState.Rol = "Seguidor"
}

//====================================================================
//	FUNCIONES INTERNAS LLAMADAS POR LAS RPC DE LA API
//====================================================================

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
func (nr *NodoRaft) para() {
	go func() { time.Sleep(5 * time.Millisecond); os.Exit(0) }()
}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
// Primer valor devuelto es el indice de este  nodo Raft el el conjunto de nodos
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) obtenerEstado() (int, int, bool, int) {
	var yo int = nr.Yo
	var mandato int = nr.CurrentState.CurrentTerm
	var esLider bool = nr.Yo == nr.IdLider
	var idLider int = nr.IdLider

	nr.Logger.Printf("Estado actual %d %d %t %d", yo, mandato, esLider, idLider)

	return yo, mandato, esLider, idLider
}

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver en
// cuanto se consiga
//
// No hay garantia que esta operacion consiga comprometerse en una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) someterOperacion(operacion TipoOperacion) (int, int,
	bool, int, string) {
	indice := len(nr.CurrentState.Logs) + 1
	mandato := nr.CurrentState.CurrentTerm
	EsLider := nr.IdLider == nr.Yo
	idLider := nr.IdLider
	valorADevolver := operacion.Operacion

	if EsLider {
		nr.Comprommised = append(nr.Comprommised, make(chan bool))
		nr.CurrentState.Logs = append(nr.CurrentState.Logs, Log{nr.CurrentState.CurrentTerm, operacion})
		<-nr.Comprommised[indice-1]
	}

	return indice, mandato, EsLider, idLider, valorADevolver
}

// -----------------------------------------------------------------------
// LLAMADAS RPC al API
//
// Si no tenemos argumentos o respuesta estructura vacia (tamaño cero)

//====================================================================
//	ESTRUCTURAS DE DATOS LLAMADAS RPC API
//====================================================================

type Vacio struct{}

type EstadoParcial struct {
	Mandato int
	EsLider bool
	IdLider int
}

type EstadoRemoto struct {
	IdNodo int
	EstadoParcial
}

type ResultadoRemoto struct {
	ValorADevolver string
	IndiceRegistro int
	EstadoParcial
}

//====================================================================
//	PROTOCOLOS RPC LLAMADAS API
//====================================================================

func (nr *NodoRaft) ParaNodo(args Vacio, reply *Vacio) error {
	defer nr.para()
	return nil
}

func (nr *NodoRaft) DejarLider(args Vacio, reply *Vacio) error {
	nr.IdLider = -1
	nr.CurrentState.VotedFor = -1
	nr.CurrentState.Rol = "Seguidor"

	return nil
}

func (nr *NodoRaft) ObtenerEstadoNodo(args Vacio, reply *EstadoRemoto) error {
	nr.Logger.Println("Estado SOLICITADO")
	reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider = nr.obtenerEstado()
	nr.Logger.Printf("Estado actual %d %d %t %d", reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider)
	return nil
}

func (nr *NodoRaft) SometerOperacionRaft(operacion TipoOperacion,
	reply *ResultadoRemoto) error {
	nr.Logger.Println("Operación solicitada: " + operacion.Operacion)
	reply.IndiceRegistro, reply.Mandato, reply.EsLider, reply.IdLider, reply.ValorADevolver = nr.someterOperacion(operacion)
	nr.Logger.Println("Operación sometida")
	return nil
}

// -----------------------------------------------------------------------
// LLAMADAS RPC protocolo RAFT
//

//====================================================================
//	ESTRUCTURAS DE DATOS RPCS INTERNOS
//====================================================================

// Structura de ejemplo de argumentos de RPC PedirVoto.
type ArgsPeticionVoto struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// Structura de ejemplo de respuesta de RPC PedirVoto,
type RespuestaPeticionVoto struct {
	Term        int
	VoteGranted bool
}

type ArgAppendEntries struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

type Results struct {
	Term   int
	Sucess bool
}

//====================================================================
//	PROTOCOLOS RPC LLAMADA INTERNA
//====================================================================

// Metodo para RPC PedirVoto
func (nr *NodoRaft) PedirVoto(peticion *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) error {
	// Vuestro codigo aqui
	if nr.getState() == "Pausa" {
		for nr.getState() == "Pausa" {
			time.Sleep(heartTime * time.Millisecond)
		}
		return nil
	}
	nr.Logger.Println("RPC PedirVoto called from other replica")

	reply.Term = nr.CurrentState.CurrentTerm
	if peticion.Term < nr.CurrentState.CurrentTerm {
		reply.VoteGranted = false
	} else if (nr.CurrentState.VotedFor == -1 ||
		peticion.CandidateId == nr.CurrentState.VotedFor) &&
		(peticion.LastLogIndex >= nr.CurrentState.CommitIndex) {
		reply.VoteGranted = true
		nr.CurrentState.VotedFor = peticion.CandidateId
		nr.StillAlive <- true
	}
	return nil
}

// Metodo de tratamiento de llamadas RPC AppendEntries
func (nr *NodoRaft) AppendEntries(args *ArgAppendEntries,
	results *Results) error {
	if nr.getState() == "Pausa" {
		for nr.getState() == "Pausa" {
			time.Sleep(heartTime * time.Millisecond)
		}
		return nil
	}
	if len(args.Entries) == 0 {
		nr.Logger.Println("HeartPulse recivied")
	} else {
		nr.Logger.Println("RPC AppendEntries called from other replica")
	}
	if args.Term < nr.CurrentState.CurrentTerm {
		nr.Logger.Println("RPC error term")
		results.Sucess = false
		return nil
	}
	if len(nr.CurrentState.Logs) != 0 && args.PrevLogIndex > 0 &&
		nr.CurrentState.Logs[args.PrevLogIndex-1].Term != args.PrevLogTerm { //log donesnt contait an entry
		nr.Logger.Println("RPC error prevLog term")
		results.Sucess = false
		return nil
	}

	nr.auxAppendEntries(args)
	results.Sucess = true
	return nil
}

//Actualiza los valores locales del nodo si recibe una llamada appendentries
//correcta
func (nr *NodoRaft) auxAppendEntries(args *ArgAppendEntries) {
	if nr.CurrentState.Rol == "Candidato" {
		nr.CurrentState.Rol = "Seguidor"
	}
	nr.IdLider = args.LeaderId
	nr.CurrentState.VotedFor = -1

	for i := 0; i < len(args.Entries); i++ {
		nr.CurrentState.Logs = append(nr.CurrentState.Logs, args.Entries[i])
		//nr.Logger.Println(args.Entries[i].Operacion)
	}
	if len(args.Entries) > 0 {
		nr.CurrentState.CurrentTerm = args.Entries[len(args.Entries)-1].Term
	}
	if args.LeaderCommit > nr.CurrentState.CommitIndex {
		nr.CurrentState.CommitIndex = int(math.Min(float64(args.LeaderCommit), float64((len(nr.CurrentState.Logs) - 1))))
	}
	nr.StillAlive <- true
	nr.Logger.Println("RPC done")
}

//====================================================================
//	LLAMADAS LOCALES PARA INVOCAR RPC INTERNO
//====================================================================

func (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) bool {

	err := nr.Nodos[nodo].CallTimeout("NodoRaft.PedirVoto", args, &reply, time.Millisecond*errorTime)
	return err == nil
}

func (nr *NodoRaft) enviarAppendEntries(nodo int, args *ArgAppendEntries,
	reply *Results) bool {

	err := nr.Nodos[nodo].CallTimeout("NodoRaft.AppendEntries", args, &reply, time.Millisecond*errorTime)
	return err == nil
}

//====================================================================
//	FLUJO PRINCIPAL DEL NODO
//====================================================================

func (nr *NodoRaft) run(canalAplicarOperacion chan AplicaOperacion) {
	go nr.runCommonTasks(canalAplicarOperacion)
	go nr.runStateMachine(canalAplicarOperacion)
	//time.Sleep(7000 * time.Millisecond)//para el test1 el cual si ya se ha elegido lider
	//falla (es para compensar el tiempo de arranque de distributed process)
	for {

		estado := nr.getState()
		switch estado {
		case "Seguidor":
			nr.runSeguidor()
		case "Candidato":
			nr.runCandidato()
		case "Lider":
			nr.runLider()
		case "Pausa":
			nr.runPausa()
		}

	}
}

func (nr *NodoRaft) runCommonTasks(canalAplicarOperacion chan AplicaOperacion) {
	for {
		if nr.CurrentState.CommitIndex > nr.CurrentState.LastApplied {
			op := AplicaOperacion{
				nr.CurrentState.LastApplied,
				nr.CurrentState.Logs[nr.CurrentState.LastApplied].Operacion,
			}
			canalAplicarOperacion <- op
			nr.CurrentState.LastApplied += 1
		}
	}
}

func (nr *NodoRaft) runStateMachine(canalAplicarOperacion chan AplicaOperacion) {
	for {
		tipoop := <- canalAplicarOperacion
		op := tipoop.Operacion
		if(op.Operacion == "lectura") {
			nr.Logger.Println("Operacion de lectura del valor: ",nr.Mapa[op.Clave])
		} else if(op.Operacion == "escritura") {
			nr.Mapa[op.Clave] = op.Valor
			nr.Logger.Println("Operación de escritura con clave: ", op.Clave, " del valor: ", op.Valor)
		}
	}
}

func (nr *NodoRaft) runSeguidor() {
	nr.Logger.Println("Now i am a follower")
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	timeToChangeMode := time.Millisecond * time.Duration(errorTime+(r1.Intn(errorTime * 3)))
	ticker := time.NewTicker(timeToChangeMode)
	for nr.getState() == "Seguidor" {
		select {
		case <-nr.StillAlive:
			//Reiniciar contador
			ticker.Reset(timeToChangeMode)
		case <-ticker.C:
			//Convertirse a candidato
			ticker.Stop()
			if nr.getState() != "Pausa" { //bug por eso if
				nr.CurrentState.Rol = "Candidato"
			}
		}
	}
}

func (nr *NodoRaft) runCandidato() {
	nr.Logger.Println("Now i am a candidate")
	nr.CurrentState.CurrentTerm++
	ticker := time.NewTicker(time.Millisecond * reelectionTime)
	nr.CurrentState.VotedFor = nr.Yo
	for nr.getState() == "Candidato" {
		votes := 1 //se vota a si mismo
		notTimeout := true
		go nr.lanzarPeticionesVotos()
		for !(votes >= ((len(nr.Nodos) / 2) + 1)) && notTimeout { //proceso elección
			select {
			case <-nr.VoteRecivied: //Sumar voto y comprobar mayoría
				nr.Logger.Println("Voto procesado")
				votes += 1
				if votes >= ((len(nr.Nodos)/2)+1) && nr.getState() != "Pausa" {
					nr.Logger.Println("Voy a ser lider")
					nr.Mux.Lock()
					nr.IdLider = nr.Yo
					nr.CurrentState.Rol = "Lider"
					nr.Mux.Unlock()
				}
			case <-ticker.C: //Reiniciar elección
				notTimeout = false
				nr.CurrentState.CurrentTerm++
			}
		}
	}
}

func (nr *NodoRaft) runLider() {
	nr.Logger.Println("Now i am a leader")
	nr.lanzarLatidos()
	// s1 := rand.NewSource(time.Now().UnixNano())
	// r1 := rand.New(s1)
	heartPulse := time.Millisecond * time.Duration(heartTime)
	ticker := time.NewTicker(heartPulse)

	for nr.getState() == "Lider" {
		select {
		case <-ticker.C:
			nr.lanzarLatidos() //Enviar latidos/append
		}
	}
	ticker.Stop()
}

func (nr *NodoRaft) runPausa() {
	nr.Logger.Println("Now i am in pause")
	for nr.getState() == "Pausa" {
		time.Sleep(heartTime * time.Millisecond)
	}
}

//====================================================================
//	FUNCIONES AUXILIARES NODO CANDIDATO
//====================================================================

func (nr *NodoRaft) lanzarPeticionesVotos() {
	args := ArgsPeticionVoto{
		nr.CurrentState.CurrentTerm,
		nr.Yo,
		nr.CurrentState.LastApplied,
		nr.CurrentState.CurrentTerm,
	}
	var reply RespuestaPeticionVoto
	for i := 0; i < len(nr.Nodos); i++ {
		if i != nr.Yo {
			go func(i int, args ArgsPeticionVoto, reply RespuestaPeticionVoto) {
				if nr.enviarPeticionVoto(i, &args, &reply) {
					nr.Logger.Println("Voto solicitado")
					if reply.VoteGranted {
						nr.Logger.Println("Voto recibido")
						nr.VoteRecivied <- true
					} else {
						nr.CurrentState.CurrentTerm = reply.Term
					}
				} else {
					nr.Logger.Println("Solicitud de voto fallido")
				}
			}(i, args, reply)
		}
	}
}

//====================================================================
//	FUNCIONES AUXILIARES NODO LIDER
//====================================================================

func (nr *NodoRaft) lanzarLatidos() {
	var reply Results
	for i := 0; i < len(nr.Nodos); i++ {
		if i != nr.Yo {
			args := ArgAppendEntries{
				nr.CurrentState.CurrentTerm,
				nr.Yo, 0, 0, make([]Log, 0),
				nr.CurrentState.CommitIndex,
			}
			if nr.CurrentState.NextIndex[i] <= len(nr.CurrentState.Logs) {
				args.Entries = nr.CurrentState.Logs[nr.CurrentState.MatchIndex[i]:len(nr.CurrentState.Logs)]
				args.PrevLogIndex = nr.CurrentState.MatchIndex[i]
				args.PrevLogTerm = nr.CurrentState.Logs[args.PrevLogIndex].Term
			}
			go nr.enviarRPC(i, args, reply)
		}
	}
	nr.actualizarEntradasComprometidas()
}

func (nr *NodoRaft) enviarRPC(i int, args ArgAppendEntries, reply Results) {
	if nr.enviarAppendEntries(i, &args, &reply) {
		if len(args.Entries) == 0 {
			nr.Logger.Println("Latido enviado")
		} else {
			nr.Logger.Println("Append enviado")

			if reply.Sucess {
				nr.CurrentState.MatchIndex[i] += len(args.Entries)
				nr.CurrentState.NextIndex[i] += len(args.Entries)
				nr.Logger.Println("Append correcto")
			}
		}
	} else {
		if len(args.Entries) == 0 {
			nr.Logger.Println("Latido fallido")
		} else {
			nr.Logger.Println("RPC fallido")
		}
	}
}

func (nr *NodoRaft) actualizarEntradasComprometidas() {
	contador := 1
	for i := 0; i < len(nr.Nodos); i++ {
		if i != nr.Yo && nr.CurrentState.MatchIndex[i] == len(nr.CurrentState.Logs) {
			contador += 1
		}
	}
	if contador >= ((len(nr.Nodos) / 2) + 1) {
		for i := nr.CurrentState.CommitIndex; i < len(nr.CurrentState.Logs); i++ {
			nr.Comprommised[i] <- true
		}

		nr.CurrentState.CommitIndex = len(nr.CurrentState.Logs)
		nr.Logger.Printf("Comprometido hasta log %d\n", nr.CurrentState.CommitIndex)
	}
}

//====================================================================
//	LLAMADAS RPC PARA TEST Y CONTROL EXTERNO
//====================================================================

func (nr *NodoRaft) ObtenerRegistro(args int,
	reply *string) error {
	nr.printLogs()
	if nr.getState() == "Pausa" {
		for nr.getState() == "Pausa" {
			time.Sleep(heartTime * time.Millisecond)
		}
		return nil
	}
	*reply = nr.CurrentState.Logs[args].Operacion.Operacion
	nr.Logger.Println("Log enviado:" + nr.CurrentState.Logs[args].Operacion.Operacion)
	return nil
}

func (nr *NodoRaft) PonerPausa(args Vacio, reply *Vacio) error {
	nr.Logger.Println("PAUSED")
	nr.CurrentState.Rol = "Pausa"
	return nil
}

func (nr *NodoRaft) QuitarPausa(args Vacio, reply *Vacio) error {
	nr.Logger.Println("CONTINUE")
	nr.CurrentState.Rol = "Seguidor"
	return nil
}

func (nr *NodoRaft) ObtenerCompromiso(args Vacio,
	reply *[]int) error {
	*reply = []int{
		len(nr.CurrentState.Logs),
		nr.CurrentState.CommitIndex,
	}
	nr.Logger.Println("Situacion de compromiso enviada")
	return nil
}

func (nr *NodoRaft) getState() string {
	return nr.CurrentState.Rol
}

//====================================================================
//	FUNCIONES AXUILIARES GENERALES
//====================================================================

func (nr *NodoRaft) printLogs() {
	nr.Logger.Println("LOGS:")
	for i := 0; i < len(nr.CurrentState.Logs); i++ {
		nr.Logger.Println(nr.CurrentState.Logs[i].Operacion.Clave)
	}
}
