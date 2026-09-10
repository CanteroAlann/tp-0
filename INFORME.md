# TP0 NIVELADOR

<br>

# Introducción

El proyecto fue desarrollado siguiendo una estructura de capas tanto para el cliente como para el servidor de la siguiente manera:

- **Entidades:** Encargada de modelar las entidades de negocio. Aqui se serializan/   deserializan dichas entidades.

- **Protocolo:** Encargada de empaquetar/desempaquetar siguiendo el protocolo desarrollado los mensajes enviados/recibidos a traves de la red.

- **Comunicacion:** Encargada del envio/recepcion de los mensajes de la aplicacion.

- **Escritua/Lectura de archivos:** Encargada de leer y escribir los distintos archivos del trabajo practico.

<br>

# Protocolo de Comunicacion

El protocolo de comunicacion consta de los siguientes mensajes:

- **Tipo bet:** Mensaje para enviar las apuestas de las agencias. Se compone de la siguiente manera:

<br>

>| Campo | Tamano | Tipo | Descripcion |
>|---|---|---|---|
>| Msg_type | 1 byte | uint8 | tipo de mensaje |
>| Agency_id| 1 byte | uint8 | identificador de la agencia |
>| Bets_amount | 4 bytes | uint32 BE | cantidad de bets dentro del paquete |
>| Payload_length| 4 bytes | uint32 BE | largo en bytes del payload |
>| Payload | variable | bytes | apuestas serializadas |

<br>

A su vez las apuestas se serializan de la siguiente manera:

>| Campo | Tamano | Tipo | Descripcion |
>|---|---|---|---|
>| First_name_length | 1 byte | uint8 | cantidad de bytes del nombre |
>| First_name | variable | string | nombre |
>| Last_name_length| 1 byte | uint8 | cantidad de bytes del apellido  |
>| Last_name | variable | string | apellido |
>| Dni | 4 bytes | uint32 BE | documento del apostador |
>| Birthday_length| 1 byte | uint8 | largo en bytes de la fecha de nacimiento |
>| Birthday | variable | string | fecha de nacimiento |
>| Number | 4 bytes | uint32 BE | numero apostado |

<br>

- **Tipo ack bet:** Mensaje para notificar al cliente de que el mensaje de apuesta fue procesado exitosamente. Se compone de la siguiente manera.

<br>

>| Campo | Tamano | Tipo | Descripcion |
>|---|---|---|---|
>| Msg_type | 1 byte | uint8 | tipo del mensaje |

<br>

- **Tipo all bet sent:** Mensaje para notificar al servidor de que todas las apuestas fueron enviadas y pueda marcar la agencia como lista para el sorteo.

<br>

>| Campo | Tamano | Tipo | Descripcion |
>|---|---|---|---|
>| Msg_type | 1 byte | uint8 | tipo del mensaje |
>| Agency_id| 1 byte | uint8 | identificador de la agencia |

<br>

- **Tipo winners:** Mensaje para notificar al cliente cuales son sus apuestas ganadoras. Se compone de la siguiente manera:

<br>


>| Campo | Tamano | Tipo | Descripcion |
>|---|---|---|---|
>| Msg_type | 1 byte | uint8 | tipo de mensaje |
>| Payload_length| 4 bytes | uint32 BE | largo en bytes del payload |
>| Bets_amount | 4 bytes | uint32 BE | cantidad de bets dentro del paquete |
>| Payload | variable | bytes | apuestas serializadas |

<br>


# Manejo de Concurrencia

Para manejar la concurrencia decidi implementar threads creando un thread distinto cada vez que se recibe una nueva conexion.
Para el manejo de la seccion critica como lo es el acceso a la clase lottery encargada del guardado y lectura de las apuestas implemente un proceso coordinador que corre en su propio thread y se comunica a traves de cola de mensajes siguiendo el modelo de actores.
Una vez alcanzado el quorum inicia el sorteo y notifica por la cola asociada al proceso de cada cliente (en el servidor) las apuestas ganadoras para que estas sean enviadas a cada agencia.

 Recibe 3 tipos de mensajes:

 - **SHUTDOWN:** Mensaje para notificar al coordinador que fue recibida un señal SIGTERM. Notifica a todos los procesos a traves de la cola asociada para que puedan finalizar de manera ordenada.

 - **STORE_BETS:** Mensaje para guardar las apuestas recibidadas. Se envian las apuestas ya deserializadas.

 - **AGENCY_SENT_ALL_BETS:** Mensaje para recibir el id de la agencia que ya envio todas sus apuestas y para recibir su cola de mensajes para notificar los ganadores si es alcanzado el quorum o para notificarles que se recibio un SIGTERM.

 A su vez esta clase envia 2 tipos de mensajes:

  - **SHUTDOWN:** Mensaje para notificar a los procesos correspondientes que hubo un SIGTERM.

 - **WINNERS:** Mensaje para notificar a los procesos correspondientes sus respectivos ganadores.


<br>

# Opcionales

Para generar el compose de manera automatica, usar el siguiente comando reemplazando OUTPUT_FILE por el nombre del archivo y CLIENT_COUNT por la cantidad de clientes a generar

```python
python3 generar-compose.py "OUTPUT_FILE" "CLIENT_COUNT"
```

Se provee tambien el script para testear el servidor, cabe destacar que no funciona para la solucion final, ya que fue hecho para testear el echo server.

