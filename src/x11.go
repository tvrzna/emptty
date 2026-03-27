package src

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Opens XDisplay via socket and tries to authenticate with Xauthority.
func openXDisplay(dispName, xauthorityPath string) (net.Conn, error) {
	displayNum, socketPath, err := resolveDisplay(dispName)
	if err != nil {
		return nil, err
	}

	authName, authData, err := readXAuthority(displayNum, xauthorityPath)
	if err != nil {
		return nil, fmt.Errorf("xauthority: %w", err)
	}

	counter := 0
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			err = performHandshake(conn, authName, authData)
			if err == nil {
				return conn, nil
			}
			logPrintf("X11 handshake[#%d] failed: %v", counter, err)
			conn.Close()
		}
		time.Sleep(50 * time.Millisecond)
		counter++
	}

	return nil, errors.New("could not open authenticated X display")
}

// Gets display number and its socket path from display name.
func resolveDisplay(display string) (string, string, error) {
	if !strings.HasPrefix(display, ":") {
		return "", "", errors.New("only local displays supported")
	}
	num := strings.Split(display[1:], ".")[0]
	socketPath := filepath.Join("/tmp/.X11-unix", "X"+num)
	return num, socketPath, nil
}

// Reads and parses XAuthority file on defined path.
func readXAuthority(displayNum, xauthorityPath string) (string, []byte, error) {
	data, err := os.ReadFile(xauthorityPath)
	if err != nil {
		return "", nil, err
	}

	r := bytes.NewReader(data)

	for r.Len() > 0 {
		var family uint16
		if err := binary.Read(r, binary.BigEndian, &family); err != nil {
			return "", nil, err
		}

		addr, err := readString(r)
		if err != nil {
			return "", nil, err
		}
		_ = addr

		display, err := readString(r)
		if err != nil {
			return "", nil, err
		}

		name, err := readString(r)
		if err != nil {
			return "", nil, err
		}

		authData, err := readString(r)
		if err != nil {
			return "", nil, err
		}

		if (family == 0 || family == 256) &&
			display == displayNum &&
			name == "MIT-MAGIC-COOKIE-1" {
			return name, []byte(authData), nil
		}
	}

	return "", nil, errors.New("no matching MIT-MAGIC-COOKIE-1 entry found")
}

// Reads strings from reader by defined length.
func readString(r *bytes.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// Performs handshake authorization with opened X display.
func performHandshake(conn net.Conn, authName string, authData []byte) error {
	authNameBytes := []byte(authName)

	pad := func(n int) int {
		if n%4 == 0 {
			return n
		}
		return n + (4 - (n % 4))
	}

	authNameLen := len(authNameBytes)
	authDataLen := len(authData)

	packetLen := 12 +
		pad(authNameLen) +
		pad(authDataLen)

	buf := make([]byte, packetLen)

	buf[0] = 'l'
	binary.LittleEndian.PutUint16(buf[2:], 11)
	binary.LittleEndian.PutUint16(buf[4:], 0)
	binary.LittleEndian.PutUint16(buf[6:], uint16(authNameLen))
	binary.LittleEndian.PutUint16(buf[8:], uint16(authDataLen))

	offset := 12
	copy(buf[offset:], authNameBytes)
	offset += pad(authNameLen)

	copy(buf[offset:], authData)

	if _, err := conn.Write(buf); err != nil {
		return err
	}

	reply := make([]byte, 8)
	if _, err := conn.Read(reply); err != nil {
		return err
	}

	if reply[0] != 1 {
		return errors.New("X server authentication failed")
	}

	return nil
}
