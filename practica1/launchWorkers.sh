#!/bin/bash
# user="a821259"
user="fjpizarro"
sleep 20
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