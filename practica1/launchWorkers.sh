#!/bin/bash
# user="a821259"
user="fjpizarro"
#
# Autores:
#	Jorge Solán Morote NIA: 816259
#	Francisco Javier Pizarro NIA:  821259
# Fecha de última revisión:
#	06/10/2022
# Descripción de la estructura del código:
#	lee ips del fichero ips.txt y lanza el usuario
#   especificado en la variable user por ssh
#   si el servidor esta vivo lanza un worker en el, todo
#   esto para cada una de las ips
#

while read ip
do  
    ssh -n ${user}@${ip} "exit" &> /dev/null
    if [ $? -eq 0 ] 
    then 
        echo "Máquina ${ip} activa"
        ssh -n ${user}@${ip} "./masterWorker --worker &> logWorkers.txt &"
    else echo "${ip} no es accesible"
    fi
done < ips.txt