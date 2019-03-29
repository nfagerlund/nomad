// +build !windows

package fifo

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// New creates a fifo at the given path and returns an io.ReadWriteCloser for it
// The fifo must not already exist
func New(path string) (io.ReadWriteCloser, error) {
	return openFifo(context.Background(), path, syscall.O_RDONLY|syscall.O_CREAT|syscall.O_NONBLOCK, 0600)
}

// Open opens a fifo that already exists and returns an io.ReadWriteCloser for it
func Open(path string) (io.ReadWriteCloser, error) {
	return openFifo(context.Background(), path, syscall.O_WRONLY, 0600)
}

// Remove a fifo that already exists at a given path
func Remove(path string) error {
	return os.Remove(path)
}

func IsClosedErr(err error) bool {
	err2, ok := err.(*os.PathError)
	if ok {
		return err2.Err == os.ErrClosed
	}
	return false
}

// openFifo opens a fifo. Returns io.ReadWriteCloser.
// Context can be used to cancel this function until open(2) has not returned.
// Accepted flags:
// - syscall.O_CREAT - create new fifo if one doesn't exist
// - syscall.O_RDONLY - open fifo only from reader side
// - syscall.O_WRONLY - open fifo only from writer side
// - syscall.O_RDWR - open fifo from both sides, never block on syscall level
// - syscall.O_NONBLOCK - return io.ReadWriteCloser even if other side of the
//     fifo isn't open. read/write will be connected after the actual fifo is
//     open or after fifo is closed.
func openFifo(ctx context.Context, fn string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if _, err := os.Stat(fn); err != nil {
		if os.IsNotExist(err) && flag&syscall.O_CREAT != 0 {
			if err := mkfifo(fn, uint32(perm&os.ModePerm)); err != nil && !os.IsExist(err) {
				return nil, errors.Wrapf(err, "error creating fifo %v", fn)
			}
		} else {
			return nil, err
		}
	}

	flag &= ^syscall.O_CREAT

	return os.OpenFile(fn, flag, 0)
}

func mkfifo(path string, mode uint32) (err error) {
	return unix.Mkfifo(path, mode)
}
