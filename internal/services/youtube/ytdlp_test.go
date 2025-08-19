package youtube

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"yogo/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("GO_TEST_MODE_YTDLP") == "1" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_STDOUT"))
		fmt.Fprint(os.Stderr, os.Getenv("MOCK_STDERR"))
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func mockExecCommand(t *testing.T, stdout, stderr string) {
	originalExecCommand := execCommand
	t.Cleanup(func() {
		execCommand = originalExecCommand
	})

	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestMain")
		cmd.Env = []string{
			"GO_TEST_MODE_YTDLP=1",
			"MOCK_STDOUT=" + stdout,
			"MOCK_STDERR=" + stderr,
		}
		return cmd
	}
}

func TestYTDLPClient_Search(t *testing.T) {
	testCases := []struct {
		name          string
		mockStdout    string
		mockStderr    string
		expectErr     bool
		expectedSongs []domain.Song
	}{
		{
			name: "Successful search",
			mockStdout: `{
				"entries": [
					{"id": "id1", "title": "Song One", "uploader": "Artist A"},
					{"id": "id2", "title": "Song Two", "channel": "Artist B"}
				]
			}`,
			mockStderr: "",
			expectErr:  false,
			expectedSongs: []domain.Song{
				{ID: "id1", Title: "Song One", Artists: []string{"Artist A"}},
				{ID: "id2", Title: "Song Two", Artists: []string{"Artist B"}},
			},
		},
		{
			name:       "yt-dlp returns an error",
			mockStdout: "",
			mockStderr: "ERROR: This video is unavailable.",
			expectErr:  true,
		},
		{
			name:       "Invalid JSON output",
			mockStdout: `{"entries": [`,
			mockStderr: "",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExecCommand(t, tc.mockStdout, tc.mockStderr)
			client := NewYTDLPClient("")

			songs, err := client.Search("test query")

			if tc.expectErr {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Expected no error but got one")
				assert.Equal(t, tc.expectedSongs, songs, "The returned songs do not match the expected songs")
			}
		})
	}
}

func TestYTDLPClient_GetStreamURL(t *testing.T) {
	testCases := []struct {
		name        string
		mockStdout  string
		mockStderr  string
		expectErr   bool
		expectedURL string
	}{
		{
			name:        "Successful URL fetch",
			mockStdout:  "https://example.com/stream.m4a",
			mockStderr:  "",
			expectErr:   false,
			expectedURL: "https://example.com/stream.m4a",
		},
		{
			name:       "yt-dlp returns an error",
			mockStdout: "",
			mockStderr: "ERROR: This video is private.",
			expectErr:  true,
		},
		{
			name:       "Empty output from yt-dlp",
			mockStdout: "",
			mockStderr: "",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExecCommand(t, tc.mockStdout, tc.mockStderr)
			client := NewYTDLPClient("")

			url, err := client.GetStreamURL("test_id")

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedURL, url)
			}
		})
	}
}
