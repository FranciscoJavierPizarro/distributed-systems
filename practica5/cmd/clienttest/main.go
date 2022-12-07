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

// Somete una operacion de lectura y coteja el resultado obtenido y el esperado
func someterLectura(
	indiceNodo int, iteracion int, nodosRaft []rpctimeout.HostPort) bool {
	operacion := raft.TipoOperacion{"lectura", "iteracion" + strconv.Itoa(iteracion),""}
	var reply raft.ResultadoRemoto
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.SometerOperacionRaft",
		operacion, &reply, compromiseTime*time.Millisecond)
	if(reply.ValorADevolver != strconv.Itoa(iteracion)) {
		fmt.Println("ERROR lectura de un valor que no corresponde con el esperado\n")
	}
	fmt.Println("Valor " + reply.ValorADevolver + " leído correctamente.")
	// check.CheckError(err, "Error en llamada RPC SometerOperacion")
	return err == nil
}


func obtenerEstadoRemoto(
	indiceNodo int, nodosRaft []rpctimeout.HostPort) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

// Obtiene todas las entradas del registro de un nodo hasta la entrada n
func obtenerRegistro(indiceNodo int, n int, nodosRaft []rpctimeout.HostPort) []string {
	results := []string{}
	var reply string
	for i := 0; i < n; i++ {
		err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerRegistro",
			i, &reply, 3*errorTime*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC obtener registro")
		results = append(results, reply)
		fmt.Printf("Entrada %d: %s\n",i,reply)
	}
	return results
}

// Obtiene el número de entradas que tiene el nodo en su registro
// y el número de entradas comprometidas
func obtenerCompromiso(
	indiceNodo int, nodosRaft []rpctimeout.HostPort) {
	vacio := raft.Vacio{}
	var reply []int
	err := nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerCompromiso",
		vacio, &reply, errorTime*2*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerCompromiso")
	fmt.Printf("Nlogs: %d, CommitIndex: %d\n", reply[0], reply[1])
	return
}


func main() {
	nodos := rpctimeout.StringArrayToHostPortArray([]string{"nodo-0.raft.default.svc.cluster.local:29000","nodo-1.raft.default.svc.cluster.local:29000","nodo-2.raft.default.svc.cluster.local:29000"})
	
	
	var option int = -1
	var nlec int = 0
	var nodo int = 0
	nOp := 0
	nL:=0
	fmt.Println("Cliente interactivo con nodos raft")
	for (option != 0) {
		fmt.Println("Opciones disponibles: \n\t1 Someter operación de escritura\n\t2 Someter operación de lectura\n\t3 Obtener registro(logs)\n\t4 Obtener nivel de compromiso\n")
		fmt.Scan(&option)
		fmt.Println("")
		switch option {
		case 0:
			break
		case 1:
			_, _, _, idLider := obtenerEstadoRemoto(0,nodos)
			someterOperacion(idLider,nOp+nL,nodos)
			nOp++
		case 2:
			fmt.Println("Introduzca el número de escritura que quiere comprobar")
			fmt.Scan(&nlec)
			fmt.Println("")
			_, _, _, idLider := obtenerEstadoRemoto(0,nodos)
			someterLectura(idLider,nlec,nodos)
			nL++
		case 3:
			fmt.Println("Introduzca el número del nodo cuyo registro quiere visualizar")
			fmt.Scan(&nodo)
			fmt.Println("")
			obtenerRegistro(nodo,nOp+nL,nodos)
		case 4:
			fmt.Println("Niveles de compromiso")
			for i:=0;i<3;i++ {
				obtenerCompromiso(i,nodos)
			}
		default:
			panic("Error opción incorrecta")
		}	
	}
	fmt.Println("Fin del servicio de cliente")
	
}
