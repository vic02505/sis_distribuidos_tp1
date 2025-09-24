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

// Nuevo struct para manejar resultados de tests con fallos
type FailureTestResult struct {
	TestName           string
	ReferenceResult    map[string]string
	DistributedResult  map[string]string
	Passed             bool
	Error              string
	WorkerFailures     int
	AttemptsRequired   int
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
		plugins:     []string{"wc.so", "inverted_index.so", "wc_with_fails.so"},
		inputFiles:  []string{"files/test.txt", "files/test2.txt"},
	}
}

func (tr *TestRunner) cleanup() {
	os.RemoveAll(filepath.Join(tr.projectRoot, "intermediate"))
	os.RemoveAll(filepath.Join(tr.projectRoot, "output"))
	os.RemoveAll(filepath.Join(tr.projectRoot, "mr-out-*"))
	os.Mkdir(filepath.Join(tr.projectRoot, "intermediate"), 0755)
	os.Mkdir(filepath.Join(tr.projectRoot, "output"), 0755)
}

func (tr *TestRunner) runSequential(plugin string) (map[string]string, error) {
	tr.cleanup()

	args := []string{"run", "sequential.go", filepath.Join("plugins", plugin)}
	args = append(args, tr.inputFiles...)

	cmd := exec.Command("go", args...)
	cmd.Dir = tr.projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando secuencial: %v\nOutput: %s", err, output)
	}

	return tr.readResults("mr-out-*")
}

func (tr *TestRunner) runDistributed(plugin string) (map[string]string, error) {
	tr.cleanup()

	coordinatorArgs := []string{"run", "coordinator/coordinator.go", "3"}
	coordinatorArgs = append(coordinatorArgs, tr.inputFiles...)

	coordinatorCmd := exec.Command("go", coordinatorArgs...)
	coordinatorCmd.Dir = tr.projectRoot

	if err := coordinatorCmd.Start(); err != nil {
		return nil, fmt.Errorf("error iniciando coordinator: %v", err)
	}
	defer coordinatorCmd.Process.Kill()

	time.Sleep(2 * time.Second)

	// Iniciar workers
	var workerCmds []*exec.Cmd
	for i := 0; i < 2; i++ {
		workerCmd := exec.Command("go", "run", "worker/worker.go", plugin)
		workerCmd.Dir = tr.projectRoot

		if err := workerCmd.Start(); err != nil {
			for _, cmd := range workerCmds {
				cmd.Process.Kill()
			}
			return nil, fmt.Errorf("error iniciando worker %d: %v", i, err)
		}
		workerCmds = append(workerCmds, workerCmd)
	}

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			for _, cmd := range workerCmds {
				cmd.Process.Kill()
			}
			return nil, fmt.Errorf("timeout esperando que termine el procesamiento distribuido")

		case <-ticker.C:
			// Verificar si hay archivos de salida
			if tr.hasOutputFiles() {
				time.Sleep(2 * time.Second)

				for _, cmd := range workerCmds {
					cmd.Process.Kill()
				}

				return tr.readResults("output/mr-out-*")
			}
		}
	}
}

func (tr *TestRunner) hasOutputFiles() bool {
	pattern := filepath.Join(tr.projectRoot, "output", "mr-out-*")
	files, err := filepath.Glob(pattern)
	return err == nil && len(files) >= 3
}

func (tr *TestRunner) readResults(pattern string) (map[string]string, error) {
	fullPattern := filepath.Join(tr.projectRoot, pattern)
	files, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("error buscando archivos de salida: %v", err)
	}

	if len(files) == 0 && strings.Contains(pattern, "mr-out-*") {
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

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		sort.Strings(lines)

		filename := filepath.Base(file)
		results[filename] = strings.Join(lines, "\n")
	}

	return results, nil
}

func (tr *TestRunner) compareResults(sequential, distributed map[string]string) (bool, string) {
	var allDistributedLines []string
	for _, content := range distributed {
		lines := strings.Split(content, "\n")
		allDistributedLines = append(allDistributedLines, lines...)
	}
	sort.Strings(allDistributedLines)
	combinedDistributed := strings.Join(allDistributedLines, "\n")

	var sequentialContent string
	if len(sequential) == 1 {
		for _, content := range sequential {
			sequentialContent = content
			break
		}
	} else {
		return false, fmt.Sprintf("Secuencial debería generar 1 archivo, pero generó %d", len(sequential))
	}

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

func (tr *TestRunner) runDistributedWithFailureDetection(plugin string) (map[string]string, int, int, error) {
	maxAttempts := 5
	workerFailureCount := 0

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("    Intento %d/%d...", attempt, maxAttempts)

		tr.cleanup()

		// Iniciar coordinador
		coordinatorArgs := []string{"run", "coordinator/coordinator.go", "3"}
		coordinatorArgs = append(coordinatorArgs, tr.inputFiles...)

		coordinatorCmd := exec.Command("go", coordinatorArgs...)
		coordinatorCmd.Dir = tr.projectRoot

		if err := coordinatorCmd.Start(); err != nil {
			fmt.Printf(" ✗ (error coordinador)\n")
			continue
		}

		time.Sleep(3 * time.Second)

		var workerCmds []*exec.Cmd
		initialWorkers := 2

		for i := 0; i < initialWorkers; i++ {
			workerCmd := exec.Command("go", "run", "worker/worker.go", plugin)
			workerCmd.Dir = tr.projectRoot

			if err := workerCmd.Start(); err != nil {
				coordinatorCmd.Process.Kill()
				fmt.Printf(" ✗ (error worker)\n")
				break
			}
			workerCmds = append(workerCmds, workerCmd)
		}

		if len(workerCmds) != initialWorkers {
			continue
		}

		success, failures := tr.monitorExecutionWithFailureDetection(coordinatorCmd, &workerCmds, plugin)
		workerFailureCount += failures

		if success {
			fmt.Printf(" ✓ (%d fallos detectados)\n", failures)

			results, err := tr.readResults("output/mr-out-*")
			if err == nil {
				return results, workerFailureCount, attempt, nil
			}
		} else {
			fmt.Printf(" ✗ (falló completamente)\n")
		}

		time.Sleep(2 * time.Second)
	}

	return nil, workerFailureCount, maxAttempts, fmt.Errorf("no se pudo completar después de %d intentos con %d fallos detectados", maxAttempts, workerFailureCount)
}

func (tr *TestRunner) monitorExecutionWithFailureDetection(coordinatorCmd *exec.Cmd, workerCmds *[]*exec.Cmd, plugin string) (bool, int) {
	timeout := time.After(120 * time.Second)
	workerReplacer := time.NewTicker(5 * time.Second)
	failureCount := 0

	defer workerReplacer.Stop()

	coordinatorDone := make(chan struct{})
	go func() {
		coordinatorCmd.Wait()
		close(coordinatorDone)
	}()

	workerDoneChannels := make(map[*exec.Cmd]chan struct{})
	for _, cmd := range *workerCmds {
		done := make(chan struct{})
		workerDoneChannels[cmd] = done
		go func(c *exec.Cmd, ch chan struct{}) {
			c.Wait()
			close(ch)
		}(cmd, done)
	}

	for {
		select {
		case <-timeout:
			coordinatorCmd.Process.Kill()
			for _, cmd := range *workerCmds {
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}
			return false, failureCount

		case <-workerReplacer.C:
			var aliveWorkers []*exec.Cmd
			newWorkerDoneChannels := make(map[*exec.Cmd]chan struct{})

			for _, cmd := range *workerCmds {
				if cmd.Process != nil {
					done := workerDoneChannels[cmd]
					select {
					case <-done:
						// Worker terminó
						failureCount++
						fmt.Printf("    Worker detectado como terminado (PID: %d)\n", cmd.Process.Pid)
					default:
						// Worker sigue vivo
						aliveWorkers = append(aliveWorkers, cmd)
						newWorkerDoneChannels[cmd] = done
					}
				}
			}

			// Actualizar canales de workers vivos
			workerDoneChannels = newWorkerDoneChannels

			workersNeeded := 2 - len(aliveWorkers)
			if workersNeeded > 0 {
				for i := 0; i < workersNeeded && i < 1; i++ {
					newWorker := exec.Command("go", "run", "worker/worker.go", plugin)
					newWorker.Dir = tr.projectRoot
					if err := newWorker.Start(); err == nil {
						aliveWorkers = append(aliveWorkers, newWorker)

						// Crear canal para el nuevo worker
						done := make(chan struct{})
						workerDoneChannels[newWorker] = done
						go func(c *exec.Cmd, ch chan struct{}) {
							c.Wait()
							close(ch)
						}(newWorker, done)

						fmt.Printf("    Nuevo worker iniciado (PID: %d)\n", newWorker.Process.Pid)
					}
				}
			}

			*workerCmds = aliveWorkers

		case <-coordinatorDone:
			if tr.hasOutputFiles() {
				time.Sleep(2 * time.Second)

				for _, cmd := range *workerCmds {
					if cmd.Process != nil {
						cmd.Process.Kill()
					}
				}

				return true, failureCount
			}
		}
	}
}

func (tr *TestRunner) runFailureTest() FailureTestResult {
	testName := "Test_wc_with_fails"

	fmt.Printf("Ejecutando %s (con detección de fallos optimizada)...\n", testName)

	fmt.Printf("  - Ejecutando versión de referencia (wc.so secuencial)...")
	reference, err := tr.runSequential("wc.so")
	if err != nil {
		return FailureTestResult{
			TestName: testName,
			Passed:   false,
			Error:    fmt.Sprintf("Error en versión de referencia: %v", err),
		}
	}
	fmt.Printf(" ✓\n")

	fmt.Printf("  - Ejecutando versión distribuida con fallos (wc_with_fails.so):\n")

	distributed, totalFailures, attempts, err := tr.runDistributedWithFailureDetection("wc_with_fails.so")
	if err != nil {
		return FailureTestResult{
			TestName:         testName,
			ReferenceResult:  reference,
			Passed:           false,
			Error:            fmt.Sprintf("Error en versión distribuida: %v", err),
			WorkerFailures:   totalFailures,
			AttemptsRequired: attempts,
		}
	}

	fmt.Printf("  - Comparando resultados...")
	passed, message := tr.compareResults(reference, distributed)
	fmt.Printf(" %s\n", func() string {
		if passed {
			return "✓"
		}
		return "✗"
	}())

	return FailureTestResult{
		TestName:          testName,
		ReferenceResult:   reference,
		DistributedResult: distributed,
		Passed:            passed,
		Error:             message,
		WorkerFailures:    totalFailures,
		AttemptsRequired:  attempts,
	}
}

func (tr *TestRunner) runAllTests() []TestResult {
	var results []TestResult

	fmt.Println("=== Ejecutando Tests Automáticos ===")
	fmt.Printf("Directorio del proyecto: %s\n", tr.projectRoot)
	fmt.Printf("Archivos de entrada: %v\n", tr.inputFiles)
	fmt.Printf("Plugins a probar: %v\n\n", tr.plugins)

	regularPlugins := []string{"wc.so", "inverted_index.so"}
	for _, plugin := range regularPlugins {
		result := tr.runTest(plugin)
		results = append(results, result)
		fmt.Println()
	}

	// Test especial para wc_with_fails.so
	failureResult := tr.runFailureTest()

	results = append(results, TestResult{
		TestName:    failureResult.TestName,
		Sequential:  failureResult.ReferenceResult,
		Distributed: failureResult.DistributedResult,
		Passed:      failureResult.Passed,
		Error:       failureResult.Error,
	})

	fmt.Printf("\n=== Resultados del Test con Fallos ===\n")
	fmt.Printf("Fallos de workers detectados: %d\n", failureResult.WorkerFailures)
	fmt.Printf("Intentos requeridos: %d\n", failureResult.AttemptsRequired)

	if failureResult.WorkerFailures > 0 {
		fmt.Printf("✅ Se detectaron fallos de workers correctamente\n")
	} else {
		fmt.Printf("⚠️  No se detectaron fallos de workers (puede ser por suerte)\n")
	}

	if failureResult.Passed {
		fmt.Printf("✅ El sistema recuperó correctamente las tareas fallidas\n")
	} else {
		fmt.Printf("❌ El sistema no pudo recuperarse de los fallos\n")
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
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("Error obteniendo directorio actual:", err)
	}

	if filepath.Base(projectRoot) == "tests" {
		projectRoot = filepath.Dir(projectRoot)
	}

	runner := NewTestRunner(projectRoot)
	results := runner.runAllTests()
	runner.printSummary(results)

	for _, result := range results {
		if !result.Passed {
			os.Exit(1)
		}
	}
}
