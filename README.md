# sis_distribuidos_tp1
Trabajo práctico #1: Map-Reduce 

Les dejo una breve descripción de la estructura del proyecto.

1. Directorios _coordinator_ y _worker_: Contienen toda la funcionalidad implementada por el coordinator y el worker.

2. Directorio _internal_: Contiene toda la lógica de una parte del programa. Tanto el coordinator como worker tienen un 
directorio internal  (el cual contiene toda la lógica necesaria para que puedan funcionar).

3. Directorio _pkg_: Contiene código reutilizable por el coordinator y el worker (código común a ambos).

4. Directorio _tests_: Código para ejecutar los tests del proyecto.