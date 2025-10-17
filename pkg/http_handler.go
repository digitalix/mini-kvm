package pkg

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

type HttpHandler struct {
	server *Server
}

func (h *HttpHandler) whepHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request to %s, method = %s\n", req.URL, req.Method)

	res.Header().Add("Access-Control-Allow-Origin", "*")
	res.Header().Add("Access-Control-Allow-Methods", "POST, PATCH, DELETE")
	res.Header().Add("Access-Control-Allow-Headers", "*")
	res.Header().Add("Access-Control-Expose-Headers", "Location")
	//res.Header().Add("Access-Control-Allow-Headers", "Authorization")

	if req.Method == http.MethodOptions {
		return
	}

	switch req.Method {
	case http.MethodPost:
		h.handleWhepPost(res, req)
	case http.MethodPatch:
		h.handleWhepPatch(res, req)
	}

}

func (h *HttpHandler) handleWhepPost(res http.ResponseWriter, req *http.Request) {
	offer, err := io.ReadAll(req.Body)
	if err != nil {
		log.Panic().Err(err).Msg("failed to read request body")
	}

	clientId, answer, err := h.server.CreateClient(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer, SDP: string(offer),
	})
	if err != nil {
		log.Panic().Err(err).Msg("failed to create client")
	}

	res.Header().Add("Location", fmt.Sprintf("/connect?id=%s", clientId))
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, answer.SDP)
}

func (h *HttpHandler) handleWhepPatch(res http.ResponseWriter, req *http.Request) {
	clientId := req.URL.Query().Get("id")
	if clientId == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	client, exists := h.server.clients.Load(clientId)
	if !exists {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Panic().Err(err).Msg("failed to read request body")
	}

	candidateLines := parseTrickleICE(string(body))
	for _, line := range candidateLines {
		if strings.HasPrefix(line, "a=candidate:") {
			candidate := webrtc.ICECandidateInit{
				Candidate: strings.TrimPrefix(line, "a="),
			}
			err := client.connection.AddICECandidate(candidate)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	res.WriteHeader(http.StatusOK)
}

func parseTrickleICE(sdpFrag string) []string {
	lines := strings.Split(sdpFrag, "\r\n")
	var candidates []string
	for _, line := range lines {
		if strings.HasPrefix(line, "a=candidate:") {
			candidates = append(candidates, line)
		}
	}
	return candidates
}
