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

## Ejecución de la versión secuencial


### Pasos para ejecutar

1. **Compilar el plugin de word count:**
   ```bash
   cd plugins/
   go build -buildmode=plugin wc.go
   cd ..
   ```

2. **Ejecutar la versión secuencial:**
   ```bash
   go run sequential.go plugins/wc.so archivos_entrada...
   ```

