package protobufext

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/MatusOllah/slogcolor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	internalLogger = slog.New(
		slogcolor.NewHandler(os.Stdout, slogcolor.DefaultOptions))
	DefaultDialOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithWriteBufferSize(10 * 1024 * 1024),
		grpc.WithReadBufferSize(10 * 1024 * 1024),
	}
)

type ConnectionPool struct {
	conns map[string]*grpc.ClientConn
	mtx   sync.RWMutex
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		conns: make(map[string]*grpc.ClientConn),
		mtx:   sync.RWMutex{},
	}
}

func (cp *ConnectionPool) GetConn(target string, opt ...grpc.DialOption) (*grpc.ClientConn, error) {
	cp.mtx.RLock()
	if c, ok := cp.conns[target]; !ok {
		cp.mtx.RUnlock()
	} else {
		// Check if the connection is still usable
		state := c.GetState()
		if state != connectivity.Shutdown {
			cp.mtx.RUnlock()
			return c, nil
		}

		// Connection is shutdown, need to recreate
		cp.mtx.RUnlock()

		// Remove the dead connection
		cp.mtx.Lock()
		delete(cp.conns, target)
		cp.mtx.Unlock()
	}

	// second check only if a new connection is added to the map while we were waiting for the first lock
	cp.mtx.Lock()
	defer cp.mtx.Unlock()

	// second check
	if c, ok := cp.conns[target]; ok {
		state := c.GetState()
		if state != connectivity.Shutdown {
			return c, nil
		}
		// Connection is shutdown, remove and recreate
		delete(cp.conns, target)
	}

	// default connection creation
	newConn, err := grpc.NewClient(target, opt...)
	if err != nil {
		internalLogger.Error("failed to create default connection", "error", err)
		return nil, err
	}

	cp.conns[target] = newConn
	return newConn, nil
}

// Close all connections with context timeout for graceful shutdown
func (cp *ConnectionPool) Close(ctx context.Context) error {
	// Copy connections to close without holding a lock
	cp.mtx.Lock()
	if len(cp.conns) == 0 {
		cp.mtx.Unlock()
		return nil
	}

	connsToClose := make(map[string]*grpc.ClientConn, len(cp.conns))
	for target, c := range cp.conns {
		connsToClose[target] = c
	}
	// Clear the map immediately to prevent new usage
	cp.conns = make(map[string]*grpc.ClientConn)
	cp.mtx.Unlock()

	// Close all connections sequentially
	var errs []error
	for target, conn := range connsToClose {
		// Check context before each close
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("context cancelled before closing %s: %w", target, err))
			break
		}

		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection for target %s: %w", target, err))
		}
	}

	if len(errs) > 0 {
		internalLogger.Error("failed to close all connections", "errors", errs)
		return fmt.Errorf("failed to close all connections: %v", errs)
	}
	return nil
}
