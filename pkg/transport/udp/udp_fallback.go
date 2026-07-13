package udp

import "time"

func (l *Listener) readBatch(buf []byte) (int, error) {
	l.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	n, _, err := l.conn.ReadFromUDP(buf[:l.cfg.PacketSize])
	if err != nil {
		return 0, err
	}

	if n < l.cfg.PacketSize {
		return 0, nil
	}

	return 1, nil
}

func (l *Listener) processBatch(buf []byte, count int) {
	offset := 0
	pktSize := l.cfg.PacketSize

	for i := 0; i < count && offset+pktSize <= len(buf); i++ {
		l.processPacket(buf[offset : offset+pktSize])
		offset += pktSize
	}
}
