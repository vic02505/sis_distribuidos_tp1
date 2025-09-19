package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"tp1/mr"

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

	tempFiles := make([]*os.File, reducerNumber)
	for i := int32(0); i < reducerNumber; i++ {
		tempFileName := fmt.Sprintf("mr-%d-%d", workerId, i)
		tempFiles[i], err = os.Create(tempFileName)
		if err != nil {
			return fmt.Errorf("error creando archivo temporal %s: %v", tempFileName, err)
		}
		defer tempFiles[i].Close()
	}

	for _, kv := range mapResult {
		reduceIndex := ihash(kv.Key) % int(reducerNumber)
		_, err = tempFiles[reduceIndex].WriteString(fmt.Sprintf("%s %s\n", kv.Key, kv.Value))
		if err != nil {
			return fmt.Errorf("error escribiendo en archivo temporal: %v", err)
		}
	}

	return nil
}

func main() {

	workerUuid := uuid.New().String()

	socketPath := "/tmp/mr-socket.sock"
	for {
		conn, err := grpc.Dial("unix://"+socketPath, grpc.WithInsecure())
		if err != nil {
			log.Printf("Error conectando al coordinator: %v", err)
			continue
		}
		defer conn.Close()

		client := pb.NewServerClient(conn)

		resp, err := client.AskForWork(context.Background(), &pb.ImFree{WorkerUuid: workerUuid})
		if err != nil {
			log.Printf("Error al solicitar trabajo: %v", err)
			continue
		}

		mapF, reduceF, err := loadPlugin("plugins/wc.so")
		_ = reduceF // Declarado pero no usado por ahora para evitar error de compilador
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
			_, _ = client.MarkWorkAsFinished(context.Background(), &pb.IFinished{WorkerUuid: workerUuid})
		case "Reduce":
			fmt.Println("Ejecutando fase Reduce...")

		case "Work finished":
			fmt.Println("Trabajo completado")
			return
		}
	}
}
