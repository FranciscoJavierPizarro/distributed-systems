package main
import (
	"escritor"
	"os"
	"log"
	"strconv"
)
func main() {
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Uso: ./main pid nProc")
	}
	nProc, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Uso: ./main pid nProc")
	}
	escritor.Start(pid, nProc)
}