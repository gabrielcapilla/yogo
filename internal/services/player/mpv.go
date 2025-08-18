package player

import (
	"encoding/json"
	"fmt"
	"net"

	"yogo/internal/ports"
)

type MpvCommand struct {
	Command []interface{} `json:"command"`
}

type MpvPlayer struct {
	socketPath string
}

func NewMpvPlayer(socketPath string) ports.PlayerService {
	return &MpvPlayer{socketPath: socketPath}
}

func (p *MpvPlayer) sendCommand(cmd MpvCommand) error {
	conn, err := net.Dial("unix", p.socketPath)
	if err != nil {
		return fmt.Errorf("no se pudo conectar al socket de mpv en %s. ¿Está mpv ejecutándose con --input-ipc-server=%s? Error: %w", p.socketPath, p.socketPath, err)
	}
	defer conn.Close()

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("error al codificar el comando mpv a JSON: %w", err)
	}

	_, err = conn.Write(append(cmdBytes, '\n'))
	if err != nil {
		return fmt.Errorf("error al enviar el comando al socket de mpv: %w", err)
	}

	return nil
}

func (p *MpvPlayer) Play(url string) error {
	cmd := MpvCommand{
		Command: []interface{}{"loadfile", url},
	}
	return p.sendCommand(cmd)
}

func (p *MpvPlayer) Pause() error {
	cmd := MpvCommand{
		Command: []interface{}{"cycle", "pause"},
	}
	return p.sendCommand(cmd)
}

func (p *MpvPlayer) Stop() error {
	cmd := MpvCommand{
		Command: []interface{}{"stop"},
	}
	return p.sendCommand(cmd)
}
