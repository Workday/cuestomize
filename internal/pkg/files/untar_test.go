package files

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUntar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		archiveName  string
		archiveFiles []archiveFile
		wantFiles    map[string]string
		wantErr      bool
		opts         []Option
	}{
		{
			name:        "extracts regular files from tar archive",
			archiveName: "module.tar",
			archiveFiles: []archiveFile{
				{name: "cue.mod/module.cue", body: "module: \"example.com/test@v0\"\n"},
				{name: "schema.cue", body: "package model\n"},
			},
			wantFiles: map[string]string{
				"cue.mod/module.cue": "module: \"example.com/test@v0\"\n",
				"schema.cue":         "package model\n",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			t.Chdir(workDir)

			destDir := t.TempDir()

			archivePath := filepath.Join(workDir, tt.archiveName)
			writeTarArchive(t, archivePath, tt.archiveFiles)

			err := Untar(archivePath, destDir, tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			for relPath, want := range tt.wantFiles {
				got, err := os.ReadFile(filepath.Join(destDir, relPath))
				require.NoError(t, err)
				require.Equal(t, want, string(got))
			}
		})
	}
}

type archiveFile struct {
	name string
	body string
}

func writeTarArchive(t *testing.T, path string, files []archiveFile) {
	t.Helper()

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: 0o644,
			Size: int64(len(file.body)),
		}
		err := tw.WriteHeader(hdr)
		require.NoError(t, err)
		_, err = tw.Write([]byte(file.body))
		require.NoError(t, err)
	}

	err = tw.Close()
	require.NoError(t, err)
}
