package main

import (
	//"errors"
	"fmt"
	//"log"
	//"os"
	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"strconv"
	"time"
)
const (
	errorTime = 40//ms
	startTime = 7000//ms
	compromiseTime = 2000//ms
)

// Somete una operacion al nodo
func someterOperacion(
	indiceNodo int, iteracion int, nodosRaft []rpctimeout.HostPort) bool {
	operacion := raft.TipoOperacion{"escritura", "iteracion" + strconv.Itoa(iteracion), strconv.Itoa(iteracion)}
	var reply raft.ResultadoRemoto
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.SometerOperacionRaft",
		operacion, &reply, compromiseTime*time.Millisecond)
	fmt.Println("Valor " + reply.ValorADevolver + " escrito.")
	// check.CheckError(err, "Error en llamada RPC SometerOperacion")
	return err == nil
}

// Pasa un nodo de cualquier estado a estado pausa
func PonerPausa(
	indiceNodo int, nodosRaft []rpctimeout.HostPort) {
	vacio := raft.Vacio{}
	var reply raft.Vacio
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.PonerPausa",
		vacio, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC PonerPausa")
	return
}

func obtenerEstadoRemoto(
	indiceNodo int, nodosRaft []rpctimeout.HostPort) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

func main() {
	//ejecutar como main direccionnodo1 direciconnodo2 ....
	// var nodos []rpctimeout.HostPort
	// // Resto de argumento son los end points como strings
	// // De todas la replicas-> pasarlos a HostPort
	// for _, endPoint := range os.Args[1:] {
	// 	nodos = append(nodos, rpctimeout.HostPort(endPoint))
	// }

	nodos := rpctimeout.StringArrayToHostPortArray([]string{"nodo-0.raft.default.svc.cluster.local:29000","nodo-1.raft.default.svc.cluster.local:29000","nodo-2.raft.default.svc.cluster.local:29000"})
	_, _, _, idLider := obtenerEstadoRemoto(0,nodos)
	someterOperacion(idLider,0,nodos)
	PonerPausa(0,nodos)
	PonerPausa(1,nodos)
	PonerPausa(2,nodos)
}
