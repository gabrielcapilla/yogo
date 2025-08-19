package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"yogo/internal/logger"
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

func (p *MpvPlayer) sendCommands(cmds ...MpvCommand) error {
	conn, err := net.Dial("unix", p.socketPath)
	if err != nil {
		return fmt.Errorf("no se pudo conectar al socket de mpv en %s. ¿Está mpv ejecutándose con --input-ipc-server=%s? Error: %w", p.socketPath, p.socketPath, err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	for _, cmd := range cmds {
		if err := encoder.Encode(cmd); err != nil {
			return fmt.Errorf("error al codificar o enviar el comando mpv a JSON: %w", err)
		}
	}

	reader := bufio.NewReader(conn)
	for i := 0; i < len(cmds); i++ {
		response, err := reader.ReadString('\n')
		if err != nil {
			logger.Log.Printf("Error menor al leer la respuesta de mpv (puede ser ignorado): %v", err)
			break
		}
		logger.Log.Printf("Respuesta recibida de mpv: %s", response)
	}

	return nil
}

func (p *MpvPlayer) Play(url string) error {
	disableVideoCmd := MpvCommand{
		Command: []interface{}{"set_property", "vid", "no"},
	}
	loadFileCmd := MpvCommand{
		Command: []interface{}{"loadfile", url, "replace"},
	}
	return p.sendCommands(disableVideoCmd, loadFileCmd)
}

func (p *MpvPlayer) Pause() error {
	cmd := MpvCommand{
		Command: []interface{}{"cycle", "pause"},
	}
	return p.sendCommands(cmd)
}

func (p *MpvPlayer) Stop() error {
	cmd := MpvCommand{
		Command: []interface{}{"stop"},
	}
	return p.sendCommands(cmd)
}
