package main

import (
	"path/filepath"
	"strings"
	"tp1/mr"
)

func Map(filename string, content string) []mr.KeyValue {
	// Usar solo el nombre del archivo, no la ruta completa
	docName := filepath.Base(filename)

	words := strings.Fields(strings.ToLower(content)) // Convertir a minúsculas para consistencia
	var result []mr.KeyValue

	// Usar un mapa para evitar duplicados por documento
	seenWords := make(map[string]bool)

	for _, word := range words {
		// Limpiar puntuación básica
		word = strings.Trim(word, ".,!?;:\"'()[]")
		if word != "" && !seenWords[word] {
			result = append(result, mr.KeyValue{
				Key:   word,
				Value: docName,
			})
			seenWords[word] = true
		}
	}

	return result
}

func Reduce(key string, values []string) string {
	// Eliminar duplicados y crear lista de documentos
	docMap := make(map[string]bool)
	for _, doc := range values {
		docMap[doc] = true
	}

	// Crear lista ordenada de documentos únicos
	var docs []string
	for doc := range docMap {
		docs = append(docs, doc)
	}

	// Retornar como string separado por comas
	return strings.Join(docs, ",")
}
