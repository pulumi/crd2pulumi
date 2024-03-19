package codegen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pulumi/crd2pulumi/internal/files"
)

// GenerateFunc is the function that is called by the generator to generate the code.
// It returns a mapping of filename to the contents of said file and any error that may have occurred.
type GenerateFunc func(pg *PackageGenerator, name string) (mapFileNameToData map[string]*bytes.Buffer, err error)

var codeGenFuncs = map[string]GenerateFunc{
	Go:     GenerateGo,
	DotNet: GenerateDotNet,
	NodeJS: GenerateNodeJS,
	Python: GeneratePython,
	Java:   GenerateJava,
}

// PulumiToolName is a symbol that identifies to Pulumi the name of this program.
const PulumiToolName = "crd2pulumi"

// GenerateFromFiles performs the entire CRD codegen process.
// The yamlPaths argument can contain both file paths and URLs.
func GenerateFromFiles(cs *CodegenSettings, yamlPaths []string) error {
	yamlReaders := make([]io.ReadCloser, 0, len(yamlPaths))
	for _, yamlPath := range yamlPaths {
		reader, err := files.ReadFromLocalOrRemote(yamlPath, map[string]string{"Accept": "application/x-yaml, text/yaml"})
		if err != nil {
			return fmt.Errorf("could not open YAML document at %s: %w", yamlPath, err)
		}
		yamlReaders = append(yamlReaders, reader)
	}
	return Generate(cs, yamlReaders)
}

// Generate performs the entire CRD codegen process, reading YAML content from the given readers.
func Generate(cs *CodegenSettings, yamls []io.ReadCloser) error {
	generate, ok := codeGenFuncs[cs.Language]
	if !ok {
		return fmt.Errorf("unsupported language %q, must be one of %q", cs.Language, SupportedLanguages)
	}

	if !cs.Overwrite {
		if dirExists(cs.Path()) {
			return fmt.Errorf("output already exists at %q, use --force to overwrite", cs.Path())
		}
	}

	// Do the actual reading of files from source, may take substantial time depending on the sources.
	pg, err := ReadPackagesFromSource(cs.PackageVersion, yamls)
	if err != nil {
		return err
	}

	// Do actual codegen
	output, err := generate(pg, cs.PackageName)
	if err != nil {
		return fmt.Errorf("failed to generate %q package %q: %w", cs.Language, cs.PackageName, err)
	}
	// Write output to disk
	err = writeFiles(output, cs.Path())
	if err != nil {
		return fmt.Errorf("failed to write %q package %q to disk: %w", cs.Language, cs.PackageName, err)
	}
	return nil
}

// dirExists returns whether a given directory exists.
func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// writeFiles writes the contents of each buffer to its file path, relative to `outputDir`.
// `files` should be a mapping from file path strings to buffers.
func writeFiles(files map[string]*bytes.Buffer, outputDir string) error {
	for path, code := range files {
		outputFilePath := filepath.Join(outputDir, path)
		err := os.MkdirAll(filepath.Dir(outputFilePath), 0755)
		if err != nil {
			return fmt.Errorf("could not create directory to %s: %w", outputFilePath, err)
		}
		file, err := os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("could not create file %s: %w", outputFilePath, err)
		}
		defer file.Close()
		if _, err := code.WriteTo(file); err != nil {
			return fmt.Errorf("could not write to file %s: %w", outputFilePath, err)
		}
	}
	return nil
}
