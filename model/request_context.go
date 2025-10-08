package model

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// verifyCodeStore holds verify codes keyed by per-goroutine UUID v4 for per-request propagation.
var verifyCodeStore sync.Map // map[string]string

// gidUUIDStore maps real goroutine ids to a stable UUID v4 for the lifetime of the goroutine.
var gidUUIDStore sync.Map // map[uint64]string

// getGoID returns the current goroutine's numeric id by parsing runtime.Stack output.
func getGoID() uint64 {
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

// getGID returns a stable UUID v4 associated with the current goroutine.
func getGID() string {
	gid := getGoID()
	if gid == 0 {
		return ""
	}
	if v, ok := gidUUIDStore.Load(gid); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s
		}
	}
	// Not found: generate and store a new UUID v4 for this goroutine id.
	u := uuid.NewV4().String()
	gidUUIDStore.Store(gid, u)
	return u
}

// SetCurrentVerifyCode stores the verify code for the current goroutine.
func SetCurrentVerifyCode(code string) {
	if strings.TrimSpace(code) == "" {
		return
	}
	gid := getGID()
	if gid == "" {
		return
	}
	verifyCodeStore.Store(gid, code)
}

// GetCurrentVerifyCode returns the verify code for the current goroutine, if any.
func GetCurrentVerifyCode() string {
	gid := getGID()
	if gid == "" {
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
	if gid == "" {
		return
	}
	verifyCodeStore.Delete(gid)
}
