package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/negroni"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/go-logger/zaplogger"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	zapLogger *otelzap.Logger
)

func initLogger() *log.Logger {
	// https://godoc.org/go.uber.org/zap#AtomicLevel.UnmarshalText
	lvl := zap.NewAtomicLevel()

	logLevel := logger.L().GetLevel()
	// TODO: change "warning" level to "warn" to match zap
	if err := lvl.UnmarshalText([]byte(logLevel)); err != nil {
		logger.L().Warning("failed to set zap logger level", helpers.Error(err))
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

	l, err := zapConf.Build()
	if err != nil {
		logger.L().Fatal(err.Error())
	}
	zapLogger = otelzap.New(l)

	otelzap.ReplaceGlobals(zapLogger)
	zap.RedirectStdLog(zapLogger.Logger)
	return zap.NewStdLog(zapLogger.Logger)
}

func CaselessMatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func initHTTPRouter() http.Handler {
	negroniRouter := negroni.New()

	negroniRouter.Use(negroni.NewRecovery())
	negroniRouter.UseFunc(filterNLogHTTPErrors)
	handler := http.Handler(http.DefaultServeMux)
	filteredEndPoints := []string{healthzEP, readyzEP}
	handler = otelhttp.NewHandler(
		handler,
		"",
		otelhttp.WithSpanNameFormatter(spanName),
		otelhttp.WithFilter(otelhttp.Filter(
			// This function return false in case the req.URL.Path is equal to "/readyz" or "/healthz".
			// If we want to exclude others endpoint from telemetry,
			// just add them in `filteredEndPoints` variable.
			func(req *http.Request) bool {
				for _, f := range filteredEndPoints {
					if req.URL.Path == f {
						return false
					}
				}
				return true
			}),
		),
	)
	negroniRouter.UseHandler(handler)
	return CaselessMatcher(negroniRouter)
}

func spanName(_ string, req *http.Request) string {
	return req.RequestURI
}

// filterNLogHTTPErrors intercept every HTTP request and in case of failure it logs the request and the response
func filterNLogHTTPErrors(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	startTime := time.Now()
	zapArr := []zapcore.Field{zap.String("method", r.Method),
		zap.String("requestURI", r.RequestURI),
		zap.String("remoteAddr", r.RemoteAddr),
	}
	if !strings.HasPrefix(r.RequestURI, "/isAlive") {
		logger.L().Debug("In filterNLogHTTPErrors", helpers.String("requestURI", r.RequestURI), helpers.String("remoteAddr", r.RemoteAddr))
	}

	nrw, _ := rw.(negroni.ResponseWriter)

	var blw bodyLogWriter
	blw.ResponseWriter = nrw
	blw.resBody = bytes.NewBuffer([]byte(""))

	bodyBuffer, err := io.ReadAll(r.Body)
	oldBody := r.Body
	defer oldBody.Close()
	r.Body = io.NopCloser(bytes.NewReader(bodyBuffer))
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
		io.Copy(io.Discard, r.Body)
	}
}

// main
func main() {
	logger.InitLogger(zaplogger.LoggerName)

	ctx := context.Background()
	// to enable otel, set OTEL_COLLECTOR_SVC=otel-collector:4317
	if otelHost, present := os.LookupEnv("OTEL_COLLECTOR_SVC"); present {
		ctx = logger.InitOtel("host-scanner",
			os.Getenv(BuildVersion),
			os.Getenv("ACCOUNT_ID"),
			os.Getenv("CLUSTER_NAME"),
			url.URL{Host: otelHost})
		defer logger.ShutdownOtel(ctx)
	}

	logger.L().Info("Starting Kubescape cluster node host scanner service", helpers.String("buildVersion", BuildVersion))
	baseLogger := initLogger()
	negroniRouter := initHTTPRouter()

	defer zapLogger.Sync()

	initHTTPHandlers()
	listeningPort := 7888
	logger.L().Info("Listening...", helpers.Int("port", listeningPort))
	if strings.Contains(os.Getenv("CADB_DEBUG"), "pprof") {
		logger.L().Debug("Debug mode - pprof on")
		go func() {
			logger.L().Error(http.ListenAndServe(":6060", nil).Error())
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
	logger.L().Ctx(ctx).Info("shutdown signal received")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 61*time.Second)
	defer ctxCancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.L().Ctx(ctx).Warning("HTTP shutdown error", helpers.Error(err))
	}

	logger.L().Ctx(ctx).Info("shutdown gracefully")

}
