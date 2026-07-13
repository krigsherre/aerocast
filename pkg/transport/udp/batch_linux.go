//go:build linux

package udp

import (
	"syscall"
	"time"
	"unsafe"
)

type mmsghdr struct {
	Msg       syscall.Msghdr
	Len       uint32
	Pad_cgo_0 [4]byte
}

func (l *Listener) readAndProcessBatch(buf []byte) error {
	fd, err := l.conn.File()
	if err != nil {
		return err
	}
	rawFD := int(fd.Fd())
	fd.Close()

	batchSize := l.cfg.BatchSize
	slotSize := len(buf) / batchSize

	hdrs := make([]mmsghdr, batchSize)
	iovs := make([]syscall.Iovec, batchSize)

	for i := 0; i < batchSize; i++ {
		offset := i * slotSize
		iovs[i].Base = (*byte)(unsafe.Pointer(&buf[offset]))
		iovs[i].SetLen(slotSize)

		hdrs[i].Msg.Iov = &iovs[i]
		hdrs[i].Msg.Iovlen = 1
	}

	timeout := syscall.NsecToTimeval(int64(100 * time.Millisecond))

	n, _, errno := syscall.Syscall6(
		syscall.SYS_RECVMMSG,
		uintptr(rawFD),
		uintptr(unsafe.Pointer(&hdrs[0])),
		uintptr(batchSize),
		uintptr(syscall.MSG_WAITFORONE),
		uintptr(unsafe.Pointer(&timeout)),
		0,
	)

	if errno != 0 {
		if errno == syscall.EAGAIN || errno == syscall.EWOULDBLOCK {
			return nil
		}
		return errno
	}

	for i := 0; i < int(n); i++ {
		pktLen := int(hdrs[i].Len)
		if pktLen > 0 {
			offset := i * slotSize
			l.processPacket(buf[offset : offset+pktLen])
		}
	}

	return nil
}
