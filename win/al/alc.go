// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package al

/*
#cgo windows  CFLAGS:  -DGOOS_windows
#cgo windows  LDFLAGS: -lopenal

#ifdef GOOS_windows
#include <stdlib.h>
#include <AL/alc.h>
#endif
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

var (
	mu      sync.Mutex
	device  unsafe.Pointer
	context unsafe.Pointer
)

// DeviceError returns the last known error from the current device.
func DeviceError() int32 {
	return alcGetError(device)
}

// TODO(jbd): Investigate the cases where multiple audio output
// devices might be needed.

// OpenDevice opens the default audio device.
// Calls to OpenDevice are safe for concurrent use.
func OpenDevice() error {
	mu.Lock()
	defer mu.Unlock()

	// already opened
	if device != nil {
		return nil
	}

	dev := alcOpenDevice("")
	if dev == nil {
		return errors.New("al: cannot open the default audio device")
	}
	ctx := alcCreateContext(dev, nil)
	if ctx == nil {
		alcCloseDevice(dev)
		return errors.New("al: cannot create a new context")
	}
	if !alcMakeContextCurrent(ctx) {
		alcCloseDevice(dev)
		return errors.New("al: cannot make context current")
	}
	device = dev
	context = ctx
	return nil
}

// CloseDevice closes the device and frees related resources.
// Calls to CloseDevice are safe for concurrent use.
func CloseDevice() {
	mu.Lock()
	defer mu.Unlock()

	if device == nil {
		return
	}

	alcCloseDevice(device)
	if context != nil {
		alcDestroyContext(context)
	}
	device = nil
	context = nil
}

func alcGetError(d unsafe.Pointer) int32 {
	dev := (*C.ALCdevice)(d)
	return int32(C.alcGetError(dev))
}

func alcOpenDevice(name string) unsafe.Pointer {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	return (unsafe.Pointer)(C.alcOpenDevice((*C.ALCchar)(unsafe.Pointer(n))))
}

func alcCloseDevice(d unsafe.Pointer) bool {
	dev := (*C.ALCdevice)(d)
	return C.alcCloseDevice(dev) == C.ALC_TRUE
}

func alcCreateContext(d unsafe.Pointer, attrs []int32) unsafe.Pointer {
	dev := (*C.ALCdevice)(d)
	return (unsafe.Pointer)(C.alcCreateContext(dev, nil))
}

func alcMakeContextCurrent(c unsafe.Pointer) bool {
	ctx := (*C.ALCcontext)(c)
	return C.alcMakeContextCurrent(ctx) == C.ALC_TRUE
}

func alcDestroyContext(c unsafe.Pointer) {
	C.alcDestroyContext((*C.ALCcontext)(c))
}
