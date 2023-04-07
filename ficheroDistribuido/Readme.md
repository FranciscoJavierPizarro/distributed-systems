# INSTRUCCIONES DEPLOY

Utilizar un equipo del lab 1.02 cuya ssh key este autorizada para todos los demás
Crear una carpeta llamada distri en el home
En dicha carpeta debemos tener:
1. users.txt
2. ejecutables de main reader y writer
3. launchProcess.sh
4. tantos ficheros de ntexto como procesos haya llamados 1-nProc.txt
Modificar en fichero launchProcess.sh el user
Modificar IPS de users.txt en base a equipos disponibles del lab
Ejecutar desde dentro de dicha carpeta el main con flag remote

#### Estructura RA

Gestiona los mensajes con los demás procesos y el acceso a SC

Internamente usa un canal de replys para contar las respuestas a las solicitudes de acceso a SC, se comunica con otros procesos por medio del buzón, tiene una lista de booleanos en forma entero para los procesos cuyo request esta pendiente de responder