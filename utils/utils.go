package utils

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgtype"

	"github.com/danthegoodman1/GoAPITemplate/gologger"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/segmentio/ksuid"
)

var logger = gologger.NewLogger()

func GetEnvOrDefault(env, defaultVal string) string {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		return e
	}
}

func GetEnvOrDefaultInt(env string, defaultVal int64) int64 {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		intVal, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			logger.Error().Msg(fmt.Sprintf("Failed to parse string to int '%s'", env))
			os.Exit(1)
		}

		return intVal
	}
}

func GenRandomID(prefix string) string {
	return prefix + gonanoid.MustGenerate("abcdefghijklmonpqrstuvwxyzABCDEFGHIJKLMONPQRSTUVWXYZ0123456789", 22)
}

func GenKSortedID(prefix string) string {
	return prefix + ksuid.New().String()
}

func GenRandomShortID() string {
	// reduced character set that's less probable to mis-type
	// change for conflicts is still only 1:128 trillion
	return gonanoid.MustGenerate("abcdefghikmonpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ0123456789", 8)
}

func DaysUntil(t time.Time, d time.Weekday) int {
	delta := d - t.Weekday()
	if delta < 0 {
		delta += 7
	}
	return int(delta)
}

func Ptr[T any](s T) *T {
	return &s
}

func Deref[T any](ref *T, fallback T) T {
	if ref == nil {
		return fallback
	}
	return *ref
}

func ArrayOrEmpty[T any](ref []T) []T {
	if ref == nil {
		return make([]T, 0)
	}
	return ref
}

var emptyJSON = pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}

func OrEmptyJSON(data pgtype.JSONB) pgtype.JSONB {
	if data.Status == pgtype.Null {
		data = emptyJSON
	}
	return data
}

func IfElse[T any](check bool, a T, b T) T {
	if check {
		return a
	}
	return b
}

func OrEmptyArray[T any](a []T) []T {
	if a == nil {
		return make([]T, 0)
	}
	return a
}

func FirstOr[T any](a []T, def T) T {
	if len(a) == 0 {
		return def
	}
	return a[0]
}

var ErrVersionBadFormat = PermError("bad version format")

// VersionToInt converts a simple semantic version string (e.e. 18.02.66)
func VersionToInt(v string) (int64, error) {
	sParts := strings.Split(v, ".")
	if len(sParts) > 3 {
		return -1, ErrVersionBadFormat
	}
	var iParts = make([]int64, 3)
	for i := range sParts {
		vp, err := strconv.ParseInt(sParts[i], 10, 64)
		if err != nil {
			return -1, fmt.Errorf("error in ParseInt: %s %w", err.Error(), ErrVersionBadFormat)
		}
		iParts[i] = vp
	}
	return iParts[0]*10_000*10_000 + iParts[1]*10_000 + iParts[2], nil
}

// FuncNameFQ returns the fully qualified name of the function.
func FuncNameFQ(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// FuncName returns the name of the function, without the package.
func FuncName(f any) string {
	fqName := FuncNameFQ(f)
	return fqName[strings.LastIndexByte(fqName, '.')+1:]
}

func AsErr[T error](err error) (te T, ok bool) {
	if err == nil {
		return te, false
	}
	return te, errors.As(err, &te)
}

// IsErr is useful for check for a class of errors (e.g. *serviceerror.WorkflowExecutionAlreadyStarted) instead of a specific error.
// E.g. Temporal doesn't even expose some errors, only their types
func IsErr[T error](err error) bool {
	_, ok := AsErr[T](err)
	return ok
}
