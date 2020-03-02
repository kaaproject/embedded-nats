// enats package provides embedded NATS server that starts in a goroutine on localhost.
// Useful for testing inter-service communication.
//
// Copyright 2020 KaaIoT Technologies, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package enats

import (
	"fmt"
	"log"
	"time"

	gnatsd "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats.go"
	"github.com/phayes/freeport"
)

// EmbeddedNATS combines an embedded NATS server and NATS client connected to it that can be used for testing
// inter-service communication.
type EmbeddedNATS struct {
	server *gnatsd.Server
	Port   int

	Conn *nats.Conn
}

// NewEmbeddedNATS creates a new embedded NATS server bound to a randomly chosen free localhost port.
// One of the return parameters is always nil.
func NewEmbeddedNATS() (*EmbeddedNATS, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	return &EmbeddedNATS{
		server: gnatsd.New(&gnatsd.Options{Host: "localhost", Port: port}),
		Port:   port,
	}, nil
}

// Start the embedded NATS server in a separate goroutine and connect the Conn to it.
func (n *EmbeddedNATS) Start() error {
	// Start NATS server
	go func() {
		if err := gnatsd.Run(n.server); err != nil {
			log.Printf("Error running embedded NATS server: %v", err)
		}
	}()

	// Wait until the server is ready to accept connections
	if !n.server.ReadyForConnections(time.Minute) {
		return fmt.Errorf("NATS server not ready")
	}

	// Start NATS connector
	conn, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", n.Port))
	if err != nil {
		return fmt.Errorf("error connecting to NATS server at localhost port %d: %v", n.Port, err)
	}

	n.Conn = conn

	return nil
}

// Stop disconnects the NATS client and shuts down the embedded NATS server.
func (n *EmbeddedNATS) Stop() {
	n.Conn.Close()
	n.server.Shutdown()
}
