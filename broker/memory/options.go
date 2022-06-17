package memory

// Worker is an option to override the default number of worker and buffer.
func Worker(worker, buffer int) Option {
	return func(b *Broker) {
		b.worker = worker
		b.buf = buffer
	}
}
