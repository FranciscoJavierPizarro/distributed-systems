/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
*/
package ra

//====================================================================
//	IMPORTS
//====================================================================

import (
    "ms"
    "sync"
    "github.com/DistributedClocks/GoVector/govec"
    "strconv"
    "log"
    "fmt"
    "os"
)

//====================================================================
//	PAQUETES DE RED + ESTRUCTURA RA
//====================================================================

type Request struct{
    Clock   int
    Pid     int
    WriteOp bool
}

//Paquete para ordenar a un proceso que actualice su fichero
type TextToUpdate struct{
    Pid  int
    Text string
}

//Paquete de respuesta a la solicitud de actualizar un fichero
type AlreadyUpdated struct{}

type Reply struct{}

type RASharedDB struct {
    OurSeqNum   int
    HigSeqNum   int
    OutRepCnt   int
    ReqCS       bool
    RepDefd     []int
    ms          *ms.MessageSystem
    syncronized chan bool
    done        chan bool
    confirm     chan bool
    Mutex       sync.Mutex // mutex para proteger concurrencia sobre las variables

    OwnPid      int
    nProc       int
    chRep       chan Reply//para contar n respuestas
    writeOp     bool
    logger      *govec.GoLog//GoVector
}

//====================================================================
//	FUNCIONES DE COMUNICACIÓN
//====================================================================

//función auxiliar que recibe las respuestas(se usa para contar)
func (ra *RASharedDB) ReceiveRep() {
	<-ra.chRep
}

//función auxiliar que recibe las confirmaciones de actualización
//de un fichero (se usa para contar)
func (ra *RASharedDB) ReceiveConfirm() {
	<-ra.confirm
}
//función auxiliar que recibe el mensaje de sincronización de la barrera
//(se usa para contar)
func (ra *RASharedDB) ReceiveSync() {
	<-ra.syncronized
}

//función que va recibiendo las peticiones y las va atendiendo
func (ra *RASharedDB) ReceiveMsg() {
	var request Request
	for{
		switch msg := ra.ms.Receive().(type) {
            case Reply:
                ra.chRep <- msg
            case AlreadyUpdated:
                ra.confirm <- true
            case TextToUpdate:
                WriteF(strconv.Itoa(ra.OwnPid)+".txt",msg.Text)
                ra.ms.Send(msg.Pid, AlreadyUpdated{})
            case ms.SyncWait:
                ra.syncronized <- true
            case []byte:
                ra.logger.UnpackReceive("Received request", msg,
                &request, govec.GetDefaultLogOptions())
                // Procesa la request
                go ra.processRequest(request)
            default:
                log.Printf("AYUDA")
		}
	}
}

//función que envia señal de sincronización al main
func (ra *RASharedDB) SendSignal() {
	ra.ms.Send(ra.nProc + 1, ms.SyncSignal{})
}

//función que envia solicitud de actualizar fichero
//a todos los procesos
func (ra *RASharedDB) UpdateFile(text string) {
    textToUpdate := TextToUpdate{
    Text:text,
    Pid:ra.OwnPid,
    }
    for i := 0; i < ra.nProc; i++ {
		if (i != (ra.OwnPid - 1)) {ra.ms.Send(i + 1, textToUpdate)}
	}
}

//función que espera a recibir las confirmaciones de modificación
func (ra *RASharedDB) WaitConfirms() {
    for i := 0; i < ra.nProc; i++ {
		if (i != (ra.OwnPid - 1)) {ra.ReceiveConfirm()}
	}
}

//====================================================================
//	FUNCIONES DE CREACIÓN DE RA Y DE CONTROL DE SC
//====================================================================

func New(pid int, usersFile string, nProc int) (*RASharedDB) {
    messageTypes := []ms.Message{Request{}, Reply{}, ms.SyncSignal{}, ms.SyncWait{}, []byte{},TextToUpdate{},AlreadyUpdated{}}
    msgs := ms.New(pid, usersFile, messageTypes)
    ra := RASharedDB{
        OurSeqNum:  0, 
        HigSeqNum: 0, 
        OutRepCnt:  0, 
        ReqCS:      false, 
        RepDefd:     make([]int, nProc),
        ms:         &msgs,  
        done:       make(chan bool),
        syncronized:make(chan bool),
        confirm:make(chan bool),  
        Mutex:      sync.Mutex{},    
        OwnPid:     pid,
        nProc:      nProc,
        chRep:      make(chan Reply, nProc),
        writeOp:    false,
        logger: govec.InitGoVector(strconv.Itoa(pid),
		"LogFile"+strconv.Itoa(pid),	govec.GetDefaultConfig()),
        //Para ahorrarnos la matriz de exclusión en la que 
        //la única operación compatible con otra es lectura+lectura
        //simplemente usamos el valor writeOp que cuando es false es
        //lectura, realizando una OR sobre dicho operando obtenemos
        //la matriz de exclusión
    }
    return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol(op bool){
    // TODO completar
    ra.Mutex.Lock()
	ra.ReqCS = true
	ra.OurSeqNum = ra.HigSeqNum + 1
	ra.writeOp = op
	ra.Mutex.Unlock()
    ra.OutRepCnt = ra.nProc - 1

    request := Request{
        Clock:  ra.OurSeqNum,
        Pid:    ra.OwnPid,
        WriteOp:ra.writeOp,
    }
    goVecMessage := ra.logger.PrepareSend("Send requests",
    request, govec.GetDefaultLogOptions())

    for i := 0; i < ra.nProc; i++ {
		if (i != (ra.OwnPid - 1)) {ra.ms.Send(i + 1, goVecMessage)}
	}

    for i := 0; i < ra.OutRepCnt; i++ {
		ra.ReceiveRep()
	}
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol(){
    // TODO completar
    ra.logger.LogLocalEvent("Salgo de la SC", govec.GetDefaultLogOptions())
    ra.Mutex.Lock()
	ra.ReqCS = false
	ra.Mutex.Unlock()
    reply := Reply{}
    for i := 0; i < ra.nProc; i++ {
		//Si tenemos marcado un nodo como pendiente de responder,
        //le respondemos y marcamos que ya le hemos respondido
		if ra.RepDefd[i] == 1 {
			ra.RepDefd[i] = 0
			ra.ms.Send(i + 1, reply)
		}
	}
}

func (ra *RASharedDB) Stop(){
    ra.ms.Stop()
    ra.done <- true
}


//====================================================================
//	FUNCIONES AUXILIARES
//====================================================================

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

// Procesa una request de acceso a SC y realiza las acciones necesarias
func (ra *RASharedDB) processRequest(request Request) {
    ra.Mutex.Lock()
    if ra.HigSeqNum < request.Clock {ra.HigSeqNum = request.Clock}
    exclusion := ra.writeOp || request.WriteOp
    priority := (((request.Clock == ra.OurSeqNum) && (request.Pid < ra.OwnPid))||
    (request.Clock > ra.OurSeqNum))
    if (ra.ReqCS && priority && exclusion) {ra.RepDefd[request.Pid - 1] = 1
    } else {
        ra.ms.Send(request.Pid, Reply{})
        ra.logger.LogLocalEvent("Respondo a solicitud SC", govec.GetDefaultLogOptions())
    }
    ra.Mutex.Unlock()
}