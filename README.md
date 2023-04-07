# Prácticas de sistemas distribuidos

En este repositorio se encuentran las prácticas relativas a la asignatura de sistemas distribuidos dentro de este encontramos las 3 prácticas.

- Aprendizaje Golang servidor con distintos funcionamientos.
- Fichero distribido de lectura escritura por medio de exclusión mutua.
- Arquitectura RAFT desplegada sobre Kubernetes con Kind.

### Práctica 1

Era la introducción a Golang en ella se encuentra programado un servidor que tiene distintas formas de funcionar internamente mediante el uso de threads y workers, dicho servidor atiende clientes que le solicitan el calculo de números primos de un intervalo dado, a parte de ser la primera práctica de la carrera desarrollada con este lenguaje, gracias a ella podemos obtener métricas de rendimiento muy interesantes sobre el rendimiento del servidor.

### Práctica 2

Consiste en un problema de lectura/escritura distribuida es decir N procesos entre los cuales hay procesos que solo leen y procesos que solo escriben en el mismo fichero, para garantizar la exclusión mutua sobre el acceso de dicho fichero de forma que este no se corrompa por accesos inadecuados se emplea el algoritmo de Ricart-Agrawala adicionalmente para mejorar el rendimiento dado que los procesos lectores se pueden solapar entre si sin causar problema alguno se emplea una matriz de exclusión.

### Práctica 3

Consiste en el despliegue de la arquitectura Raft la cual ofrece sistemas de consenso, consistencia y secuencialidad entre otras cosas sobre alta disponibilidad, el despliegue se realiza sobre Kubernetes concretamente con la ayuda de la herramienta Kind empleando la tecnología de contenedores sobre contenedores, el despliegue cuenta con un statefull set para garantizar aunque sea de forma eventual que el número de replicas es el adecuado, adicionalmente se emplea un Service que hace de DNS. Se emplean imagenes locales de Docker.


## Formas de ejecución

Para ejecutar la primera, el servidor con distintos funcionamientos ejecutar:

```
go run server.go
```

Para ejecutar la segunda, el fichero distribuido ejecutar:
```
go run main.go
```

Para ejecutar la tercera, raft sobre kubernetes realizar los siguientes pasos:

- Compilar los archivos main.go de las carpetas clienttest y srvraft de la carpeta cmd.
- Generar las imagenes locales de docker empleando los compilados anteriores y los DockerFile
- Ejecutar el script kind-with-registry para poner en funcionamiento la herramienta Kind
- Ejecutar el script launch.sh