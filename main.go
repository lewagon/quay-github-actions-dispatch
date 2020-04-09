package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/imroc/req"
)

// QuayPayload represents the incoming webhook payload from Quay:
// https://docs.quay.io/guides/notifications.html
type QuayPayload struct {
	Repository      string   `json:"repository"`
	Namespace       string   `json:"namespace"`
	Name            string   `json:"name"`
	DockerURL       string   `json:"docker_url"`
	Homepage        string   `json:"homepage"`
	BuildID         string   `json:"build_id"`
	DockerTags      []string `json:"docker_tags"`
	TriggerKind     string   `json:"trigger_kind"`
	TriggerID       string   `json:"trigger_id"`
	TriggerMetadata struct {
		DefaultBranch string `json:"default_branch"`
		Ref           string `json:"ref"`
		Commit        string `json:"commit"`
		CommitInfo    struct {
			URL     string `json:"url"`
			Message string `json:"message"`
			Date    string `json:"date"`
			Author  struct {
				Username  string `json:"username"`
				URL       string `json:"url"`
				AvatarURL string `json:"avatar_url"`
			} `json:"author"`
			Committer struct {
				Username  string `json:"username"`
				URL       string `json:"url"`
				AvatarURL string `json:"avatar_url"`
			} `json:"committer"`
		} `json:"commit_info"`
	} `json:"trigger_metadata"`
}

// GithubPayload represents the outcoming payload go GithHub repository Dispatch webhook:
// https://developer.github.com/v3/repos/#create-a-repository-dispatch-event
type GithubPayload struct {
	EventType     string        `json:"event_type"`
	ClientPayload ClientPayload `json:"client_payload"`
}

// ClientPayload is part of GithubPayload that allows to set arbitrary text
// to be consumed inside the workflow triggered by repository_dispatch event
type ClientPayload struct {
	Text string `json:"text"`
}

// Used for debugging
func logHelper(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))
}

func handleIncoming(rw http.ResponseWriter, r *http.Request) {
	var in QuayPayload

	if os.Getenv("DEBUG") == "true" {
		req.Debug = true
		logHelper(r)
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&in)
	if err != nil {
		log.Println(err)
	}

	// Verify that the payload sender is indeed Quay
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cn := strings.ToLower(r.TLS.PeerCertificates[0].Subject.CommonName)
		if cn == "*.quay.io" {
			go postToActions(&in)
			rw.WriteHeader(http.StatusNoContent)
			return
		}
	}

	rw.WriteHeader(http.StatusForbidden)
}

func postToActions(in *QuayPayload) {
	sha7 := in.TriggerMetadata.Commit[0:7]
	out := &GithubPayload{
		EventType: "QUAY_BUILD_SUCCESS",
		ClientPayload: ClientPayload{
			Text: sha7,
		},
	}
	ghURL := fmt.Sprintf("https://api.github.com/repos/%s/dispatches", in.Repository)
	headers := req.Header{
		"Accept":        "application/vnd.github.everest-preview+json",
		"Authorization": fmt.Sprintf("token %s", os.Getenv("GH_TOKEN")),
	}
	req.Post(ghURL, headers, req.BodyJSON(&out))
}

func main() {
	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
		},
	}

	http.HandleFunc("/incoming", handleIncoming)

	log.Println(server.ListenAndServeTLS(os.Getenv("CRT"), os.Getenv("KEY")))
}
