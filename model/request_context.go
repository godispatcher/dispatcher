package model

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// verifyCodeStore holds verify codes keyed by goroutine id for per-request propagation.
var verifyCodeStore sync.Map // map[uint64]string

// getGID returns the current goroutine id by parsing runtime.Stack output.
// Note: This relies on runtime internals; keep usage minimal and confined to per-request context.
func getGID() uint64 {
	// Allocate a small buffer; single-line header fits.
	var b [64]byte
	n := runtime.Stack(b[:], false)
	// Stack header: "goroutine 123 [running]:\n"
	fields := strings.Fields(strings.TrimPrefix(string(b[:n]), "goroutine "))
	if len(fields) > 0 {
		if id, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
			return id
		}
	}
	return 0
}

// SetCurrentVerifyCode stores the verify code for the current goroutine.
func SetCurrentVerifyCode(code string) {
	if strings.TrimSpace(code) == "" {
		return
	}
	gid := getGID()
	if gid == 0 {
		return
	}
	verifyCodeStore.Store(gid, code)
}

// GetCurrentVerifyCode returns the verify code for the current goroutine, if any.
func GetCurrentVerifyCode() string {
	gid := getGID()
	if gid == 0 {
		return ""
	}
	if v, ok := verifyCodeStore.Load(gid); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// ClearCurrentVerifyCode removes any stored verify code for the current goroutine.
func ClearCurrentVerifyCode() {
	gid := getGID()
	if gid == 0 {
		return
	}
	verifyCodeStore.Delete(gid)
}
