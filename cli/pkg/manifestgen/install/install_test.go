package install

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	opts := MakeDefaultOptions()
	opts.TolerationKeys = []string{"node.kubernetes.io/controllers"}
	output, err := Generate(opts, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, component := range opts.Components {
		img := fmt.Sprintf("%s/%s", opts.Registry, component)
		if !strings.Contains(output.Content, img) {
			t.Errorf("component image '%s' not found", img)
		}
	}

	if !strings.Contains(output.Content, opts.TolerationKeys[0]) {
		t.Errorf("toleration key '%s' not found", opts.TolerationKeys[0])
	}

	warning := GetGenWarning(opts)
	if !strings.HasPrefix(output.Content, warning) {
		t.Errorf("Generation warning '%s' not found", warning)
	}

	fmt.Println(output)
}
