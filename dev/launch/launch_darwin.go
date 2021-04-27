// +build arm,darwin amd,darwin

package launch

/*
#include <errno.h>
#include <launch.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"net"
	"os"
	"unsafe"
)

var (
	ErrNotExist         = errors.New("The socket name specified does not exist in the caller's launchd.plist")
	ErrNotManaged       = errors.New("The calling process is not managed by launchd")
	ErrAlreadyActivated = errors.New("The specified socket has already been activated")
)

func activateSocket(name string) ([]int, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var cFds *C.int
	cCnt := C.size_t(0)
	if ret := C.launch_activate_socket(cName, &cFds, &cCnt); ret != 0 {
		switch ret {
		case C.ENOENT:
			return nil, ErrNotExist
		case C.ESRCH:
			return nil, ErrNotManaged
		case C.EALREADY:
			return nil, ErrAlreadyActivated
		default:
			return nil, errors.New(C.GoString(C.strerror(ret)))
		}
	}
	ptr := unsafe.Pointer(cFds)
	defer C.free(ptr)
	cnt := int(cCnt)
	fds := (*[1 << 30]C.int)(ptr)[:cnt:cnt]
	res := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		res[i] = int(fds[i])
	}
	return res, nil
}

func SocketFiles(name string) ([]*os.File, error) {
	fds, err := activateSocket(name)
	if err != nil {
		return nil, err
	}

	files := make([]*os.File, 0)
	for _, fd := range fds {
		file := os.NewFile(uintptr(fd), "")
		files = append(files, file)
	}

	return files, nil
}

func SocketListeners(name string) ([]net.Listener, error) {
	files, err := SocketFiles(name)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0)
	for _, file := range files {
		listener, err := net.FileListener(file)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}
