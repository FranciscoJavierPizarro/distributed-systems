/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
*/
package ra

import (
    "ms"
    "sync"
    "github.com/DistributedClocks/GoVector/govec"
    "strconv"
    "log"
)

type Request struct{
    Clock   int
    Pid     int
    WriteOp bool
}

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
    //chrep       chan bool
    Mutex       sync.Mutex // mutex para proteger concurrencia sobre las variables

    // TODO: completar
    OwnPid      int
    nProc       int
    chRep       chan Reply//para contar n respuestas
    writeOp     bool
    logger      *govec.GoLog//GoVector
}


func New(pid int, usersFile string, nProc int) (*RASharedDB) {
    messageTypes := []ms.Message{Request{}, Reply{}, ms.SyncSignal{}, ms.SyncWait{}, []byte{}}
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

//función auxiliar que recibe las respuestas
func (ra *RASharedDB) ReceiveRep() {
	<-ra.chRep
}

func (ra *RASharedDB) ReceiveSync() {
	<-ra.syncronized
}

func (ra *RASharedDB) processRequest(request Request) {
    ra.Mutex.Lock()
    if ra.HigSeqNum < request.Clock {ra.HigSeqNum = request.Clock}
    exclusion := ra.writeOp || request.WriteOp
    priority := (((request.Clock == ra.OurSeqNum) && (request.Pid < ra.OwnPid))||
    (request.Clock > ra.OurSeqNum))
    if (ra.ReqCS && priority && exclusion) {ra.RepDefd[request.Pid - 1] = 1
    } else {ra.ms.Send(request.Pid, Reply{})}
    ra.Mutex.Unlock()
}

//función que va recibiendo las peticiones y las va atendiendo
func (ra *RASharedDB) ReceiveMsg() {
	var request Request
	for{
		switch msg := ra.ms.Receive().(type) {
            case Reply:
                ra.chRep <- msg
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

//función que va recibiendo las peticiones y las va atendiendo
func (ra *RASharedDB) SendSignal() {
	ra.ms.Send(ra.nProc + 1, ms.SyncSignal{})
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol(){
    // TODO completar
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
