package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/murlokswarm/app/internal/file"
)

type cefPackage struct {
	// Enable build with race condition checks.
	Race bool

	// Force full build.
	Force bool

	// The directory that contains the sources.
	Sources string

	// The name of the generated package.
	Output string

	// The working directory.
	workingDir string

	// The source directory that contains the resources.
	sourceResources string

	// The package directory that contains the resources.
	resources string

	// The package directory that contains the go executable.
	execDir string

	// The package go exec filename.
	Exec string

	// The system tmp directory.
	tmpDir string

	// The source directory that contains chromium embedded files.
	sourceCef string

	// The package directory that contains chromium embedded files.
	cef string

	once sync.Once
}

func (pkg *cefPackage) init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pkg.workingDir = wd

	if len(pkg.Sources) == 0 || pkg.Sources == "." || pkg.Sources == "./" {
		sources, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		pkg.Sources = sources
	}

	if len(pkg.Output) == 0 {
		out, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}

		out = filepath.Base(out) + ".app"
		out = filepath.Join(wd, out)
		pkg.Output = out
	}

	if !strings.HasSuffix(pkg.Output, ".app") {
		pkg.Output += ".app"
	}

	pkg.sourceResources = filepath.Join(pkg.Sources, "resources")
	pkg.resources = resourcesDir(pkg.Output)
	pkg.execDir, pkg.Exec = executable(pkg.Output)
	pkg.tmpDir = tmpDir()

	pkg.sourceCef = filepath.Join(pkg.Sources, "cef")
	pkg.cef = cefDir(pkg.Output)
}

// Build build the package.
func (pkg *cefPackage) Build(ctx context.Context) error {
	pkg.once.Do(pkg.init)

	printVerbose("building %s", filepath.Base(pkg.Output))
	if err := pkg.buildOutput(); err != nil {
		return err
	}

	printVerbose("building go exec")
	if err := pkg.buildGoExec(ctx); err != nil {
		return err
	}

	printVerbose("building cef")
	if err := pkg.buildCef(ctx); err != nil {
		return err
	}

	return nil
}

func (pkg *cefPackage) buildOutput() error {
	dirs := []string{
		pkg.Output,
		pkg.resources,
		pkg.execDir,
		pkg.cef,
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (pkg *cefPackage) buildGoExec(ctx context.Context) error {
	tmpExec := filepath.Base(pkg.Exec)
	tmpExec = filepath.Join(pkg.tmpDir, tmpExec)

	cmd := []string{"go", "build",
		"-ldflags", "-s",
		"-o", tmpExec,
	}

	if verbose {
		cmd = append(cmd, "-v")
	}

	if pkg.Force {
		cmd = append(cmd, "-a")
	}

	if pkg.Race {
		cmd = append(cmd, "-race")
	}

	if err := execute(ctx, cmd[0], cmd[1:]...); err != nil {
		return err
	}

	return file.Copy(pkg.Exec, tmpExec)
}

func (pkg *cefPackage) buildCef(ctx context.Context) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	src := filepath.Join(pkg.cef, "Chromium Embedded Framework.framework")
	dst := filepath.Join(pkg.cef, "cef.framework")

	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		return err
	}

	return nil
}

func resourcesDir(out string) string {
	return filepath.Join(out, "Contents", "Resources")
}

func executable(out string) (string, string) {
	dir := filepath.Join(out, "Contents", "MacOS")

	ex := filepath.Base(out)
	ex = strings.TrimSuffix(ex, ".app")
	ex = filepath.Join(dir, ex)

	return dir, ex
}

func tmpDir() string {
	return os.Getenv("TMPDIR")
}

func cefDir(out string) string {
	return filepath.Join(out, "Contents", "Frameworks")
}
