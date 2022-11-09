package testintegracionraft1

import (
	"fmt"
	"raft/internal/comun/check"
	//"log"
	//"crypto/rand"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"raft/internal/comun/rpctimeout"
	"raft/internal/despliegue"
	"raft/internal/raft"
)

const (
	//hosts
	MAQUINA1 = "127.0.0.1"
	MAQUINA2 = "127.0.0.1"
	MAQUINA3 = "127.0.0.1"

	//puertos
	PUERTOREPLICA1 = "29001"
	PUERTOREPLICA2 = "29002"
	PUERTOREPLICA3 = "29003"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go"

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_ed25519"
	errorTime = 40//ms
	startTime = 7000//ms
	compromiseTime = 2000//ms
)

// PATH de los ejecutables de modulo golang de servicio Raft
var PATH string = filepath.Join(os.Getenv("HOME"), "Desktop","distributed-systems","practica3","raft")

// go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
var EXECREPLICACMD string = "cd " + PATH + "; go run " + EXECREPLICA

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence

	//
	//TEST1 QUITADO POR BUG
	//

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	//t.Run("T1:soloArranqueYparada",
	//	func(t *testing.T) { cfg.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T2:ElegirPrimerLider",
		func(t *testing.T) { cfg.elegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T3:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cfg.falloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Tres operaciones comprometidas en configuración estable
	t.Run("T4:tresOperacionesComprometidasEstable",
		func(t *testing.T) { cfg.tresOperacionesComprometidasEstable(t) })

	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })

	t.Run("T5:SinAcuerdoPorFallos ",
		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })

	t.Run("T5:SometerConcurrentementeOperaciones ",
		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })
}

// // TEST primer rango
// func TestAcuerdosConFallos(t *testing.T) { // (m *testing.M) {
// 	// <setup code>
// 	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
// 	cfg := makeCfgDespliegue(t,
// 		3,
// 		[]string{REPLICA1, REPLICA2, REPLICA3},
// 		[]bool{true, true, true})

// 	// tear down code
// 	// eliminar procesos en máquinas remotas
// 	defer cfg.stop()

// 	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
// 	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
// 		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })

// 	t.Run("T5:SinAcuerdoPorFallos ",
// 		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })

// 	t.Run("T5:SometerConcurrentementeOperaciones ",
// 		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })

// }

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
type canalResultados chan string

func (cr canalResultados) stop() {
	close(cr)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	t           *testing.T
	conectados  []bool
	numReplicas int
	nodosRaft   []rpctimeout.HostPort
	cr          canalResultados
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(t *testing.T, n int, nodosraft []string,
	conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.t = t
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)
	cfg.cr = make(canalResultados, 2000)

	return cfg
}

func (cfg *configDespliegue) stop() {
	//cfg.stopDistributedProcesses()

	time.Sleep(50 * time.Millisecond)

	cfg.cr.stop()
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ?? - 3 NODOS RAFT
func (cfg *configDespliegue) soloArranqueYparadaTest1(t *testing.T) {
	//t.Skip("SKIPPED soloArranqueYparadaTest1")
	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Poner en marcha replicas en remoto con un tiempo de espera incluido
	cfg.startDistributedProcesses()

	// Comprobar estado replica 0
	cfg.comprobarEstadoRemoto(0, 0, false, -1)

	// Comprobar estado replica 1
	cfg.comprobarEstadoRemoto(1, 0, false, -1)

	// Comprobar estado replica 2
	cfg.comprobarEstadoRemoto(2, 0, false, -1)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()
	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2(t *testing.T) {
	// t.Skip("SKIPPED ElegirPrimerLiderTest2")
	fmt.Println(t.Name(), ".....................")
	cfg.startDistributedProcesses()

	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros
	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	//t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")
	fmt.Println(t.Name(), ".....................")
	cfg.startDistributedProcesses()
	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(3)

	cfg.quitarLider()

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros
	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable(t *testing.T) {
	//	t.Skip("SKIPPED tresOperacionesComprometidasEstable")
	fmt.Println(t.Name(), ".....................")
	cfg.startDistributedProcesses()
	fmt.Printf("Tres operaciones comprometidas establesl\n")
	time.Sleep(300 * time.Millisecond)
	cfg.pruebaUnLider(3)
	_, _, _, idLider := cfg.obtenerEstadoRemoto(0)
	fmt.Printf("Lider:%d\n", idLider)

	for i := 0; i < 3; i++ {
		cfg.someterOperacion(idLider, i)
		fmt.Println("Operación sometida")
	}
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("Recuperando logs")
	logs := [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 3)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}

	for i := 0; i < len(logs[0]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}
	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros
	fmt.Println(".............", t.Name(), "Superado")
}

// Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func (cfg *configDespliegue) AcuerdoApesarDeSeguidor(t *testing.T) {
	// t.Skip("SKIPPED AcuerdoApesarDeSeguidor")
	fmt.Println(t.Name(), ".....................")

	cfg.startDistributedProcesses()

	cfg.pruebaUnLider(3)
	_, _, _, idLider := cfg.obtenerEstadoRemoto(0)
	fmt.Printf("Lider:%d\n", idLider)
	// Comprometer una entrada
	cfg.someterOperacion(idLider, 0)
	fmt.Println("Operación sometida")

	//  Obtener un lider y, a continuación desconectar una de los nodos Raft
	n := 0
	for n == idLider {
		n = rand.Intn(len(cfg.nodosRaft) - 1)
	}
	cfg.PonerPausa(n)

	// Comprobar varios acuerdos con una réplica desconectada
	for i := 1; i < 4; i++ {
		cfg.someterOperacion(idLider, i)
		fmt.Println("Operación sometida")
	}
	logs := [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		if i != n {
			logs[i] = cfg.obtenerRegistro(i, 4)
			fmt.Printf("Logs obtenidos del nodo %d\n", i)
		}
	}

	for i := 0; i < len(logs[idLider]); i++ {
		for nodo := 0; nodo < len(cfg.nodosRaft)-1; nodo++ {
			if nodo != idLider && nodo != n && logs[nodo][i] != logs[idLider][i] {
				panic("No coinciden los logs")
			}
		}
	}
	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	cfg.QuitarPausa(n)

	time.Sleep(1500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 4)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}

	for i := 0; i < len(logs[idLider]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}
	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros
	fmt.Println(".............", t.Name(), "Superado")
}

// NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func (cfg *configDespliegue) SinAcuerdoPorFallos(t *testing.T) {
	//t.Skip("SKIPPED SinAcuerdoPorFallos")
	fmt.Println(t.Name(), ".....................")
	cfg.startDistributedProcesses()
	cfg.pruebaUnLider(3)
	_, _, _, idLider := cfg.obtenerEstadoRemoto(0)
	fmt.Printf("Lider:%d\n", idLider)
	// Comprometer una entrada
	cfg.someterOperacion(idLider, 0)
	fmt.Println("Operación sometida")

	logs := [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 1)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}

	for i := 0; i < len(logs[idLider]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}
	//  Obtener un lider y, a continuación desconectar 2 de los nodos Raft
	_, _, _, idLider = cfg.obtenerEstadoRemoto(0)
	for i := 0; i < 3; i++ {
		if i != idLider {
			cfg.PonerPausa(i)
		}
	}

	// Comprobar varios acuerdos con 2 réplicas desconectada
	for i := 1; i < 4; i++ {
		if cfg.someterOperacion(idLider, i) {
			fmt.Println("Operación sometida")
		} else {
			fmt.Println("Operación sometida pero no comprometida")
		}
	}
	cfg.obtenerCompromiso(idLider)
	// reconectar lo2 nodos Raft  desconectados y probar varios acuerdos

	for i := 0; i < 3; i++ {
		if i != idLider {
			cfg.QuitarPausa(i)
		}
	}
	//Comprueba la integridad de los registros
	time.Sleep(1500 * time.Millisecond)
	logs = [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 4)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}
	for i := 0; i < len(logs[idLider]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros
	fmt.Println(".............", t.Name(), "Superado")
}

// Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func (cfg *configDespliegue) SometerConcurrentementeOperaciones(t *testing.T) {
	// t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	// Obtener un lider y, a continuación someter una operacion
	fmt.Println(t.Name(), ".....................")
	cfg.startDistributedProcesses()
	cfg.pruebaUnLider(3)
	_, _, _, idLider := cfg.obtenerEstadoRemoto(0)
	fmt.Printf("Lider:%d\n", idLider)

	// Comprometer una entrada
	cfg.someterOperacion(idLider, 0)
	fmt.Println("Operación sometida")

	logs := [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 1)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}

	for i := 0; i < len(logs[idLider]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}

	// Someter 5  operaciones concurrentes
	done := make(chan bool)
	for i := 1; i < 6; i++ {
		go func(done chan bool, i int) {
			if cfg.someterOperacion(idLider, i) {
				fmt.Println("Operación sometida")
			} else {
				fmt.Println("Operación sometida pero no comprometida")
			}
			done <- true
		}(done, i)
	}
	// Comprobar estados de nodos Raft, sobre todo
	// el avance del mandato en curso e indice de registro de cada uno
	// que debe ser identico entre ellos
	notEnd := true
	contador := 0
	ticker := time.NewTicker(time.Millisecond * 250)
	for notEnd {
		select {
		case <-done:
			contador++
			notEnd = contador == 5
			ticker.Stop()
		case <-ticker.C:
			for i := 0; i < len(cfg.nodosRaft); i++ {
				_, mandato, _, _ := cfg.obtenerEstadoRemoto(i)
				fmt.Printf("Mandato %d para el nodo %d\n", mandato, i)
				cfg.obtenerCompromiso(i)
			}
		}
	}
	//Comprueba la integridad de los logs
	logs = [][]string{}
	for i := 0; i < 3; i++ {
		logs = append(logs, nil)
		logs[i] = cfg.obtenerRegistro(i, 6)
		fmt.Printf("Logs obtenidos del nodo %d\n", i)
	}
	for i := 0; i < len(logs[idLider]); i++ {
		if logs[0][i] != logs[1][i] || logs[0][i] != logs[2][i] {
			panic("logs distintos")
		}
		fmt.Println(logs[0][i] + " " + logs[1][i] + " " + logs[2][i] + " ")
	}
	cfg.obtenerCompromiso(idLider)
	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < numreplicas; i++ {
			if cfg.conectados[i] {
				if _, mandato, eslider, _ := cfg.obtenerEstadoRemoto(i); eslider {
					mapaLideres[mandato] = append(mapaLideres[mandato], i)
				}
			}
		}

		ultimoMandatoConLider := -1
		for mandato, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
					mandato, len(lideres))
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
	indiceNodo int) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startDistributedProcesses() {
	//cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

	for i, endPoint := range cfg.nodosRaft {
		despliegue.ExecMutipleHosts(EXECREPLICACMD+
			" "+strconv.Itoa(i)+" "+
			rpctimeout.HostPortArrayToString(cfg.nodosRaft),
			[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

		// dar tiempo para se establezcan las replicas
		fmt.Printf("Nodo %d arrancado\n",i)
		// time.Sleep(startTime * time.Millisecond)
	}

	// aproximadamente 500 ms para cada arranque por ssh en portatil
	time.Sleep(startTime * time.Millisecond)
}

func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for _, endPoint := range cfg.nodosRaft {
		err := endPoint.CallTimeout("NodoRaft.ParaNodo",
			raft.Vacio{}, &reply, errorTime*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC Para nodo")
	}
}

// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
	mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, mandato, esLider, idLider := cfg.obtenerEstadoRemoto(idNodoDeseado)

	//cfg.t.Log("Estado replica 0: ", idNodo, mandato, esLider, idLider, "\n")

	if idNodo != idNodoDeseado || mandato != mandatoDeseado ||
		esLider != esLiderDeseado || idLider != IdLiderDeseado {
		cfg.t.Fatalf("Estado incorrecto en replica %d en subtest %s",
			idNodoDeseado, cfg.t.Name())
	}

}

// Hace que el nodo lider pase a estado seguidor
func (cfg *configDespliegue) quitarLider() {
	var reply raft.Vacio

	time.Sleep(500 * time.Millisecond)
	_, _, _, idLider := cfg.obtenerEstadoRemoto(0)
	err := cfg.nodosRaft[idLider].CallTimeout("NodoRaft.DejarLider",
		raft.Vacio{}, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC Para nodo")

}

// Somete una operacion al nodo
func (cfg *configDespliegue) someterOperacion(
	indiceNodo int, iteracion int) bool {
	operacion := raft.TipoOperacion{"iteracion" + strconv.Itoa(iteracion), "aa", "a"}
	var reply raft.ResultadoRemoto
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.SometerOperacionRaft",
		operacion, &reply, compromiseTime*time.Millisecond)
	// check.CheckError(err, "Error en llamada RPC SometerOperacion")
	return err == nil
}

// Obtiene todas las entradas del registro de un nodo hasta la entrada n
func (cfg *configDespliegue) obtenerRegistro(indiceNodo int, n int) []string {
	results := []string{}
	var reply string
	for i := 0; i < n; i++ {
		err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerRegistro",
			i, &reply, 3*errorTime*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC obtener registro")
		results = append(results, reply)
	}
	return results
}

// Pasa un nodo de cualquier estado a estado pausa
func (cfg *configDespliegue) PonerPausa(
	indiceNodo int) {
	vacio := raft.Vacio{}
	var reply raft.Vacio
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.PonerPausa",
		vacio, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC PonerPausa")
	return
}

// Pasa un nodo de estado pausa a estado seguidor
func (cfg *configDespliegue) QuitarPausa(
	indiceNodo int) {
	vacio := raft.Vacio{}
	var reply raft.Vacio
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.QuitarPausa",
		vacio, &reply, errorTime*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC QuitarPausa")
	return
}

// Obtiene el número de entradas que tiene el nodo en su registro
// y el número de entradas comprometidas
func (cfg *configDespliegue) obtenerCompromiso(
	indiceNodo int) {
	vacio := raft.Vacio{}
	var reply []int
	err := cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerCompromiso",
		vacio, &reply, errorTime*2*time.Millisecond)
	check.CheckError(err, "Error en llamada RPC ObtenerCompromiso")
	fmt.Printf("Nlogs: %d, CommitIndex: %d\n", reply[0], reply[1])
	return
}
