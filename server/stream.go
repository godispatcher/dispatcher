package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/model"
)

// ServStreamApi starts a lightweight persistent TCP server (NDJSON) as an alternative transport.
// Each request and response is a single line JSON (newline-delimited JSON).
// The request JSON must conform to model.Document, at minimum including department, transaction, and form.
// Responses mirror HTTP behavior and contain a model.Document with either output or error.
func ServStreamApi(register *department.RegisterDispatcher) {
	// Derive stream port by incrementing HTTP port by 1 (e.g., 9000 -> 9001)
	port := deriveStreamPort(register.Port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("stream api listen error on port %s: %v", port, err)
		return
	}
	log.Printf("stream api listening on :%s (NDJSON)\n", port)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("stream api accept error: %v", err)
				continue
			}
			go handleStreamConn(conn)
		}
	}()
}

func deriveStreamPort(httpPort string) string {
	if httpPort == "" {
		return "9001"
	}
	if n, err := strconv.Atoi(httpPort); err == nil {
		return strconv.Itoa(n + 1)
	}
	// if not numeric, append suffix
	return strings.TrimSpace(httpPort) + "-stream"
}

func handleStreamConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewScanner(conn)
	// increase max token size to allow larger payloads (~10MB)
	buf := make([]byte, 0, 64*1024)
	reader.Buffer(buf, 10*1024*1024)

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}
		var document model.Document
		if err := json.Unmarshal([]byte(line), &document); err != nil {
			writeStreamError(conn, err)
			continue
		}

		responseDoc := executeDocument(document)
		b, err := json.Marshal(responseDoc)
		if err != nil {
			writeStreamError(conn, err)
			continue
		}
		// Send response followed by newline
		if _, err := fmt.Fprintln(conn, string(b)); err != nil {
			return
		}
	}
	// Optionally log scanner error
	if err := reader.Err(); err != nil {
		log.Printf("stream conn scanner error: %v", err)
	}
}

func writeStreamError(conn net.Conn, err error) {
	out := model.Document{Type: "Error", Error: err.Error()}
	b, _ := json.Marshal(out)
	_, _ = fmt.Fprintln(conn, string(b))
}

// executeDocument mirrors the core logic from HTTP RegisterMainFunc without HTTP specifics.
func executeDocument(document model.Document) model.Document {
	if (&document) == nil {
		return model.Document{Type: "Error", Error: errors.New("invalid document").Error()}
	}
	// Find transaction
	ta := department.DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if ta != nil {
		outputDoc := (*ta).GetTransaction().Init(document)

		// Chain dispatchings if provided
		if document.Dispatchings != nil {
			for _, v := range document.Dispatchings {
				cta := department.DispatcherHolder.GetTransaction(v.Department, v.Transaction)
				if cta != nil {
					dOutputDoc := (*cta).GetTransaction().Init(*v)
					outputDoc.Dispatchings = append(outputDoc.Dispatchings, &dOutputDoc)
					// if an error occurs in a chained dispatching, stop early
					if dOutputDoc.Error != nil {
						break
					}
				}
			}
		}
		return outputDoc
	}
	return model.Document{Department: document.Department, Transaction: document.Transaction, Error: errors.New("transaction not found").Error(), Type: "Error"}
}
