package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TestResult struct {
	TestName    string
	Sequential  map[string]string
	Distributed map[string]string
	Passed      bool
	Error       string
}

type TestRunner struct {
	projectRoot string
	testDir     string
	plugins     []string
	inputFiles  []string
}

func NewTestRunner(projectRoot string) *TestRunner {
	return &TestRunner{
		projectRoot: projectRoot,
		testDir:     filepath.Join(projectRoot, "tests"),
		plugins:     []string{"wc.so", "inverted_index.so"},
		inputFiles:  []string{"files/test.txt", "files/test2.txt"},
	}
}

func (tr *TestRunner) cleanup() {
	// Limpiar archivos de pruebas anteriores
	os.RemoveAll(filepath.Join(tr.projectRoot, "intermediate"))
	os.RemoveAll(filepath.Join(tr.projectRoot, "output"))
	os.RemoveAll(filepath.Join(tr.projectRoot, "mr-out-*"))
	os.Mkdir(filepath.Join(tr.projectRoot, "intermediate"), 0755)
	os.Mkdir(filepath.Join(tr.projectRoot, "output"), 0755)
}

func (tr *TestRunner) runSequential(plugin string) (map[string]string, error) {
	tr.cleanup()

	// Construir comando secuencial
	args := []string{"run", "sequential.go", filepath.Join("plugins", plugin)}
	args = append(args, tr.inputFiles...)

	cmd := exec.Command("go", args...)
	cmd.Dir = tr.projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando secuencial: %v\nOutput: %s", err, output)
	}

	// Leer resultados
	return tr.readResults("mr-out-*")
}

func (tr *TestRunner) runDistributed(plugin string) (map[string]string, error) {
	tr.cleanup()

	// Iniciar coordinador en background
	coordinatorArgs := []string{"run", "coordinator/coordinator.go", "3"}
	coordinatorArgs = append(coordinatorArgs, tr.inputFiles...)

	coordinatorCmd := exec.Command("go", coordinatorArgs...)
	coordinatorCmd.Dir = tr.projectRoot

	if err := coordinatorCmd.Start(); err != nil {
		return nil, fmt.Errorf("error iniciando coordinator: %v", err)
	}
	defer coordinatorCmd.Process.Kill()

	// Esperar a que el coordinador inicie
	time.Sleep(2 * time.Second)

	// Iniciar workers
	var workerCmds []*exec.Cmd
	for i := 0; i < 2; i++ { // 2 workers
		workerCmd := exec.Command("go", "run", "worker/worker.go", plugin)
		workerCmd.Dir = tr.projectRoot

		if err := workerCmd.Start(); err != nil {
			// Terminar workers anteriores
			for _, cmd := range workerCmds {
				cmd.Process.Kill()
			}
			return nil, fmt.Errorf("error iniciando worker %d: %v", i, err)
		}
		workerCmds = append(workerCmds, workerCmd)
	}

	// Esperar a que termine el procesamiento (máximo 60 segundos)
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Terminar todos los procesos
			for _, cmd := range workerCmds {
				cmd.Process.Kill()
			}
			return nil, fmt.Errorf("timeout esperando que termine el procesamiento distribuido")

		case <-ticker.C:
			// Verificar si hay archivos de salida
			if tr.hasOutputFiles() {
				// Esperar un poco más para asegurar que todo termine
				time.Sleep(2 * time.Second)

				// Terminar workers
				for _, cmd := range workerCmds {
					cmd.Process.Kill()
				}

				// Leer resultados
				return tr.readResults("output/mr-out-*")
			}
		}
	}
}

func (tr *TestRunner) hasOutputFiles() bool {
	pattern := filepath.Join(tr.projectRoot, "output", "mr-out-*")
	files, err := filepath.Glob(pattern)
	return err == nil && len(files) >= 3 // Esperamos al menos 3 archivos de salida
}

func (tr *TestRunner) readResults(pattern string) (map[string]string, error) {
	fullPattern := filepath.Join(tr.projectRoot, pattern)
	files, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("error buscando archivos de salida: %v", err)
	}

	// Si no encuentra archivos con el patrón, intentar buscar archivo secuencial
	if len(files) == 0 && strings.Contains(pattern, "mr-out-*") {
		// Para secuencial, buscar output/mr-out-0
		seqFile := filepath.Join(tr.projectRoot, "output", "mr-out-0")
		if _, err := os.Stat(seqFile); err == nil {
			files = []string{seqFile}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no se encontraron archivos de salida con patrón: %s", pattern)
	}

	results := make(map[string]string)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error leyendo archivo %s: %v", file, err)
		}

		// Normalizar contenido: ordenar líneas para comparación consistente
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		sort.Strings(lines)

		filename := filepath.Base(file)
		results[filename] = strings.Join(lines, "\n")
	}

	return results, nil
}

func (tr *TestRunner) compareResults(sequential, distributed map[string]string) (bool, string) {

	// Combinar todos los resultados distribuidos
	var allDistributedLines []string
	for _, content := range distributed {
		lines := strings.Split(content, "\n")
		allDistributedLines = append(allDistributedLines, lines...)
	}
	sort.Strings(allDistributedLines)
	combinedDistributed := strings.Join(allDistributedLines, "\n")

	// Obtener contenido secuencial
	var sequentialContent string
	if len(sequential) == 1 {
		for _, content := range sequential {
			sequentialContent = content
			break
		}
	} else {
		return false, fmt.Sprintf("Secuencial debería generar 1 archivo, pero generó %d", len(sequential))
	}

	// Comparar contenidos combinados
	if sequentialContent != combinedDistributed {
		return false, fmt.Sprintf("Contenido diferente:\nSecuencial:\n%s\nDistribuido (combinado):\n%s",
			sequentialContent, combinedDistributed)
	}

	return true, "Resultados idénticos (secuencial vs distribuido combinado)"
}

func (tr *TestRunner) runTest(plugin string) TestResult {
	testName := fmt.Sprintf("Test_%s", strings.TrimSuffix(plugin, ".so"))

	fmt.Printf("Ejecutando %s...\n", testName)

	// Ejecutar versión secuencial
	fmt.Printf("  - Ejecutando versión secuencial...")
	sequential, err := tr.runSequential(plugin)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Error:    fmt.Sprintf("Error en versión secuencial: %v", err),
		}
	}
	fmt.Printf(" ✓\n")

	// Ejecutar versión distribuida
	fmt.Printf("  - Ejecutando versión distribuida...")
	distributed, err := tr.runDistributed(plugin)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Error:    fmt.Sprintf("Error en versión distribuida: %v", err),
		}
	}
	fmt.Printf(" ✓\n")

	// Comparar resultados
	fmt.Printf("  - Comparando resultados...")
	passed, message := tr.compareResults(sequential, distributed)
	fmt.Printf(" %s\n", func() string {
		if passed {
			return "✓"
		}
		return "✗"
	}())

	return TestResult{
		TestName:    testName,
		Sequential:  sequential,
		Distributed: distributed,
		Passed:      passed,
		Error:       message,
	}
}

func (tr *TestRunner) runAllTests() []TestResult {
	var results []TestResult

	fmt.Println("=== Ejecutando Tests Automáticos ===")
	fmt.Printf("Directorio del proyecto: %s\n", tr.projectRoot)
	fmt.Printf("Archivos de entrada: %v\n", tr.inputFiles)
	fmt.Printf("Plugins a probar: %v\n\n", tr.plugins)

	for _, plugin := range tr.plugins {
		result := tr.runTest(plugin)
		results = append(results, result)
		fmt.Println()
	}

	return results
}

func (tr *TestRunner) printSummary(results []TestResult) {
	fmt.Println("=== Resumen de Tests ===")

	passed := 0
	failed := 0

	for _, result := range results {
		status := "✗ FALLO"
		if result.Passed {
			status = "✓ ÉXITO"
			passed++
		} else {
			failed++
		}

		fmt.Printf("%s: %s\n", result.TestName, status)
		if !result.Passed {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\nTotal: %d pruebas, %d exitosas, %d fallaron\n", len(results), passed, failed)

	if failed == 0 {
		fmt.Println("🎉 ¡Todos los tests pasaron!")
	} else {
		fmt.Println("❌ Algunos tests fallaron")
	}
}

func main() {
	// Obtener directorio del proyecto
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("Error obteniendo directorio actual:", err)
	}

	// Si estamos en el directorio tests, subir un nivel
	if filepath.Base(projectRoot) == "tests" {
		projectRoot = filepath.Dir(projectRoot)
	}

	// Crear y ejecutar tests
	runner := NewTestRunner(projectRoot)
	results := runner.runAllTests()
	runner.printSummary(results)

	// Exit code basado en resultados
	for _, result := range results {
		if !result.Passed {
			os.Exit(1)
		}
	}
}
