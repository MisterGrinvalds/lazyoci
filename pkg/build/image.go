package build

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const lazyociBuilderName = "lazyoci"

// buildImage builds a container image from a Dockerfile using docker buildx.
// Returns the path to a temporary OCI layout directory.
func (b *Builder) buildImage(ctx context.Context, artifact *Artifact) (string, error) {
	// Verify docker buildx is available
	if err := checkBuildx(); err != nil {
		return "", err
	}

	dockerfile := artifact.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	buildContext := artifact.Context
	if buildContext == "" {
		buildContext = "."
	}

	// Resolve paths relative to .lazy file
	dockerfile = b.resolvePath(dockerfile)
	buildContext = b.resolvePath(buildContext)

	// Determine platforms
	platforms := artifact.Platforms
	if len(b.opts.Platforms) > 0 {
		platforms = b.opts.Platforms // CLI override
	}

	b.logf("  Building image from %s...\n", dockerfile)
	if len(platforms) > 0 {
		b.logf("  Platforms: %s\n", strings.Join(platforms, ", "))
	}

	isMultiPlatform := len(platforms) > 1

	// The default "docker" buildx driver does not support OCI export.
	// We need a "docker-container" driver builder for --output type=oci.
	// For single-platform builds we can fall back to docker build + docker save.
	builderName, err := ensureOCICapableBuilder(ctx, b.opts.Quiet)
	if err != nil {
		if isMultiPlatform {
			return "", fmt.Errorf("multi-platform builds require a docker-container buildx builder: %w\n\nCreate one with:\n  docker buildx create --name lazyoci --driver docker-container --use", err)
		}
		// Single-platform: fall back to docker build + docker save + convert
		b.logf("  OCI exporter not available, falling back to docker build + save...\n")
		return b.buildImageFallback(ctx, artifact, dockerfile, buildContext, platforms)
	}

	// Create temp directory for OCI output
	tmpDir, err := os.MkdirTemp("", "lazyoci-image-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	outputTar := filepath.Join(tmpDir, "output.tar")

	// Build the docker buildx command
	args := []string{"buildx", "build", "--builder", builderName}

	// Dockerfile
	args = append(args, "-f", dockerfile)

	// Platforms
	if len(platforms) > 0 {
		args = append(args, "--platform", strings.Join(platforms, ","))
	}

	// Build args
	for k, v := range artifact.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	// Output as OCI tarball
	args = append(args, "--output", fmt.Sprintf("type=oci,dest=%s", outputTar))

	// Build context
	args = append(args, buildContext)

	b.logf("  Running: docker %s\n", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	if !b.opts.Quiet {
		cmd.Stdout = b.opts.Output
		cmd.Stderr = b.opts.Output
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("docker buildx build failed: %s", errMsg)
	}

	// Extract the OCI tarball to an OCI layout directory
	ociLayoutDir := filepath.Join(tmpDir, "oci-layout")
	if err := extractTar(outputTar, ociLayoutDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract OCI output: %w", err)
	}

	// Clean up the tarball (we only need the extracted layout)
	os.Remove(outputTar)

	b.logf("  Image built successfully\n")

	return ociLayoutDir, nil
}

// buildImageFallback builds a single-platform image using plain docker build + docker save,
// then converts the Docker save format to OCI layout.
// This is used when no docker-container buildx builder is available.
func (b *Builder) buildImageFallback(ctx context.Context, artifact *Artifact, dockerfile, buildContext string, platforms []string) (string, error) {
	// Generate a temporary image tag for the build
	tmpTag := fmt.Sprintf("lazyoci-build-%d:latest", os.Getpid())

	// Build the docker build command
	args := []string{"build", "-f", dockerfile, "-t", tmpTag}

	// Single platform via DOCKER_DEFAULT_PLATFORM or --platform
	if len(platforms) == 1 {
		args = append(args, "--platform", platforms[0])
	}

	// Build args
	for k, v := range artifact.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, buildContext)

	b.logf("  Running: docker %s\n", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	if !b.opts.Quiet {
		cmd.Stdout = b.opts.Output
		cmd.Stderr = b.opts.Output
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("docker build failed: %s", errMsg)
	}

	// Clean up the temporary image when done
	defer func() {
		exec.Command("docker", "rmi", tmpTag).Run()
	}()

	// Now docker save + convert to OCI layout (reuse the docker type handler logic)
	tmpDir, err := os.MkdirTemp("", "lazyoci-image-fallback-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	saveTar := filepath.Join(tmpDir, "docker-save.tar")
	var saveStderr bytes.Buffer
	saveCmd := exec.CommandContext(ctx, "docker", "save", "-o", saveTar, tmpTag)
	saveCmd.Stderr = &saveStderr

	if err := saveCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(saveStderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("docker save failed: %s", errMsg)
	}

	// Convert Docker save format to OCI layout
	ociLayoutDir := filepath.Join(tmpDir, "oci-layout")
	if err := dockerSaveToOCILayout(saveTar, ociLayoutDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to convert to OCI layout: %w", err)
	}

	os.Remove(saveTar)

	b.logf("  Image built successfully (via docker build fallback)\n")

	return ociLayoutDir, nil
}

// ensureOCICapableBuilder finds or creates a buildx builder that supports OCI export.
// Returns the builder name, or an error if none can be made available.
func ensureOCICapableBuilder(ctx context.Context, quiet bool) (string, error) {
	// Check if the "lazyoci" builder already exists
	if builderExists(ctx, lazyociBuilderName) {
		return lazyociBuilderName, nil
	}

	// Check if any existing builder uses docker-container driver
	if name, ok := findContainerBuilder(ctx); ok {
		return name, nil
	}

	// Try to create a docker-container builder
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "buildx", "create",
		"--name", lazyociBuilderName,
		"--driver", "docker-container",
	)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create buildx builder: %s", strings.TrimSpace(stderr.String()))
	}

	return lazyociBuilderName, nil
}

// builderExists checks if a named buildx builder exists.
func builderExists(ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", name)
	return cmd.Run() == nil
}

// findContainerBuilder looks for any existing buildx builder with docker-container driver.
func findContainerBuilder(ctx context.Context) (string, bool) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "buildx", "ls")
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", false
	}

	// Parse the output: each builder starts at column 0, nodes are indented.
	// Format: NAME/NODE  DRIVER/ENDPOINT  STATUS  BUILDKIT  PLATFORMS
	for _, line := range strings.Split(out.String(), "\n") {
		// Skip header, empty lines, and node lines (indented with space or \)
		if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\\") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := strings.TrimSuffix(fields[0], "*") // active builder has *
			driver := fields[1]
			if driver == "docker-container" {
				return name, true
			}
		}
	}

	return "", false
}

// checkBuildx verifies that docker buildx is available.
func checkBuildx() error {
	cmd := exec.Command("docker", "buildx", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker buildx not available: %w (install with: docker buildx install)", err)
	}
	return nil
}

// extractTar extracts a tar archive to a destination directory.
func extractTar(tarPath, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		// Sanitize: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("tar entry %q escapes destination", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
