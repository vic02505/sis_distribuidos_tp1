# sis_distribuidos_tp1
Trabajo práctico #1: Map-Reduce 

## Estructura del proyecto

1. **Directorios _coordinator_ y _worker_**: Contienen toda la funcionalidad implementada por el coordinator y el worker.

2. **Directorio _internal_**: Contiene toda la lógica de una parte del programa. Tanto el coordinator como worker tienen un 
directorio internal (el cual contiene toda la lógica necesaria para que puedan funcionar).

3. **Directorio _pkg_**: Contiene código reutilizable por el coordinator y el worker (código común a ambos).

4. **Directorio _tests_**: Código para ejecutar los tests del proyecto.

5. **Directorio _plugins_**: Contiene los plugins de Go con las funciones Map y Reduce para diferentes aplicaciones.

6. **Directorio _mr_**: Contiene tipos comunes compartidos entre el sistema y los plugins.

## Como usar


### Pasos para ejecutar

1. **Compilar los plugins:**
   ```bash
   cd plugins/
   go build -buildmode=plugin tu_plugin.go
   cd ..
   ```

2. **Ejecutar la versión secuencial:**
   ```bash
   go run sequential.go plugins/tu_plugin.so archivos_entrada...
   ```

3. **Ejecutar la version distribuida:**
   - En una terminal, iniciar el coordinator:
     ```bash
     go run coordinator.go cant_reducers archivos_entrada...
     ```
   - En otras terminales, iniciar los workers:
     ```bash
     go run worker.go plugins/tu_plugin.so
     ```
4. **Ejecutar los tests:**
   ```bash
   cd tests/
   go run test_runner.go
   cd ..
   ```