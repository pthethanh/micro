package timeutil

import (
	"fmt"
	"sync"
	"time"
)

type (
	// Recorder is a helper for recording execution time of a execution context.
	// An execution context can contains multiple spans/actions.
	Recorder struct {
		ctx   string
		s     time.Time
		t     time.Time
		m     map[string]time.Duration
		meta  map[string]any
		mutex *sync.Mutex
	}

	// Record hold execution time information of a recorder.
	Record struct {
		Context  string                   `json:"context"`
		Duration time.Duration            `json:"duration"`
		Spans    map[string]time.Duration `json:"spans"`
		Meta     map[string]any           `json:"meta"`
	}
)

// NewRecorder start and return new recorder of the given execution context.
// Additional metadata can be provided as key/value pairs.
func NewRecorder(context string, meta ...any) *Recorder {
	r := &Recorder{
		ctx:   context,
		m:     make(map[string]time.Duration),
		mutex: &sync.Mutex{},
		meta:  make(map[string]any),
	}
	for i := 0; i < len(meta)/2; i++ {
		r.meta[fmt.Sprintf("%v", meta[i])] = meta[i+1]
	}
	if len(meta)%2 != 0 {
		r.meta["msg"] = meta[len(meta)-1]
	}
	t := time.Now()
	r.s = t
	r.t = t
	return r
}

// Done records the execution time of the given span since the last time Done was called.
// Done updates the internal clock of the recorder.
func (r *Recorder) Done(span string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.m[span] = time.Since(r.t)
	r.t = time.Now()
}

// Last return the last recorded time of the internal clock of the recorder.
func (r *Recorder) Last() time.Time {
	return r.t
}

// DoneSince records the execution time of the given span since the given time.
// Note that DoneSince doesn't update the internal clock of the recorder.
// Normally it is used in case there is a span run concurrently with others.
func (r *Recorder) DoneSince(span string, t time.Time) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.m[span] = time.Since(t)
}

// Info return time execution info of the context and detail
// of the recorded spans at the moment.
func (r *Recorder) Info() Record {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return Record{
		Context:  r.ctx,
		Duration: time.Since(r.s),
		Spans:    r.m,
		Meta:     r.meta,
	}
}

// String return string value of the record.
func (r Record) String() string {
	return fmt.Sprintf("context:%s, duration:%v, spans:%v, meta:%v", r.Context, r.Duration, r.Spans, r.Meta)
}
