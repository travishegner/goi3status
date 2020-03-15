package modules

import (
	"bufio"
	"fmt"
	"os"

	"github.com/travishegner/goi3status/types"
)

var modules = make(map[string]types.CreateModule)

func addModMap(name string, newFunc types.CreateModule) {
	modules[name] = newFunc
}

// GetModule returns a newly created module based on it's configuration name
func GetModule(name string, mc types.ModuleConfig) (types.Module, error) {
	cm, ok := modules[name]
	if !ok {
		return nil, fmt.Errorf("no module named %v is registered", name)
	}

	return cm(mc), nil
}

// GetColor returns a color between green and red where 0 = green and 100 = red
func GetColor(n float64) string {
	if n > 1 {
		n = 1
	}
	// #00FF00
	r := int(255 * (n * 2))
	g := 255
	b := 0

	if r >= 255 {
		r = 255
		g = int(255 * ((1 - n) * 2))
	}

	if g > 255 {
		g = 255
	}

	return fmt.Sprintf("#%0.2x%0.2x%0.2x", r, g, b)
}

func readLine(path string) string {
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	return scanner.Text()
}
