package version

import (
	"strings"
	"testing"
)

func TestCadenaIncluyeVersionYCommit(t *testing.T) {
	cadena := Cadena()
	if !strings.Contains(cadena, Version) {
		t.Errorf("Cadena() = %q; debería contener la versión %q", cadena, Version)
	}
	if !strings.Contains(cadena, Commit) {
		t.Errorf("Cadena() = %q; debería contener el commit %q", cadena, Commit)
	}
}

func TestMetadatosNoVacios(t *testing.T) {
	if Repositorio == "" || URLInstaladorLinux == "" || URLInstaladorWindows == "" {
		t.Error("los metadatos del proyecto no deberían estar vacíos")
	}
}
