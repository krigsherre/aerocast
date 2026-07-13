//go:build !linux

package udp

import "time"

func (l *Listener) readAndProcessBatch(buf []byte) error {
	l.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	n, _, err := l.conn.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	if n > 0 {
		l.processPacket(buf[:n])
	}

	return nil
}
