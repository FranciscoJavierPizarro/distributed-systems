#### Estructura RA

Gestiona los mensajes con los demás procesos y el acceso a SC

Internamente usa un canal de replys para contar las respuestas a las solicitudes de acceso a SC, se comunica con otros procesos por medio del buzón, tiene una lista de booleanos en forma entero para los procesos cuyo request esta pendiente de responder