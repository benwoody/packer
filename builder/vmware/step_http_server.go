package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
	"net/http"
)

// This step creates and runs the HTTP server that is serving the files
// specified by the 'http_files` configuration parameter in the template.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type stepHTTPServer struct {
	l net.Listener
}

func (s *stepHTTPServer) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	var httpPort uint = 0
	if config.HTTPDir == "" {
		state["http_port"] = httpPort
		return multistep.ActionContinue
	}

	// Find an available TCP port for our HTTP server
	var httpAddr string
	portRange := int(config.HTTPPortMax - config.HTTPPortMin)
	for {
		var err error
		httpPort = uint(rand.Intn(portRange)) + config.HTTPPortMin
		httpAddr = fmt.Sprintf(":%d", httpPort)
		log.Printf("Trying port: %d", httpPort)
		s.l, err = net.Listen("tcp", httpAddr)
		if err == nil {
			break
		}
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on port %d", httpPort))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(config.HTTPDir))
	server := &http.Server{Addr: httpAddr, Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state["http_port"] = httpPort

	return multistep.ActionContinue
}

func (s *stepHTTPServer) Cleanup(map[string]interface{}) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
