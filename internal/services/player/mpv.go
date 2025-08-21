package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
	"yogo/internal/logger"
	"yogo/internal/ports"
)

const (
	socketCheckRetries   = 20
	socketCheckInterval  = 100 * time.Millisecond
	socketReadDeadline   = 500 * time.Millisecond
	mpvCommandReqIDPause = 1
	mpvCommandReqIDPos   = 2
	mpvCommandReqIDDur   = 3
)

type MpvCommand struct {
	Command   []any `json:"command"`
	RequestID int   `json:"request_id,omitempty"`
}

type MpvResponse struct {
	Error     string `json:"error"`
	Data      any    `json:"data"`
	RequestID int    `json:"request_id"`
	Event     string `json:"event"`
}

type MpvPlayer struct {
	socketPath string
	cmd        *exec.Cmd
	mu         sync.Mutex
}

func NewMpvPlayer(socketPath string) ports.PlayerService {
	os.Remove(socketPath)
	return &MpvPlayer{socketPath: socketPath}
}

func (p *MpvPlayer) isProcessRunning() bool {
	return p.cmd != nil && p.cmd.Process != nil
}

func (p *MpvPlayer) startMpvProcess() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isProcessRunning() {
		if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
			p.cmd = nil
		} else {
			return nil
		}
	}

	logger.Log.Info().Msg("Starting new mpv process...")
	args := []string{
		"--idle",
		"--input-ipc-server=" + p.socketPath,
		"--no-video",
	}
	p.cmd = exec.Command("mpv", args...)
	p.cmd.Stdout = logger.Log
	p.cmd.Stderr = logger.Log

	if err := p.cmd.Start(); err != nil {
		p.cmd = nil
		return fmt.Errorf("could not start mpv process: %w", err)
	}

	for range socketCheckRetries {
		if _, err := os.Stat(p.socketPath); err == nil {
			logger.Log.Info().Msg("mpv socket detected. Process ready.")
			return nil
		}
		time.Sleep(socketCheckInterval)
	}

	logger.Log.Error().Str("socket", p.socketPath).Msg("Timed out waiting for mpv socket.")
	p.cmd.Process.Kill()
	p.cmd = nil
	return fmt.Errorf("mpv process started but socket did not appear at %s", p.socketPath)
}

func (p *MpvPlayer) sendCommands(cmds ...MpvCommand) ([]MpvResponse, error) {
	conn, err := net.Dial("unix", p.socketPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to mpv socket: %w", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(socketReadDeadline))

	encoder := json.NewEncoder(conn)
	for _, cmd := range cmds {
		if err := encoder.Encode(cmd); err != nil {
			return nil, fmt.Errorf("error sending mpv command: %w", err)
		}
	}

	var responses []MpvResponse
	scanner := bufio.NewScanner(conn)
	for len(responses) < len(cmds) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				logger.Log.Error().Err(err).Msg("Error reading from mpv socket")
			}
			break
		}

		line := scanner.Bytes()
		var resp MpvResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			logger.Log.Warn().Str("line", string(line)).Err(err).Msg("Could not parse line from mpv")
			continue
		}

		if resp.Event == "" && resp.RequestID > 0 {
			responses = append(responses, resp)
			logger.Log.Debug().RawJSON("response", line).Msg("Processed mpv command response")
		} else {
			logger.Log.Debug().RawJSON("event", line).Msg("Ignored mpv event")
		}
	}
	return responses, nil
}

func (p *MpvPlayer) Play(url string) error {
	if err := p.startMpvProcess(); err != nil {
		return err
	}

	disableVideoCmd := MpvCommand{Command: []any{"set_property", "vid", "no"}}
	loadFileCmd := MpvCommand{Command: []any{"loadfile", url, "replace"}}
	_, err := p.sendCommands(disableVideoCmd, loadFileCmd)
	return err
}

func (p *MpvPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isProcessRunning() {
		return nil
	}
	cmd := MpvCommand{Command: []any{"cycle", "pause"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isProcessRunning() {
		return nil
	}
	cmd := MpvCommand{Command: []any{"stop"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) Seek(seconds int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isProcessRunning() {
		return nil
	}
	cmd := MpvCommand{Command: []any{"seek", seconds, "relative"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) GetState() (ports.PlayerState, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	state := ports.PlayerState{}
	if !p.isProcessRunning() {
		return state, nil
	}

	pauseCmd := MpvCommand{Command: []any{"get_property", "pause"}, RequestID: mpvCommandReqIDPause}
	posCmd := MpvCommand{Command: []any{"get_property", "time-pos"}, RequestID: mpvCommandReqIDPos}
	durCmd := MpvCommand{Command: []any{"get_property", "duration"}, RequestID: mpvCommandReqIDDur}

	responses, err := p.sendCommands(pauseCmd, posCmd, durCmd)
	if err != nil {
		return state, err
	}

	if len(responses) < 3 {
		logger.Log.Warn().Int("received_count", len(responses)).Int("expected_count", 3).Msg("Unexpected number of responses from GetState")
	}

	for _, resp := range responses {
		if resp.Error != "success" {
			continue
		}
		switch resp.RequestID {
		case mpvCommandReqIDPause:
			if isPaused, ok := resp.Data.(bool); ok {
				state.IsPlaying = !isPaused
			}
		case mpvCommandReqIDPos:
			if pos, ok := resp.Data.(float64); ok {
				state.Position = pos
			}
		case mpvCommandReqIDDur:
			if dur, ok := resp.Data.(float64); ok {
				state.Duration = dur
			}
		}
	}
	return state, nil
}

func (p *MpvPlayer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	logger.Log.Info().Msg("Closing player service...")
	if p.isProcessRunning() {
		logger.Log.Info().Msg("Terminating mpv process...")
		if err := p.cmd.Process.Kill(); err != nil {
			logger.Log.Error().Err(err).Msg("Error terminating mpv process")
		}
	}
	os.Remove(p.socketPath)
	logger.Log.Info().Msg("Player service closed.")
	return nil
}
