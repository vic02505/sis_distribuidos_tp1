package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "plugin"
    "sort"
    "tp1/mr"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Uso: go run sequential.go plugin.so inputfiles...\n")
		os.Exit(1)
	}

	pluginFile := os.Args[1]
	inputFiles := os.Args[2:]

	plug, err := plugin.Open(pluginFile)
	if err != nil {
		log.Fatalf("Error abriendo plugin %s: %v", pluginFile, err)
	}

	mapFunc, err := plug.Lookup("Map")
	if err != nil {
		log.Fatalf("Error encontrando función Map: %v", err)
	}
	
	reduceFunc, err := plug.Lookup("Reduce")
	if err != nil {
		log.Fatalf("Error encontrando función Reduce: %v", err)
	}

	mapF := mapFunc.(func(string, string) []mr.KeyValue)
	reduceF := reduceFunc.(func(string, []string) string)

	fmt.Println("Ejecutando fase Map...")
	var intermediate []mr.KeyValue
	
	for i, filename := range inputFiles {
		fmt.Printf("Procesando archivo %d: %s\n", i, filename)
		
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error leyendo %s: %v", filename, err)
		}

		kva := mapF(filename, string(content))
		intermediate = append(intermediate, kva...)
	}

	fmt.Println("Agrupando resultados...")
	groups := make(map[string][]string)
	
	for _, kv := range intermediate {
		groups[kv.Key] = append(groups[kv.Key], kv.Value)
	}

	fmt.Println("Ejecutando fase Reduce...")

	var keys []string
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	outputFile := "output/mr-out-0"
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creando archivo de salida: %v", err)
	}
	defer file.Close()


	for _, key := range keys {
		values := groups[key]
		result := reduceF(key, values)

		fmt.Fprintf(file, "%v %v\n", key, result)
	}

	fmt.Printf("Resultado guardado en %s\n", outputFile)
	fmt.Printf("Procesadas %d claves únicas de %d pares totales\n", len(keys), len(intermediate))
}

