package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"strings"
	"time"
	"tp1/mr"
	"path/filepath"

	"github.com/google/uuid"

	pb "tp1/protocol/messages"

	"google.golang.org/grpc"
)

func loadPlugin(pluginPath string) (func(string) []mr.KeyValue, func(string, []string) string, error) {
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo plugin %s: %v", pluginPath, err)
	}

	mapFunc, err := plug.Lookup("Map")
	if err != nil {
		return nil, nil, fmt.Errorf("error encontrando función Map: %v", err)
	}

	reduceFunc, err := plug.Lookup("Reduce")
	if err != nil {
		return nil, nil, fmt.Errorf("error encontrando función Reduce: %v", err)
	}

	mapF := mapFunc.(func(string) []mr.KeyValue)
	reduceF := reduceFunc.(func(string, []string) string)

	return mapF, reduceF, nil
}

func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func executeMapTask(mapF func(string) []mr.KeyValue, filePath string, workerId int32, reducerNumber int32) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo %s: %v", filePath, err)
	}

	mapResult := mapF(string(content))

	fmt.Printf("DEBUG: workerId=%d, reducerNumber=%d, mapResult length=%d\n",
		workerId, reducerNumber, len(mapResult))

	if reducerNumber <= 0 {
		return fmt.Errorf("reducerNumber debe ser mayor que 0, recibido: %d", reducerNumber)
	}

	tempFiles := make([]*os.File, reducerNumber)
	for i := int32(0); i < reducerNumber; i++ {
		tempFileName := fmt.Sprintf("intermediate/mr-%d-%d", workerId, i+1)
		tempFiles[i], err = os.Create(tempFileName)
		if err != nil {
			return fmt.Errorf("error creando archivo temporal %s: %v", tempFileName, err)
		}
		defer tempFiles[i].Close()
	}

	for _, kv := range mapResult {
		hashValue := ihash(kv.Key)
		reduceIndex := hashValue % int(reducerNumber)

		fmt.Printf("DEBUG: key='%s', hash=%d, reduceIndex=%d\n",
			kv.Key, hashValue, reduceIndex)

		_, err = tempFiles[reduceIndex].WriteString(fmt.Sprintf("%s %s\n", kv.Key, kv.Value))
		if err != nil {
			return fmt.Errorf("error escribiendo en archivo temporal: %v", err)
		}
	}

	return nil
}

func executeReduceTask(reduceF func(string, []string) string, reduceTaskId int32, nMapTasks int32) error {

	pattern := fmt.Sprintf("intermediate/mr-*-%d", reduceTaskId)
    files, err := filepath.Glob(pattern)
    if err != nil {
        return fmt.Errorf("error buscando archivos con patrón %s: %v", pattern, err)
    }
    
    fmt.Printf("DEBUG: Encontrados %d archivos: %v\n", len(files), files)
    
    var allKeyValues []mr.KeyValue
    
    for _, filename := range files {
        fmt.Printf("DEBUG: Leyendo archivo: %s\n", filename)
        
        content, err := ioutil.ReadFile(filename)
        if err != nil {
            log.Printf("Error leyendo %s: %v", filename, err)
            continue
        }
        
        keyValues := parseIntermediateFile(string(content))
        allKeyValues = append(allKeyValues, keyValues...)
    }

	grouped := groupByKey(allKeyValues)

	outputFile := fmt.Sprintf("output/mr-out-%d", reduceTaskId)
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creando archivo de salida: %v", err)
	}
	defer file.Close()

	for key, values := range grouped {
		result := reduceF(key, values)
		file.WriteString(fmt.Sprintf("%s %s\n", key, result))
	}

	return nil
}

func parseIntermediateFile(content string) []mr.KeyValue {
	var keyValues []mr.KeyValue
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			keyValues = append(keyValues, mr.KeyValue{
				Key:   parts[0],
				Value: parts[1],
			})
		}
	}

	return keyValues
}

func groupByKey(keyValues []mr.KeyValue) map[string][]string {
	grouped := make(map[string][]string)

	for _, kv := range keyValues {
		grouped[kv.Key] = append(grouped[kv.Key], kv.Value)
	}

	return grouped
}

func main() {

	workerUuid := uuid.New().String()

	socketPath := "/tmp/mr-socket.sock"
	for {
		conn, err := grpc.Dial("unix://"+socketPath, grpc.WithInsecure())
		if err != nil {
			log.Printf("Error conectando al coordinator: %v", err)
			return
		}
		defer conn.Close()

		client := pb.NewServerClient(conn)

		time.Sleep(1 * time.Second)

		resp, err := client.AskForWork(context.Background(), &pb.ImFree{WorkerUuid: workerUuid})
		if err != nil {
            if strings.Contains(err.Error(), "connection") &&
               strings.Contains(err.Error(), "Unavailable") {
                log.Printf("Worker %s - Coordinator parece cerrado, terminando", workerUuid)
                return
            }
			log.Printf("Error al solicitar trabajo: %v", err)
			continue
		}

		mapF, reduceF, err := loadPlugin("plugins/wc.so")

		if err != nil {
			log.Printf("Error cargando plugin: %v", err)
			continue
		}

		switch resp.WorkType {
		case "Map":
			err = executeMapTask(mapF, resp.FilePath, resp.WorkerId, resp.ReducerNumber)
			if err != nil {
				log.Printf("Error ejecutando Map: %v", err)
				continue
			}
			_, err = client.MarkWorkAsFinished(context.Background(), &pb.IFinished{WorkerUuid: workerUuid, WorkFinished: resp.FilePath, WorkType: "Map"})
			if err != nil {
                if strings.Contains(err.Error(), "connection") &&
                   strings.Contains(err.Error(), "Unavailable") {
                    log.Printf("Worker %s - Coordinator parece cerrado, terminando", workerUuid)
                    return
                }
                log.Printf("Error marcando Map como terminado: %v", err)
                continue
            }
		case "Reduce":
			fmt.Printf("DEBUG: reduceTaskId=%d, nMapTasks=%d\n", resp.WorkerId, resp.MapNumber)
			err = executeReduceTask(reduceF, resp.WorkerId, resp.MapNumber)
			if err != nil {
				log.Printf("Error ejecutando Reduce: %v", err)
				continue
			}
			_, err = client.MarkWorkAsFinished(context.Background(), &pb.IFinished{WorkerUuid: workerUuid, WorkFinished: resp.FilePath, WorkType: "Reduce"})
            if err != nil {
                if strings.Contains(err.Error(), "connection") &&
                   strings.Contains(err.Error(), "Unavailable") {
                    log.Printf("Worker %s - Coordinator parece cerrado, terminando", workerUuid)
                    return
                }
                log.Printf("Error marcando Reduce como terminado: %v", err)
                continue
            }

		case "Work finished":
			fmt.Println("Trabajo completado")
			return
		}
	}
}
