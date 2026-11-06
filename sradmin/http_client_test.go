package sradmin

import (
	"fmt"
	"ktea/config"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListSubjectsCapsConnections(t *testing.T) {
	const subjectCount = 50
	var names []string
	for i := range subjectCount {
		names = append(names, fmt.Sprintf(`"s%d"`, i))
	}
	subjects := "[" + strings.Join(names, ",") + "]"

	var inFlight, maxInFlight atomic.Int32
	var conns atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			m := maxInFlight.Load()
			if n <= m || maxInFlight.CompareAndSwap(m, n) {
				break
			}
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/subjects":
			fmt.Fprint(w, subjects)
		case strings.HasSuffix(r.URL.Path, "/versions"):
			time.Sleep(10 * time.Millisecond)
			fmt.Fprint(w, "[1]")
		case strings.HasPrefix(r.URL.Path, "/config/"):
			time.Sleep(10 * time.Millisecond)
			fmt.Fprint(w, `{"compatibilityLevel":"BACKWARD"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	server.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			conns.Add(1)
		}
	}
	server.Start()
	defer server.Close()

	sra := New(&config.SchemaRegistryConfig{Url: server.URL})
	msg := sra.ListSubjects().(SubjectListingStartedMsg)
	listed, ok := msg.AwaitCompletion().(SubjectsListedMsg)

	assert.True(t, ok)
	assert.Len(t, listed.Subjects, subjectCount)
	assert.LessOrEqual(t, maxInFlight.Load(), int32(maxConnsPerRegistry))
	assert.LessOrEqual(t, conns.Load(), int32(maxConnsPerRegistry))
}
