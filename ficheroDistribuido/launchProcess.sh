#!/bin/bash
user="a821259"
#
# Autores:
#	Jorge Solán Morote NIA: 816259
#	Francisco Javier Pizarro NIA:  821259
# Fecha de última revisión:
#	20/10/2022
# Guía de uso:
# ./launchProcess.sh reader 1 5
# el primer parametro es el nombre del proceso a invocar si reader o writer
# el segundo parametro es el pid que tiene asignado dicho proceso
# el tercer parametro es el numero total de procesos
# Descripción del fichero:
#	lee la fila pid del fichero users.txt
#   la procesa y extrae de ella la ip
#   prueba si es accesible la ip por medio de ssh
#   en caso de que si lo sea entra en la carpeta
#   distri(contenido explicado en readme.md)
#   y lanza o writer o reader para crear un proceso
#   escritor o lector remoto

IFS=:
nLinea=$2
linea=$(sed -n "${nLinea} p" users.txt)
echo "$linea" | (
    read ip port
    ssh -n ${user}@${ip} "exit" &> /dev/null
    if [ $? -eq 0 ] 
    then 
        echo "Máquina ${ip} activa"
        ssh -n ${user}@${ip} "cd ./distri/;./$1 $2 $3"
    else echo "${ip} no es accesible"
    fi

)
