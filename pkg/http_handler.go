package pkg

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

type HttpHandler struct {
	server *Server
}

func (h *HttpHandler) whepHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request to %s, method = %s\n", req.URL, req.Method)

	res.Header().Add("Access-Control-Allow-Origin", "*")
	res.Header().Add("Access-Control-Allow-Methods", "POST")
	res.Header().Add("Access-Control-Allow-Headers", "*")
	res.Header().Add("Access-Control-Allow-Headers", "Authorization")

	if req.Method == http.MethodOptions {
		return
	}

	offer, err := io.ReadAll(req.Body)
	if err != nil {
		log.Panic().Err(err).Msg("failed to read request body")
	}

	answer, err := h.server.CreateClient(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer, SDP: string(offer),
	})
	if err != nil {
		log.Panic().Err(err).Msg("failed to create client")
	}

	res.Header().Add("Location", "/connect")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, answer.SDP)
}
