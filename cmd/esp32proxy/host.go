package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/meyskens/esp32proxy/pkg/endpoints"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewHostCmd())
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type config struct {
	Endpoints map[string]string `json:"endpoints"`
}

type hostCmdOptions struct {
	BindAddr   string
	Port       int
	ConfigFile string

	db     *endpoints.EndpointDB
	config *config
}

// NewHostCmd generates the `host` command
func NewHostCmd() *cobra.Command {
	s := hostCmdOptions{}
	c := &cobra.Command{
		Use:     "host",
		Short:   "hosts the public endpoint",
		Long:    `hosts the public endpoint on the given bind address`,
		PreRunE: s.Validate,
		RunE:    s.RunE,
	}
	c.Flags().StringVarP(&s.BindAddr, "bind-address", "b", "0.0.0.0", "address to bind port to")
	c.Flags().IntVarP(&s.Port, "port", "p", 80, "port to bind port to")
	c.Flags().StringVarP(&s.ConfigFile, "config", "c", "config.json", "config file to use")

	return c
}

func (h *hostCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (h *hostCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	// load config
	h.config = &config{}
	f, err := os.Open(h.ConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(h.config); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h.db = endpoints.NewEndpointsDB()

	go h.serveHTTP()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *hostCmdOptions) serveHTTP() {
	log.Print("Started HTTP proxy on port ", h.Port)

	// start server
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		ep, err := h.db.Get(strings.Split(req.Host, ":")[0]) // remove port
		if err != nil || ep == nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("no endpoint found"))
			return
		}

		resp := ep.Request(req)

		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(resp.StatusCode)

		// copy body to response
		if resp.Body != nil {
			defer resp.Body.Close()
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				log.Print("error copying body: ", err)
			}
		}
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, req *http.Request) {
		// check token in header
		token := req.Header.Get("Token")
		endpoint := ""
		for ep, t := range h.config.Endpoints {
			if t == token {
				endpoint = ep
				break
			}
		}

		if endpoint == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
			return
		}

		log.Println("Got a connetion from an ESP, welcome!")

		// set up web socket
		ws, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		h.db.Add(endpoint, endpoints.NewEndpointDialer(ws))

		// remove on close
		ws.SetCloseHandler(func(code int, text string) error {
			h.db.Remove(endpoint)
			return nil
		})
	})

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", h.BindAddr, h.Port), nil); err != nil {
		log.Fatal(err)
	}
}
