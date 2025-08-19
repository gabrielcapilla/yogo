package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
	"yogo/internal/logger"
	"yogo/internal/ports"
)

type MpvCommand struct {
	Command   []interface{} `json:"command"`
	RequestID int           `json:"request_id,omitempty"`
}

type MpvResponse struct {
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
	RequestID int         `json:"request_id"`
	Event     string      `json:"event"`
}

type MpvPlayer struct {
	socketPath string
	cmd        *exec.Cmd
}

func NewMpvPlayer(socketPath string) ports.PlayerService {
	os.Remove(socketPath)
	return &MpvPlayer{socketPath: socketPath}
}

func (p *MpvPlayer) startMpvProcess() error {
	if p.cmd != nil && p.cmd.Process != nil {
		if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
			p.cmd = nil
		} else {
			return nil
		}
	}

	logger.Log.Println("Starting new mpv process...")
	args := []string{
		"--idle",
		"--input-ipc-server=" + p.socketPath,
		"--no-video",
	}
	p.cmd = exec.Command("mpv", args...)

	p.cmd.Stdout = logger.Log.Writer()
	p.cmd.Stderr = logger.Log.Writer()

	if err := p.cmd.Start(); err != nil {
		p.cmd = nil
		return fmt.Errorf("could not start mpv process: %w", err)
	}

	for i := 0; i < 20; i++ {
		if _, err := os.Stat(p.socketPath); err == nil {
			logger.Log.Println("mpv socket detected. Process ready.")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

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

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

	encoder := json.NewEncoder(conn)
	for _, cmd := range cmds {
		if err := encoder.Encode(cmd); err != nil {
			return nil, fmt.Errorf("error sending mpv command: %w", err)
		}
	}

	var responses []MpvResponse
	scanner := bufio.NewScanner(conn)

	for len(responses) < len(cmds) {
		if scanner.Scan() {
			line := scanner.Bytes()
			var resp MpvResponse
			if err := json.Unmarshal(line, &resp); err == nil {
				if resp.Event == "" && resp.RequestID > 0 {
					responses = append(responses, resp)
					logger.Log.Printf("Processed mpv command response: %s", string(line))
				} else {
					logger.Log.Printf("Ignored mpv event: %s", string(line))
				}
			} else {
				logger.Log.Printf("Could not parse line from mpv: %s", string(line))
			}
		} else {
			if err := scanner.Err(); err != nil {
				logger.Log.Printf("Error reading from mpv socket: %v", err)
			}
			break
		}
	}
	return responses, nil
}

func (p *MpvPlayer) Play(url string) error {
	if err := p.startMpvProcess(); err != nil {
		return err
	}

	disableVideoCmd := MpvCommand{Command: []interface{}{"set_property", "vid", "no"}}
	loadFileCmd := MpvCommand{Command: []interface{}{"loadfile", url, "replace"}}
	_, err := p.sendCommands(disableVideoCmd, loadFileCmd)
	return err
}

func (p *MpvPlayer) Pause() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	cmd := MpvCommand{Command: []interface{}{"cycle", "pause"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	cmd := MpvCommand{Command: []interface{}{"stop"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) Seek(seconds int) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	cmd := MpvCommand{Command: []interface{}{"seek", seconds, "relative"}}
	_, err := p.sendCommands(cmd)
	return err
}

func (p *MpvPlayer) GetState() (ports.PlayerState, error) {
	state := ports.PlayerState{}
	if p.cmd == nil || p.cmd.Process == nil {
		return state, nil
	}

	pauseCmd := MpvCommand{Command: []interface{}{"get_property", "pause"}, RequestID: 1}
	posCmd := MpvCommand{Command: []interface{}{"get_property", "time-pos"}, RequestID: 2}
	durCmd := MpvCommand{Command: []interface{}{"get_property", "duration"}, RequestID: 3}

	responses, err := p.sendCommands(pauseCmd, posCmd, durCmd)
	if err != nil {
		return state, err
	}

	if len(responses) < 3 {
		logger.Log.Printf("Expected 3 responses from GetState, but received %d", len(responses))
	}

	for _, resp := range responses {
		if resp.Error != "success" {
			continue
		}
		switch resp.RequestID {
		case 1:
			if isPaused, ok := resp.Data.(bool); ok {
				state.IsPlaying = !isPaused
			}
		case 2:
			if pos, ok := resp.Data.(float64); ok {
				state.Position = pos
			}
		case 3:
			if dur, ok := resp.Data.(float64); ok {
				state.Duration = dur
			}
		}
	}
	return state, nil
}

func (p *MpvPlayer) Close() error {
	logger.Log.Println("Closing player service...")
	if p.cmd != nil && p.cmd.Process != nil {
		logger.Log.Println("Terminating mpv process...")
		if err := p.cmd.Process.Kill(); err != nil {
			logger.Log.Printf("Error terminating mpv process: %v", err)
			return err
		}
	}
	os.Remove(p.socketPath)
	logger.Log.Println("Player service closed.")
	return nil
}
