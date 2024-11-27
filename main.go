package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/terminalstatic/go-xsd-validate"
)

const (
	keyServerAddr = "serverAddr"
	ansSchemaRoot = "http://www.ans.gov.br/padroes/tiss/schemas"
)

var xsdHandler *xsdvalidate.XsdHandler

type Response struct {
	Valid bool  `json:"valid"`
	Error error `json:"error"`
}

func postValidateXML(w http.ResponseWriter, r *http.Request) {
	var versionURL string

	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	err := r.ParseMultipartForm(5 << 20)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		response := Response{
			Valid: false,
			Error: err,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := r.Context()

	xsdvalidate.Init()
	defer xsdvalidate.Cleanup()

	fmt.Printf("%s: Got XML Validation Request\n", ctx.Value(keyServerAddr))

	file, handler, err := r.FormFile("file")
	if err != nil {
		response := Response{
			Valid: false,
			Error: nil,
		}
		return
	}

	version := r.PostFormValue("version")
	if version == "" {
		scanner := bufio.NewScanner(r.FormFile("file"))
		for scanner.Scan() {
			ansMensagemTISS := strings.Split(scanner.Text(), " ")
			versionURL = ansMensagemTISS[len(ansMensagemTISS)-1]
			versionURL = versionURL[0 : len(versionURL)-2]
			break
		}
	} else {
		version = strings.ReplaceAll(version, ".", "_")
		versionURL = fmt.Sprintf("%s/tissV%s.xsd", ansSchemaRoot, version)
	}

	fmt.Println(versionURL)

	xsdHandler, err := xsdvalidate.NewXsdHandlerUrl(versionURL, xsdvalidate.ParsErrVerbose)
	if err != nil {
		response := Response{
			Valid: false,
			Error: err,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer xsdHandler.Free()

	response := Response{
		Valid: true,
		Error: nil,
	}
	json.NewEncoder(w).Encode(response)
}

/*
Lists available xml versions using file system
*/
func getAvailableXMLVersions(w http.ResponseWriter, r *http.Request) {
	var versions []string

	files, _ := filepath.Glob("./resources/xsd/tissV*.xsd")

	for _, file := range files {
		file_array := strings.Split(file, "/")
		file = file_array[len(file_array)-1]

		if len(file) == 16 {
			file = file[5:12]
			file = strings.ReplaceAll(file, "_", ".")
			versions = append(versions, file)
		}
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(versions)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./web/")))
	mux.HandleFunc("/available-xml-versions", getAvailableXMLVersions)
	mux.HandleFunc("/validate-xml", postValidateXML)

	ctx, cancelCtx := context.WithCancel(context.Background())
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server is closed")
	} else if err != nil {
		fmt.Printf("Error listening for server: %s\n", err)
	}
	cancelCtx()
}
