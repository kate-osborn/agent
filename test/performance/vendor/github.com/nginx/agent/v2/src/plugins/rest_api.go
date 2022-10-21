package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/nginx/agent/sdk/v2"
	"github.com/nginx/agent/sdk/v2/proto"
	"github.com/nginx/agent/v2/src/core"
	"github.com/nginx/agent/v2/src/core/config"
	log "github.com/sirupsen/logrus"
)
var (
    instancesRegex = regexp.MustCompile(`^\/nginx[\/]*$`)
    configRegex = regexp.MustCompile(`^\/nginx/config[\/]*$`)
)

type NginxHandler struct {
	config          *config.Config
	env             core.Environment
	pipeline        core.MessagePipeInterface
	nginxBinary     core.NginxBinary
	responseChannel chan *proto.Command_NginxConfigResponse
}

type SingleConfig struct {
	messageId string
	nginxId   string
	file      *proto.File
}

type RestApi struct {
	config      *config.Config
	env         core.Environment
	pipeline    core.MessagePipeInterface
	nginxBinary core.NginxBinary
	handler     *NginxHandler
}

func NewRestApi(config *config.Config, env core.Environment, nginxBinary core.NginxBinary) *RestApi {
	return &RestApi{config: config, env: env, nginxBinary: nginxBinary}
}

func (r *RestApi) Init(pipeline core.MessagePipeInterface) {
	log.Info("REST API initializing")
	r.pipeline = pipeline
	go r.createHttpServer()
}

func (r *RestApi) Close() {
	log.Info("REST API is wrapping up")
}

func (r *RestApi) Process(message *core.Message) {
	log.Errorf("Process function in the rest_api.go, %s %v", message.Topic(), message.Data())
	switch message.Topic() {
	case core.RestAPIConfigApplyResponse:
		switch response := message.Data().(type) {
		case *proto.Command_NginxConfigResponse:
			log.Error("Command_NginxConfigResponse!!!!!!!")
			r.handler.responseChannel <- response
		}
	}
}

func (r *RestApi) Info() *core.Info {
	return core.NewInfo("REST API Plugin", "v0.0.1")
}

func (r *RestApi) Subscriptions() []string {
	return []string{
		core.RestAPIConfigApplyResponse,
	}
}

func (r *RestApi) createHttpServer()  {
	mux := http.NewServeMux()
    r.handler = &NginxHandler{r.config, r.env, r.pipeline, r.nginxBinary, make(chan *proto.Command_NginxConfigResponse, 1)}
    mux.Handle("/nginx/", r.handler)
    mux.Handle("/nginx/", r.handler)

	log.Info("Starting REST API HTTP server")

	server := http.Server{
        Addr:    ":9090",
        Handler:  mux,
    }

	if err := server.ListenAndServeTLS("/home/ubuntu/server.crt", "/home/ubuntu/server.key"); err != nil {
        log.Fatalf("error listening to port: %v", err)
    }
}

func (h *NginxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/json")
    switch {
    case r.Method == http.MethodGet && instancesRegex.MatchString(r.URL.Path):
        h.GetInstances(w, r)
        return
    case r.Method == http.MethodPut && configRegex.MatchString(r.URL.Path):
        h.Update(w, r)
        return
    default:
        notFound(w, r)
        return
    }
}

func (h *NginxHandler) GetInstances(w http.ResponseWriter, r *http.Request) {
	var nginxDetails []*proto.NginxDetails

	for _, proc := range h.env.Processes() {
		// only need master process for registration
		if proc.IsMaster {
			nginxDetails = append(nginxDetails, h.nginxBinary.GetNginxDetailsFromProcess(proc))
		} else {
			log.Tracef("NGINX non-master process: %d", proc.Pid)
		}
	}

	if len(nginxDetails) == 0 {
		log.Info("No master process found")
	}

	responseBodyBytes := new(bytes.Buffer)
	json.NewEncoder(responseBodyBytes).Encode(nginxDetails)

	w.WriteHeader(http.StatusOK)
    w.Write(responseBodyBytes.Bytes())
}

func (h *NginxHandler) Update(w http.ResponseWriter, r *http.Request) {
	log.Error("Update!!!!!!")
    r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Errorf("Can't read form file, %v", err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)


	var nginxDetails []*proto.NginxDetails

	for _, proc := range h.env.Processes() {
		// only need master process for registration
		if proc.IsMaster {
			nginxDetails = append(nginxDetails, h.nginxBinary.GetNginxDetailsFromProcess(proc))
		} else {
			log.Tracef("NGINX non-master process: %d", proc.Pid)
		}
	}

	if len(nginxDetails) == 0 {
		log.Info("No master process found")
	}


	log.Errorf("nginxDetails: %v", nginxDetails)

	for _, nginxDetail := range nginxDetails {
		fullFilePath := nginxDetail.ConfPath
		nginxId := nginxDetail.NginxId

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			log.Errorf("Can't read file, %v", err)
			return
		}

		protoFile := &proto.File{
			Name: fullFilePath,
			Permissions: "0755", // What permissions will we use?
			Contents: buf.Bytes(),
		}

		log.Errorf("protoFile: %v", protoFile)

		var configApply *sdk.ConfigApply
		configApply, err := sdk.NewConfigApply(protoFile.GetName(), h.config.AllowedDirectoriesMap)
	
		err = h.env.WriteFiles(configApply, []*proto.File{protoFile}, "", h.config.AllowedDirectoriesMap)
		if err != nil {
			rollbackErr := configApply.Rollback(err)
			log.Errorf("Config rollback failed: %v", rollbackErr)
			return
		}
		err = configApply.Complete()
		if err != nil {
			log.Errorf("unable to write config, %v", err)
			return
		}


		log.Error("File written!!!!!!!")

		conf, err := h.nginxBinary.ReadConfig(fullFilePath, nginxId, h.env.GetSystemUUID())
		if err != nil {
			log.Errorf("unable to read config , %v", err)
			return
		}

		log.Error("ReadConfig!!!!!!!")

		h.pipeline.Process(core.NewMessage(core.CommNginxConfig, conf))

		select {
		case response := <-h.responseChannel:
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(response)
			w.Write(reqBodyBytes.Bytes())
		case <-time.After(30*time.Second):
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("not found"))
}
