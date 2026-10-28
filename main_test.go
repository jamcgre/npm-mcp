package main

import (
	"context"
	"errors"
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

func TestParseParts(t *testing.T) {
	tests := []struct {
		in                  string
		major, minor, patch int
		ok                  bool
	}{
		{in: "19.1.1", major: 19, minor: 1, patch: 1, ok: true},
		{in: "1.2", major: 1, minor: 2, patch: 0, ok: true},
		{in: "1", major: 1, minor: 0, patch: 0, ok: true},
		{in: "0.2.3", major: 0, minor: 2, patch: 3, ok: true},
		{in: "1.2.3.4", ok: false},
		{in: "1.2.beta", ok: false},
		{in: "", ok: false},
		{in: "^19.1.1", ok: false},
	}

	for _, test := range tests {
		major, minor, patch, ok := parseParts(test.in)
		if ok != test.ok || major != test.major || minor != test.minor || patch != test.patch {
			t.Fatalf("parseParts(%q) = %d, %d, %d, %t, want %d, %d, %d, %t", test.in, major, minor, patch, ok, test.major, test.minor, test.patch, test.ok)
		}
	}
}

func TestRangeStatusExact(t *testing.T) {
	tests := []struct {
		rng    string
		latest string
		want   string
	}{
		{rng: "1.2.3", latest: "1.2.3", want: "inside"},
		{rng: "1.0.0", latest: "1.0.0", want: "inside"},
		{rng: "1.2.3", latest: "1.2.4", want: "outside"},
		{rng: "1.2.3", latest: "1.2.2", want: "outside"},
		{rng: "1.2.3+build", latest: "1.2.3", want: "inside"},
		{rng: "1.2.3", latest: "1.2.3+sha", want: "inside"},
		{rng: "1.2.3-beta", latest: "1.2.3", want: "unparsed"},
		{rng: "1.2.3", latest: "1.2.4-beta", want: "unparsed"},
		{rng: "1.2", latest: "1.2.0", want: "unparsed"},
		{rng: "1.2.3.4", latest: "1.2.3.4", want: "unparsed"},
		{rng: "^19.1.1", latest: "19.3.1", want: "inside"},
		{rng: "^19.1.1", latest: "20.0.0", want: "outside"},
		{rng: "^1.2", latest: "1.9.0", want: "inside"},
		{rng: "^1", latest: "1.9.0", want: "inside"},
		{rng: "^0.2.3", latest: "0.2.9", want: "inside"},
		{rng: "^0.2.3", latest: "0.3.0", want: "outside"},
		{rng: "^0.2", latest: "1.0.0", want: "outside"},
		{rng: "^0.0.3", latest: "0.0.4", want: "outside"},
		{rng: "^0.0.3", latest: "0.1.0", want: "outside"},
		{rng: "^0.0", latest: "0.0.5", want: "inside"},
		{rng: "^0", latest: "0.9.0", want: "inside"},
		{rng: "^0", latest: "1.0.0", want: "outside"},
		{rng: "~1.2.3", latest: "1.2.9", want: "inside"},
		{rng: "~1.2.3", latest: "1.3.0", want: "outside"},
		{rng: "~1.2.3", latest: "2.0.0", want: "outside"},
		{rng: "~1.2", latest: "1.2.9", want: "inside"},
		{rng: "~1", latest: "1.9.0", want: "inside"},
		{rng: "~1", latest: "2.0.0", want: "outside"},
		{rng: "~0.2.3", latest: "0.2.9", want: "inside"},
		{rng: "~0", latest: "0.9.0", want: "inside"},
		{rng: "^19.1.1", latest: "19.1.1", want: "inside"},
		{rng: "^19.1.1", latest: "19.1.0", want: "outside"},
		{rng: "^1.2", latest: "1.2.0", want: "inside"},
		{rng: "^1", latest: "1.0.0", want: "inside"},
		{rng: "*", latest: "9.9.9", want: "inside"},
		{rng: "*", latest: "1.0.0-beta", want: "inside"},
	}

	for _, test := range tests {
		got := rangeStatus(test.rng, test.latest)
		if got != test.want {
			t.Fatalf("rangeStatus(%q, %q) = %s, want %s", test.rng, test.latest, got, test.want)
		}
	}
}

func TestSupportedRange(t *testing.T) {
	supported := []string{"*", "^1.2", "^19.1.1", "~1", "~1.2.3", "1.2.3"}
	for _, rng := range supported {
		if !supportedRange(rng) {
			t.Fatalf("supportedRange(%q) = false", rng)
		}
	}

	unsupported := []string{"1.2", "foo.bar.baz", "1.2.3-beta", ">=1.2.3", "1.2.3.4"}
	for _, rng := range unsupported {
		if supportedRange(rng) {
			t.Fatalf("supportedRange(%q) = true", rng)
		}
	}
}

func withLookup(t *testing.T, fn func(context.Context, string) (string, error)) {
	t.Helper()
	original := lookupLatest
	t.Cleanup(func() { lookupLatest = original })
	lookupLatest = fn
}

func TestDependencyFreshness(t *testing.T) {
	withLookup(t, func(ctx context.Context, name string) (string, error) {
		switch name {
		case "express":
			return "5.0.0", nil
		case "left-pad":
			return "1.0.0", nil
		case "missing-pkg":
			return "", errorPackageMissing
		case "react":
			return "19.3.1", nil
		case "typescript":
			return "5.9.3", nil
		default:
			t.Fatalf("unexpected lookup %s", name)
			return "", nil
		}
	})

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{
		"name": "root",
		"workspaces": ["client"],
		"dependencies": {"left-pad": "1.0.0"}
	}`)
	writeFile(t, filepath.Join(dir, "client", "package.json"), `{
		"name": "client",
		"dependencies": {
			"express": "^4.21.2",
			"missing-pkg": "^1.0.0",
			"react": "^19.1.1",
			"star": "*",
			"weird": "git:somewhere"
		},
		"devDependencies": {"typescript": "~5.9.2"},
		"peerDependencies": {"react": "^19.1.1"}
	}`)

	clientPath := filepath.Join(dir, "client", "package.json")
	rootPath := filepath.Join(dir, "package.json")
	_, got, err := dependencyFreshness(context.Background(), nil, dependencyFreshnessArgs{Directory: dir})
	if err != nil {
		t.Fatal(err)
	}

	want := []freshnessRow{
		{Path: clientPath, Name: "express", Source: "dependencies", Range: "^4.21.2", Latest: "5.0.0", Status: "outside"},
		{Path: clientPath, Name: "missing-pkg", Source: "dependencies", Range: "^1.0.0", Status: "missing"},
		{Path: clientPath, Name: "react", Source: "dependencies", Range: "^19.1.1", Latest: "19.3.1", Status: "inside"},
		{Path: clientPath, Name: "star", Source: "dependencies", Range: "*", Status: "inside"},
		{Path: clientPath, Name: "weird", Source: "dependencies", Range: "git:somewhere", Status: "unparsed"},
		{Path: clientPath, Name: "typescript", Source: "devDependencies", Range: "~5.9.2", Latest: "5.9.3", Status: "inside"},
		{Path: clientPath, Name: "react", Source: "peerDependencies", Range: "^19.1.1", Latest: "19.3.1", Status: "inside"},
		{Path: rootPath, Name: "left-pad", Source: "dependencies", Range: "1.0.0", Latest: "1.0.0", Status: "inside"},
	}
	if len(got.Rows) != len(want) {
		t.Fatalf("rows %#v", got.Rows)
	}
	for i := range want {
		if got.Rows[i] != want[i] {
			t.Fatalf("row %d %#v, want %#v", i, got.Rows[i], want[i])
		}
	}
}

func TestDependencyFreshnessEmpty(t *testing.T) {
	withLookup(t, func(ctx context.Context, name string) (string, error) {
		t.Fatalf("unexpected lookup %s", name)
		return "", nil
	})

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"name":"app","version":"1.0.0"}`)

	_, got, err := dependencyFreshness(context.Background(), nil, dependencyFreshnessArgs{Directory: dir})
	if err != nil {
		t.Fatal(err)
	}
	if got.Rows == nil || len(got.Rows) != 0 {
		t.Fatalf("rows %#v", got.Rows)
	}
}

func TestDependencyFreshnessLookupError(t *testing.T) {
	withLookup(t, func(ctx context.Context, name string) (string, error) {
		return "", errors.New("registry down")
	})

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{"dependencies":{"react":"^19.1.1"}}`)

	_, _, err := dependencyFreshness(context.Background(), nil, dependencyFreshnessArgs{Directory: dir})
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, errorPackageMissing) {
		t.Fatal("missing package")
	}
}
