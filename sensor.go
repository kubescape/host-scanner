package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/negroni"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type bodyLogWriter struct {
	negroni.ResponseWriter
	resBody *bytes.Buffer
}

// Common HTTP utils
func (blw bodyLogWriter) Write(b []byte) (int, error) {
	blw.resBody.Write(b)
	return blw.ResponseWriter.Write(b)
}

var (
	zapLogger *zap.Logger
)

func initLogger() *log.Logger {
	var err error
	// https://godoc.org/go.uber.org/zap#AtomicLevel.UnmarshalText
	lvl := zap.NewAtomicLevel()
	// if config.LogLevel == "" {
	// 	config.LogLevel = "warn"
	// }
	logLevel := "debug"
	if err := lvl.UnmarshalText([]byte(logLevel)); err != nil {
		panic(err)
	}
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	zapConf := zap.Config{DisableCaller: true, DisableStacktrace: true, Level: lvl,
		Encoding: "json", EncoderConfig: ec,
		OutputPaths: []string{"stdout"}, ErrorOutputPaths: []string{"stderr"}}
	// if config.LogFileName != "" { // empty string means the output is stdout
	// 	zapConf.OutputPaths = []string{config.LogFileName}
	// 	zapConf.ErrorOutputPaths = append(zapConf.ErrorOutputPaths, config.LogFileName)
	// }

	zapLogger, err = zapConf.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(zapLogger)
	zap.RedirectStdLog(zapLogger)
	return zap.NewStdLog(zapLogger)
}

func initHTTPRouter() http.Handler {
	negroniRouter := negroni.New()

	nLogger := negroni.NewLogger()
	nLogger.SetFormat("{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}}" + " {{.Request.RequestURI}}")
	negroniRouter.Use(negroni.NewRecovery())
	negroniRouter.Use(nLogger)
	negroniRouter.UseFunc(filterNLogHTTPErrors)
	negroniRouter.UseHandler(http.DefaultServeMux)
	return negroniRouter
}

// filterNLogHTTPErrors intercept every HTTP request and in case of failure it logs the request and the response
func filterNLogHTTPErrors(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	startTime := time.Now()
	zapArr := []zapcore.Field{zap.String("method", r.Method),
		zap.String("requestURI", r.RequestURI),
		zap.String("remoteAddr", r.RemoteAddr),
	}
	if !strings.HasPrefix(r.RequestURI, "/isAlive") {
		zap.L().Debug("In filterNLogHTTPErrors", zapArr...)
	}

	nrw, _ := rw.(negroni.ResponseWriter)

	var blw bodyLogWriter
	blw.ResponseWriter = nrw
	blw.resBody = bytes.NewBuffer([]byte(""))

	bodyBuffer, err := ioutil.ReadAll(r.Body)
	oldBody := r.Body
	defer oldBody.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(bodyBuffer))
	defer r.Body.Close()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "Cannot read body: %s", err.Error())
	} else {
		next(blw, r)
	}
	if blw.Status() < 200 || blw.Status() >= 300 || startTime.Before(time.Now().Add(time.Second*50*(-1))) {
		zapLogger.With(append(zapArr,
			zap.Timep("requestStartTime", &startTime),
			zap.String("Request body", string(bodyBuffer)),
			zap.Int("HTTP status", blw.Status()),
			zap.Int("Response size", blw.Size()),
			zap.String("Response body", blw.resBody.String()))...).Error("Request failed")
	}
	if r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
	}
}

// main
func main() {
	fmt.Println("Starting Kubescape cluster node host scanner service")
	if BuildVersion == "" {
		BuildVersion = "unknown"
	}
	fmt.Println("Build version: " + BuildVersion)
	baseLogger := initLogger()
	negroniRouter := initHTTPRouter()

	defer zapLogger.Sync()

	sensorManagerAddress := os.Getenv("ARMO_SENSORS_MANAGER")
	connectSensorsManagerWebSocket(sensorManagerAddress)
	initHTTPHandlers()
	listeningPort := 7888
	zapLogger.Info("Listening...", zap.Int("port", listeningPort))
	if strings.Contains(os.Getenv("CADB_DEBUG"), "pprof") {
		fmt.Println("Debug mode - pprof on")
		go func() {

			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
	listenAddress := fmt.Sprintf(":%d", listeningPort)
	server := http.Server{Addr: listenAddress, Handler: negroniRouter, ErrorLog: baseLogger, TLSConfig: &tls.Config{}}

	go func() {
		server.ListenAndServe()
	}()

	termChan := make(chan os.Signal, 1)
	//  os.Kill,syscall.SIGKILL, cannot be trapped
	signal.Notify(termChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-termChan // Blocks here until either SIGINT or SIGTERM is received.
	zap.L().Warn("signal received")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 61*time.Second)
	defer ctxCancel()
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("HTTP shutdown error", zap.Error(err))
	}

	zap.L().Warn("shutdown gracefully")

}

func connectSensorsManagerWebSocket(sensorManagerAddress string) {
	zap.L().Warn("Not implemented")
}
