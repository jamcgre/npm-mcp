package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func manifestFromJSON(t *testing.T, contents string) packageManifestResult {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return manifestAt(t, path)
}

func manifestAt(t *testing.T, path string) packageManifestResult {
	t.Helper()
	_, got, err := packageManifest(context.Background(), nil, packageManifestArgs{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func requireEmptyMap(t *testing.T, name string, values map[string]string) {
	t.Helper()
	if values == nil {
		t.Fatalf("%s is nil", name)
	}
	if len(values) != 0 {
		t.Fatalf("%s %#v", name, values)
	}
}

func TestStringLicense(t *testing.T) {
	got := manifestFromJSON(t, `{
		"name": "left-pad",
		"version": "1.0.0",
		"description": "pad a string",
		"license": "(MIT OR Apache-2.0)",
		"dependencies": {"react": "^19.0.0"},
		"peerDependencies": {"react": "^19.0.0"},
		"scripts": {"test": "vitest"}
	}`)

	if got.Name != "left-pad" {
		t.Fatalf("name %s", got.Name)
	}
	if got.Version != "1.0.0" {
		t.Fatalf("version %s", got.Version)
	}
	if got.Description != "pad a string" {
		t.Fatalf("description %s", got.Description)
	}
	if got.Private {
		t.Fatal("private")
	}
	if got.License.Type != "(MIT OR Apache-2.0)" {
		t.Fatalf("license type %s", got.License.Type)
	}
	if got.License.URL != "" {
		t.Fatalf("license url %s", got.License.URL)
	}
	if got.Dependencies["react"] != "^19.0.0" {
		t.Fatalf("dependencies %#v", got.Dependencies)
	}
	if got.PeerDependencies["react"] != "^19.0.0" {
		t.Fatalf("peerDependencies %#v", got.PeerDependencies)
	}
	if got.Scripts["test"] != "vitest" {
		t.Fatalf("scripts %#v", got.Scripts)
	}
	requireEmptyMap(t, "devDependencies", got.DevDependencies)
	if got.Workspaces == nil || len(got.Workspaces) != 0 {
		t.Fatalf("workspaces %#v", got.Workspaces)
	}
}

func TestObjectLicense(t *testing.T) {
	got := manifestFromJSON(t, `{
		"name": "pkg",
		"license": {"type": "MIT", "url": "https://example.com/license"}
	}`)

	if got.License.Type != "MIT" {
		t.Fatalf("license type %s", got.License.Type)
	}
	if got.License.URL != "https://example.com/license" {
		t.Fatalf("license url %s", got.License.URL)
	}
}

func TestMissingVersionLicenseAndWorkspaces(t *testing.T) {
	got := manifestFromJSON(t, `{
		"name": "app",
		"private": true
	}`)

	if got.Name != "app" {
		t.Fatalf("name %s", got.Name)
	}
	if !got.Private {
		t.Fatal("private")
	}
	if got.Version != "" {
		t.Fatalf("version %s", got.Version)
	}
	if got.License.Type != "" || got.License.URL != "" {
		t.Fatalf("license %#v", got.License)
	}
	if got.Workspaces == nil || len(got.Workspaces) != 0 {
		t.Fatalf("workspaces %#v", got.Workspaces)
	}
	requireEmptyMap(t, "dependencies", got.Dependencies)
	requireEmptyMap(t, "devDependencies", got.DevDependencies)
	requireEmptyMap(t, "peerDependencies", got.PeerDependencies)
	requireEmptyMap(t, "scripts", got.Scripts)
}

func TestWorkspaceGlob(t *testing.T) {
	got := manifestFromJSON(t, `{
		"name": "monorepo",
		"workspaces": ["packages/*", "apps/*"]
	}`)

	if len(got.Workspaces) != 2 || got.Workspaces[0] != "packages/*" || got.Workspaces[1] != "apps/*" {
		t.Fatalf("workspaces %#v", got.Workspaces)
	}
}

func TestInvalidLicense(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(path, []byte(`{"name":"pkg","license":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	_, _, err := packageManifest(context.Background(), nil, packageManifestArgs{Path: path})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestClemsonOffenseRoot(t *testing.T) {
	path := filepath.Join("..", "clemson-offense", "package.json")
	if _, err := os.Stat(path); err != nil {
		t.Skip("clemson-offense is not beside this repo")
	}
	got := manifestAt(t, path)

	if got.Name != "clemson-offense" {
		t.Fatalf("name %s", got.Name)
	}
	if !got.Private {
		t.Fatal("private")
	}
	if got.Version != "" {
		t.Fatalf("version %s", got.Version)
	}
	if len(got.Workspaces) != 2 || got.Workspaces[0] != "client" || got.Workspaces[1] != "server" {
		t.Fatalf("workspaces %#v", got.Workspaces)
	}
	if got.Scripts["test"] != "npm test -w server" {
		t.Fatalf("scripts %#v", got.Scripts)
	}
	requireEmptyMap(t, "dependencies", got.Dependencies)
	requireEmptyMap(t, "devDependencies", got.DevDependencies)
	requireEmptyMap(t, "peerDependencies", got.PeerDependencies)
}

func TestClemsonOffenseClient(t *testing.T) {
	path := filepath.Join("..", "clemson-offense", "client", "package.json")
	if _, err := os.Stat(path); err != nil {
		t.Skip("clemson-offense is not beside this repo")
	}
	got := manifestAt(t, path)

	if got.Name != "client" {
		t.Fatalf("name %s", got.Name)
	}
	if !got.Private {
		t.Fatal("private")
	}
	if got.Dependencies["react"] != "^19.1.1" {
		t.Fatalf("dependencies %#v", got.Dependencies)
	}
	if got.DevDependencies["vite"] != "^7.1.10" {
		t.Fatalf("devDependencies %#v", got.DevDependencies)
	}
	requireEmptyMap(t, "peerDependencies", got.PeerDependencies)
	if got.Workspaces == nil || len(got.Workspaces) != 0 {
		t.Fatalf("workspaces %#v", got.Workspaces)
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func projectAt(t *testing.T, directory string) projectManifestResult {
	t.Helper()
	_, got, err := projectManifest(context.Background(), nil, projectManifestArgs{Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestProjectManifestSinglePackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"name":"app","version":"1.0.0"}`)

	got := projectAt(t, dir)
	if len(got.Packages) != 1 {
		t.Fatalf("packages %#v", got.Packages)
	}
	if got.Packages[0].Path != filepath.Join(dir, "package.json") {
		t.Fatalf("path %s", got.Packages[0].Path)
	}
	if got.Packages[0].Manifest.Name != "app" {
		t.Fatalf("name %s", got.Packages[0].Manifest.Name)
	}
}

func TestProjectManifestGlob(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{
		"name": "monorepo",
		"workspaces": ["packages/*"]
	}`)
	writeFile(t, filepath.Join(dir, "packages", "app", "package.json"), `{"name":"app","dependencies":{"react":"^19.0.0"}}`)
	writeFile(t, filepath.Join(dir, "packages", "lib", "package.json"), `{"name":"lib"}`)
	writeFile(t, filepath.Join(dir, "packages", "empty", "README.md"), "skip")
	writeFile(t, filepath.Join(dir, "packages", "notes.txt"), "skip")

	got := projectAt(t, dir)
	if len(got.Packages) != 3 {
		t.Fatalf("packages %#v", got.Packages)
	}

	root := got.Packages[0]
	if root.Path != filepath.Join(dir, "package.json") || root.Manifest.Name != "monorepo" {
		t.Fatalf("root %#v", root)
	}

	app := got.Packages[1]
	if app.Path != filepath.Join(dir, "packages", "app", "package.json") {
		t.Fatalf("app path %s", app.Path)
	}
	if app.Manifest.Dependencies["react"] != "^19.0.0" {
		t.Fatalf("app dependencies %#v", app.Manifest.Dependencies)
	}

	lib := got.Packages[2]
	if lib.Path != filepath.Join(dir, "packages", "lib", "package.json") || lib.Manifest.Name != "lib" {
		t.Fatalf("lib %#v", lib)
	}
}

func TestProjectManifestClemsonOffense(t *testing.T) {
	dir := filepath.Join("..", "clemson-offense")
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		t.Skip("clemson-offense is not beside this repo")
	}

	got := projectAt(t, dir)
	if len(got.Packages) != 3 {
		t.Fatalf("packages %#v", got.Packages)
	}

	if got.Packages[0].Path != filepath.Join(dir, "package.json") || got.Packages[0].Manifest.Name != "clemson-offense" {
		t.Fatalf("root %#v", got.Packages[0])
	}
	if got.Packages[1].Path != filepath.Join(dir, "client", "package.json") || got.Packages[1].Manifest.Dependencies["react"] != "^19.1.1" {
		t.Fatalf("client %#v", got.Packages[1])
	}
	if got.Packages[2].Path != filepath.Join(dir, "server", "package.json") || got.Packages[2].Manifest.Dependencies["express"] != "^4.21.2" {
		t.Fatalf("server %#v", got.Packages[2])
	}
}

func TestProjectManifestMissingDirectory(t *testing.T) {
	_, _, err := projectManifest(context.Background(), nil, projectManifestArgs{Directory: filepath.Join(t.TempDir(), "missing")})
	if err == nil {
		t.Fatal("expected an error")
	}
}
