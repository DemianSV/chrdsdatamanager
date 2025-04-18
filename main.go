package main

import (
	"bytes"
	"compress/zlib"
	"context"

	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/gocql/gocql"

	"chrdsdatamanager/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const (
	// Version
	Version string = "1.0.4"
	// Version string = "1.0.5"
	// LogFile General log file
	LogFile string = "chrdsdatamanager.log"
	// ConfigFile General configuration file
	ConfigFile string = "chrdsdatamanager.json"
)

// FlagLog Type of log FILE - STDOUT
var FlagLog string

var Session *gocql.Session

// Zabbix Agent
const ZBXHeaderSize = 4 + 1 + 4 + 4
const ZBXTCPProtocol = byte(0x01)
const ZBXZLibCompress = byte(0x03)
const ZBXCompress = true

var ZBXAgentMap map[string]string

var ConsistencyRead gocql.Consistency

type RequestAgentTypeT struct {
	Request string `json:"request"`
}

type RequestAgentT struct {
	Request string `json:"request"`
	Host    string `json:"host"`
	Version string `json:"version"`
	Session string `json:"session"`
}

type ResponseServerDataT struct {
	Itemid      int64  `json:"itemid"`
	Key         string `json:"key"`
	KeyOrig     string `json:"key_orig"`
	Delay       string `json:"delay"`
	Lastlogsize int    `json:"lastlogsize"`
	Mtime       int    `json:"mtime"`
}

type ResponseServerT struct {
	Response string                `json:"response"`
	Data     []ResponseServerDataT `json:"data"`
}

type RequestAgentMetricsDataT struct {
	ID     int32  `json:"id"`
	ItemID int64  `json:"itemid"`
	Value  string `json:"value"`
	Clock  int32  `json:"clock"`
	NS     int32  `json:"ns"`
}

var AgentStart map[string]bool

// Zabbix Agent END

type JobT struct {
	key   string
	value string
}

type RequestAgentMetricsT struct {
	Request string                     `json:"request"`
	Host    string                     `json:"host"`
	Data    []RequestAgentMetricsDataT `json:"data"`
	Session string                     `json:"session"`
}

type ResponsServerMetricsT struct {
	Response string `json:"response"`
	Info     string `json:"info"` // "info": "processed: 2; failed: 0; total: 2; seconds spent: 0.003534"
}

type ResponseTaskGetT struct {
	Object   string `json:"object"`
	Metric   string `json:"metric"`
	SpaceID  string `json:"spaceid"`
	Status   int    `json:"status"`
	Critical string `json:"critical"`
	Warning  string `json:"warning"`
	Interval int64  `json:"interval"`
}

type ResponseTaskGetArrayT struct {
	Tasks []ResponseTaskGetT `json:"data"`
}

type ResponseVersionT struct {
	Version    string `json:"version"`
	APIVersion string `json:"apiversion"`
}

type RequestRawDataSaveT struct {
	SpaceID   string  `json:"spaceid"`
	Metric    string  `json:"metric"`
	Value     float32 `json:"value"`
	EventTime int64   `json:"eventtime"`
	Object    string  `json:"object"`
}

type RequestRawDataSaveArrayT struct {
	Data []RequestRawDataSaveT `json:"data"`
}

type RequestRawTextSaveT struct {
	SpaceID   string `json:"spaceid"`
	Metric    string `json:"metric"`
	Value     string `json:"value"`
	EventTime int64  `json:"eventtime"`
	Object    string `json:"object"`
}

type RequestRawTextSaveArrayT struct {
	Data []RequestRawTextSaveT `json:"data"`
}

type ResponseRawDataSaveT struct {
	SpaceID string `json:"spaceid"`
	Metric  string `json:"metric"`
	Object  string `json:"object"`
	Status  int    `json:"status"`
}

type ResponseRawDataSaveArrayT struct {
	Data []ResponseRawDataSaveT `json:"data"`
}

type RequestRawDataGetT struct {
	SpaceID string `json:"spaceid"`
	Metric  string `json:"metric"`
}

type ResponseRawDataGetT struct {
	Metric     string  `json:"metric"`
	Value      float32 `json:"value"`
	CreateTime int64   `json:"createtime"`
	EventTime  int64   `json:"eventtime"`
	Status     int     `json:"status"`
	SpaceID    string  `json:"spaceid"`
	Object     string  `json:"object"`
}

type ResponseRawDataGetArrayT struct {
	Data []ResponseRawDataGetT `json:"data"`
}

type ResponseRawTextGetT struct {
	Metric     string `json:"metric"`
	Value      string `json:"value"`
	CreateTime int64  `json:"createtime"`
	EventTime  int64  `json:"eventtime"`
	Status     int    `json:"status"`
	SpaceID    string `json:"spaceid"`
	Object     string `json:"object"`
}

type ResponseRawTextGetArrayT struct {
	Data []ResponseRawTextGetT `json:"data"`
}

func init() {
	flag.StringVar(&FlagLog, "log", "stdout", "[-log file, stdout (default)]")
	flag.Parse()
	if FlagLog == "file" {
		fileLog, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Print("An error in opening a log file!")
			os.Exit(-1)
		}
		log.SetOutput(fileLog)
	}

	err := AppConfig.LoadConfig()
	if err != nil {
		log.Fatal("Configuration loading error (", err.Error(), ")!")
		return
	}
}

func done() {
}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api_key
func main() {
	defer done()
	defer log.Print("The service is stopped!")

	// Connect to DB cluster
	cluster := gocql.NewCluster(AppConfig.DB.HOSTS...)
	cluster.Timeout = time.Duration(AppConfig.DB.TIMEOUT) * time.Millisecond
	cluster.Keyspace = AppConfig.DB.KEYSPACE

	AppConfig.DB.CONSISTENCY = strings.ToUpper(AppConfig.DB.CONSISTENCY)
	switch AppConfig.DB.CONSISTENCY {
	case "QUORUM":
		cluster.Consistency = gocql.Quorum
	case "ANY":
		cluster.Consistency = gocql.Any
	case "ONE":
		cluster.Consistency = gocql.One
	case "TWO":
		cluster.Consistency = gocql.Two
	case "THREE":
		cluster.Consistency = gocql.Three
	case "ALL":
		cluster.Consistency = gocql.All
	case "LOCALQUORUM":
		cluster.Consistency = gocql.LocalQuorum
	case "EACHQUORUM":
		cluster.Consistency = gocql.EachQuorum
	case "LOCALONE":
		cluster.Consistency = gocql.LocalOne
	default:
		cluster.Consistency = gocql.Quorum
	}
	AppConfig.DB.CONSISTENCYREAD = strings.ToUpper(AppConfig.DB.CONSISTENCYREAD)
	switch AppConfig.DB.CONSISTENCYREAD {
	case "QUORUM":
		ConsistencyRead = gocql.Quorum
	case "ANY":
		ConsistencyRead = gocql.Any
	case "ONE":
		ConsistencyRead = gocql.One
	case "TWO":
		ConsistencyRead = gocql.Two
	case "THREE":
		ConsistencyRead = gocql.Three
	case "ALL":
		ConsistencyRead = gocql.All
	case "LOCALQUORUM":
		ConsistencyRead = gocql.LocalQuorum
	case "EACHQUORUM":
		ConsistencyRead = gocql.EachQuorum
	case "LOCALONE":
		ConsistencyRead = gocql.LocalOne
	default:
		ConsistencyRead = gocql.Quorum
	}

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: AppConfig.DB.USERNAME,
		Password: AppConfig.DB.PASSWORD,
	}
	if AppConfig.DB.TLS {
		cluster.SslOpts = &gocql.SslOptions{
			EnableHostVerification: AppConfig.DB.HOSTVERIFICATION,
			CertPath:               AppConfig.DB.CERTPATH,
			KeyPath:                AppConfig.DB.KEYPATH,
			CaPath:                 AppConfig.DB.CAPATH,
		}
	}

	var err error
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("It was not possible to establish a connection with the database (", err, ")!")
	} else {
		log.Print("The connection to the database is established... OK")
	}
	defer Session.Close()

	// ZBX Server start
	go zbxServer()

	// WorkerPool start
	go workersRun()

	// HTTP Server
	docs.SwaggerInfo.Title = AppConfig.SWG.TITLE
	docs.SwaggerInfo.Description = AppConfig.SWG.DESCRIPTION
	docs.SwaggerInfo.Version = AppConfig.SWG.VERSION
	docs.SwaggerInfo.Host = AppConfig.SWG.HOST
	docs.SwaggerInfo.BasePath = AppConfig.SWG.BASEPATH
	docs.SwaggerInfo.Schemes = []string{AppConfig.SWG.SCHEME}

	r := chi.NewRouter() // Initialization GoChi Router

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger) // Expanded logging
	r.Use(middleware.URLFormat)

	r.Use(httprate.Limit(
		AppConfig.HTTP.RL, // Requests
		1*time.Second,     // Per duration
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return r.Header.Get("api_key"), nil
		}),
	))

	// Requests to the API
	r.Group(func(r chi.Router) {
		r.Route("/api/v{version}/version", func(r chi.Router) {
			r.Get("/", getVersion)
		})
		r.Route("/api/v{version}/task", func(r chi.Router) {
			r.Get("/get", getTask)
		})
		r.Route("/api/v{version}/data", func(r chi.Router) {
			r.Post("/save", postRawDataSave)
			r.Post("/get", postRawDataGet)
		})
		r.Route("/api/v{version}/text", func(r chi.Router) {
			r.Post("/save", postRawTextSave)
			r.Post("/get", postRawTextGet)
		})
	})

	r.Mount("/swagger", httpSwagger.WrapHandler)

	if AppConfig.HTTP.TLS { // HTTPS connection
		crt, err := tls.LoadX509KeyPair(AppConfig.HTTP.CERTPATH, AppConfig.HTTP.KEYPATH)
		if err != nil {
			log.Fatal("Failed to download the certificate (", err, ")!")
			return
		}
		var crtPool *x509.CertPool
		if AppConfig.HTTP.CAPATH == "" {
			crtPool = nil
		} else {
			crtPool = x509.NewCertPool()
			if crtCA, err := os.ReadFile(AppConfig.HTTP.CAPATH); err != nil {
				log.Fatal("Failed to download CA certificate (", err, ")!")
			} else if ok := crtPool.AppendCertsFromPEM(crtCA); !ok {
				log.Fatal("It was not possible to apply CA certificate!")
			}
		}
		tlsConfig := &tls.Config{
			RootCAs:      crtPool,
			Certificates: []tls.Certificate{crt},
		}

		httpServer := &http.Server{Addr: AppConfig.HTTP.HOST + ":" + AppConfig.HTTP.PORT,
			Handler:      r,
			ReadTimeout:  time.Duration(AppConfig.HTTP.READTIMEOUT) * time.Second,
			WriteTimeout: time.Duration(AppConfig.HTTP.WRITETIMEOUT) * time.Second,
			TLSConfig:    tlsConfig}
		log.Print("Datamanager service is launched at the port ", AppConfig.HTTP.PORT, "(HTTPS)... ОК")
		httpServer.ListenAndServeTLS("", "")
	} else { // HTTP соединение
		httpServer := &http.Server{Addr: AppConfig.HTTP.HOST + ":" + AppConfig.HTTP.PORT,
			Handler:      r,
			ReadTimeout:  time.Duration(AppConfig.HTTP.READTIMEOUT) * time.Second,
			WriteTimeout: time.Duration(AppConfig.HTTP.WRITETIMEOUT) * time.Second}
		log.Print("Datamanager service is launched at the port ", AppConfig.HTTP.PORT, "(HTTP)... ОК")
		httpServer.ListenAndServe()
	}
}

// @Summary		Get a list of all your task to fulfill
// @Tags		Task
// @Accept		json
// @Produce		json
// @Success		200 {object} ResponseTaskGetArrayT
// @Success		204 "NoContent"
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/task/get [get]
// @Security	ApiKeyAuth
func getTask(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	ctx := context.Background()
	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			var responseTasksArray ResponseTaskGetArrayT

			scanner := Session.Query(`SELECT object, metric, status, space_id, critical, warning, interval FROM task WHERE module_id = ?`, r.Header.Get("api_key")).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			for scanner.Next() {
				var object string
				var metric string
				var status int
				var spaceid string
				var critical string
				var warning string
				var interval int64
				err := scanner.Scan(&object, &metric, &status, &spaceid, &critical, &warning, &interval)
				if err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					if status == 1 {
						responseTask := &ResponseTaskGetT{
							Object:   object,
							Metric:   metric,
							Status:   status,
							SpaceID:  spaceid,
							Critical: critical,
							Warning:  warning,
							Interval: interval,
						}
						responseTasksArray.Tasks = append(responseTasksArray.Tasks, *responseTask)
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(responseTasksArray.Tasks) == 0 {
				log.Print("No tasks were found on request!")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			responseTasksArayJSON, err := json.Marshal(responseTasksArray)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseTasksArayJSON)
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary		Get the version of the DataManager module
// @Tags		Version
// @Accept		json
// @Produce		json
// @Success		200 {object} ResponseVersionT
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/version [get]
// @Security	ApiKeyAuth
func getVersion(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			responseVersion := &ResponseVersionT{
				Version:    Version,
				APIVersion: versionAPI,
			}

			responseVersionJSON, err := json.Marshal(responseVersion)
			if err != nil {
				log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(responseVersionJSON)
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary		Save monitoring data in the database
// @Tags		Data
// @Accept		json
// @Produce		json
// @Param		data body RequestRawDataSaveArrayT true "Input structure"
// @Success		200 {object} ResponseRawDataSaveArrayT "OK, The status of 200 is always returned, the data is successfully preserved if in the return structure status = 1"
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/data/save [post]
// @Security	ApiKeyAuth
func postRawDataSave(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	ctx := context.Background()
	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var requestRawDataSaveArray RequestRawDataSaveArrayT
			var responseRawDataSaveArray ResponseRawDataSaveArrayT

			if err := json.Unmarshal(b, &requestRawDataSaveArray); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				for index := range requestRawDataSaveArray.Data {
					var responseRawDataSave ResponseRawDataSaveT
					responseRawDataSave.SpaceID = requestRawDataSaveArray.Data[index].SpaceID
					responseRawDataSave.Metric = requestRawDataSaveArray.Data[index].Metric

					spaceStatus := 0
					if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, requestRawDataSaveArray.Data[index].SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
						log.Print(err)
					}
					if spaceStatus == 1 {
						log.Print("The raw_data table is used...")
						err := Session.Query(`
						BEGIN BATCH
							INSERT INTO raw_data01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, ?)
							INSERT INTO raw_data02 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, ?)
							INSERT INTO raw_data03 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, ?)
							INSERT INTO space_metric_data (space_id, metric) VALUES (?, ?)
							INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'data', 0)
						APPLY BATCH
						`,
							requestRawDataSaveArray.Data[index].Metric, requestRawDataSaveArray.Data[index].Value,
							requestRawDataSaveArray.Data[index].EventTime, requestRawDataSaveArray.Data[index].SpaceID, time.Now().Format("2006.01"),
							requestRawDataSaveArray.Data[index].Object,
							requestRawDataSaveArray.Data[index].Metric, requestRawDataSaveArray.Data[index].Value,
							requestRawDataSaveArray.Data[index].EventTime, requestRawDataSaveArray.Data[index].SpaceID, time.Now().Format("2006.01"),
							requestRawDataSaveArray.Data[index].Object,
							requestRawDataSaveArray.Data[index].Metric, requestRawDataSaveArray.Data[index].Value,
							requestRawDataSaveArray.Data[index].EventTime, requestRawDataSaveArray.Data[index].SpaceID, time.Now().Format("2006.01"),
							requestRawDataSaveArray.Data[index].Object,
							requestRawDataSaveArray.Data[index].SpaceID, requestRawDataSaveArray.Data[index].Metric,
							requestRawDataSaveArray.Data[index].SpaceID, time.Now().Format("2006.01")).WithContext(ctx).Exec()
						if err != nil {
							log.Print(err)
							responseRawDataSave.Status = 0
						} else {
							responseRawDataSave.Status = 1
						}
					} else {
						log.Printf("The space of %v is blocked, the recording is impossible!", requestRawDataSaveArray.Data[index].SpaceID)
						responseRawDataSave.Status = 0
					}
					responseRawDataSaveArray.Data = append(responseRawDataSaveArray.Data, responseRawDataSave)
				}

				responseRawDataSaveJSON, err := json.Marshal(responseRawDataSaveArray)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(responseRawDataSaveJSON)
				}
				w.WriteHeader(http.StatusOK)
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary		Save text monitoring data in the database
// @Tags		Data
// @Accept		json
// @Produce		json
// @Param		data body RequestRawTextSaveArrayT true "Input structure"
// @Success		200 {object} ResponseRawDataSaveArrayT "OK, The status of 200 is always returned, the data is successfully preserved if in the return structure status = 1"
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/text/save [post]
// @Security	ApiKeyAuth
func postRawTextSave(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	ctx := context.Background()
	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var requestRawTextSaveArray RequestRawTextSaveArrayT
			var responseRawDataSaveArray ResponseRawDataSaveArrayT

			if err := json.Unmarshal(b, &requestRawTextSaveArray); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				for index := range requestRawTextSaveArray.Data {
					var responseRawDataSave ResponseRawDataSaveT
					responseRawDataSave.SpaceID = requestRawTextSaveArray.Data[index].SpaceID
					responseRawDataSave.Metric = requestRawTextSaveArray.Data[index].Metric

					spaceStatus := 0
					if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, requestRawTextSaveArray.Data[index].SpaceID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
						log.Print(err)
					}
					if spaceStatus == 1 {

						log.Print("The raw_text01, 02 table is used...")
						err := Session.Query(`
						BEGIN BATCH
							INSERT INTO raw_text01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, ?)
							INSERT INTO raw_text02 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, ?)
							INSERT INTO space_metric_text (space_id, metric) VALUES (?, ?)
							INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'text', 0)
						APPLY BATCH`,
							requestRawTextSaveArray.Data[index].Metric, requestRawTextSaveArray.Data[index].Value,
							requestRawTextSaveArray.Data[index].EventTime, requestRawTextSaveArray.Data[index].SpaceID, time.Now().Format("2006.01"),
							requestRawTextSaveArray.Data[index].Object,
							requestRawTextSaveArray.Data[index].Metric, requestRawTextSaveArray.Data[index].Value,
							requestRawTextSaveArray.Data[index].EventTime, requestRawTextSaveArray.Data[index].SpaceID, time.Now().Format("2006.01"),
							requestRawTextSaveArray.Data[index].Object,
							requestRawTextSaveArray.Data[index].SpaceID, requestRawTextSaveArray.Data[index].Metric,
							requestRawTextSaveArray.Data[index].SpaceID, time.Now().Format("2006.01")).WithContext(ctx).Exec()
						if err != nil {
							log.Print(err)
							responseRawDataSave.Status = 0
						} else {
							responseRawDataSave.Status = 1
						}
					} else {
						log.Printf("The space of %v is blocked, the recording is impossible!", requestRawTextSaveArray.Data[index].SpaceID)
						responseRawDataSave.Status = 0
					}
					responseRawDataSaveArray.Data = append(responseRawDataSaveArray.Data, responseRawDataSave)
				}

				responseRawDataSaveJSON, err := json.Marshal(responseRawDataSaveArray)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(responseRawDataSaveJSON)
				}
				w.WriteHeader(http.StatusOK)
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary		Get operational monitoring data from the database
// @Tags		Data
// @Accept		json
// @Produce		json
// @Param		data body RequestRawDataGetT true "Input structure"
// @Success		200 {object} ResponseRawDataGetArrayT "OK"
// @Success		204 "NoContent"
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/data/get [post]
// @Security	ApiKeyAuth
func postRawDataGet(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	ctx := context.Background()
	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var requestRawDataGet RequestRawDataGetT
			var responseRawDataGetArray ResponseRawDataGetArrayT

			if err := json.Unmarshal(b, &requestRawDataGet); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				if requestRawDataGet.SpaceID == "" || requestRawDataGet.Metric == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				log.Print("The raw_data01 table is used...")
				scanner := Session.Query(`SELECT metric, value, create_time, event_time, status, object FROM raw_data01 WHERE space_id = ? AND synt_key = ? AND event_time > ? LIMIT 1000`,
					requestRawDataGet.SpaceID, time.Now().Format("2006.01"), makeTimestamp()-24*60*60*1000).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var metric string
					var value float32
					var createtime int64
					var eventtime int64
					var status int
					var object string
					err := scanner.Scan(&metric, &value, &createtime, &eventtime, &status, &object)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						if requestRawDataGet.Metric == metric {
							responseRawDataGet := &ResponseRawDataGetT{
								Metric:     metric,
								Value:      value,
								CreateTime: createtime,
								EventTime:  eventtime,
								Status:     status,
								SpaceID:    requestRawDataGet.SpaceID,
								Object:     object,
							}
							responseRawDataGetArray.Data = append(responseRawDataGetArray.Data, *responseRawDataGet)
						}
					}
				}
				if err := scanner.Err(); err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(responseRawDataGetArray.Data) == 0 {
					log.Print("No records were found on request!")
					w.WriteHeader(http.StatusNoContent)
					return
				}

				responseRawDataGetArayJSON, err := json.Marshal(responseRawDataGetArray)
				if err != nil {
					log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(responseRawDataGetArayJSON)
				}
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

// @Summary		Get text data from the database
// @Tags		Data
// @Accept		json
// @Produce		json
// @Param		data body RequestRawDataGetT true "Input structure"
// @Success		200 {object} ResponseRawTextGetArrayT "OK"
// @Success		204 "NoContent"
// @Failure     403 "Forbidden, The monitoring module key is incorrect, or not registered"
// @Failure     500 "Internal server error, The internal error of the service"
// @Router		/text/get [post]
// @Security	ApiKeyAuth
func postRawTextGet(w http.ResponseWriter, r *http.Request) {
	versionAPI := chi.URLParam(r, "version")

	ctx := context.Background()
	if registrationCheck(r.Header.Get("api_key")) {
		if versionAPI == "1" {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var requestRawDataGet RequestRawDataGetT
			var responseRawTextGetArray ResponseRawTextGetArrayT

			if err := json.Unmarshal(b, &requestRawDataGet); err != nil {
				log.Print("JSON UNMARSHAL ERROR (" + err.Error() + ")!")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				if requestRawDataGet.SpaceID == "" || requestRawDataGet.Metric == "" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				log.Print("The raw_text01 table is used...")
				scanner := Session.Query(`SELECT metric, value, create_time, event_time, status, object FROM raw_text01 WHERE space_id = ? AND synt_key = ? AND event_time > ? LIMIT 1000`,
					requestRawDataGet.SpaceID, time.Now().Format("2006.01"), makeTimestamp()-24*60*60*1000).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
				for scanner.Next() {
					var metric string
					var value string
					var create_time int64
					var eventtime int64
					var status int
					var object string
					err := scanner.Scan(&metric, &value, &create_time, &eventtime, &status, &object)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						if requestRawDataGet.Metric == metric {
							responseRawTextGet := &ResponseRawTextGetT{
								Metric:     metric,
								Value:      value,
								CreateTime: create_time,
								EventTime:  eventtime,
								Status:     status,
								SpaceID:    requestRawDataGet.SpaceID,
								Object:     object,
							}
							responseRawTextGetArray.Data = append(responseRawTextGetArray.Data, *responseRawTextGet)
						}
					}
				}
				if err := scanner.Err(); err != nil {
					log.Print(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(responseRawTextGetArray.Data) == 0 {
					log.Print("No records were found on request!")
					w.WriteHeader(http.StatusNoContent)
					return
				}

				responseRawTextGetArayJSON, err := json.Marshal(responseRawTextGetArray)
				if err != nil {
					log.Print("JSON MARSHAL ERROR (" + err.Error() + ")!")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(responseRawTextGetArayJSON)
				}
			}
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func registrationCheck(id string) bool {
	ctx := context.Background()
	scanner := Session.Query(`SELECT status FROM registration WHERE id = ?`, id).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
	for scanner.Next() {
		var status int

		err := scanner.Scan(&status)
		if err != nil {
			return false
		} else {
			if status == 1 {
				return true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print(err)
		return false
	}
	return false
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func zbxServer() {
	zbxType := AppConfig.ZBX.TYPE
	zbxURL := AppConfig.ZBX.HOST + ":" + AppConfig.ZBX.PORT

	AgentStart = make(map[string]bool)
	ZBXAgentMap = make(map[string]string)
	var mutex01 sync.Mutex

	// TLS for ZBX server (In development)
	/*
		caCertPEM, err := os.ReadFile("ca.pem")
		if err != nil {
			log.Fatal("Failed to open CA certificate!")
		}
		ca := x509.NewCertPool()
		statusCA := ca.AppendCertsFromPEM(caCertPEM)
		if !statusCA {
			panic("Failed to parse CA certificate!")
		}
	*/

	/*
		crt, err := tls.LoadX509KeyPair("scylla.crt", "scylla.key")
		if err != nil {
			log.Fatalf("ZBX Server: Loadkeys: %s", err)
		}
		config := tls.Config{
			Certificates:       []tls.Certificate{crt},
			ClientAuth:         tls.RequireAnyClientCert,
			InsecureSkipVerify: true,
			RoorCAs:          ca,
		}
		config.Rand = crand.Reader
	*/

	// Listen for incoming connections
	lsn, err := net.Listen(zbxType, zbxURL)
	// lsn, err := tls.Listen(ZBX_CONN_TYPE, ZBX_CONN_URL, &config)

	if err != nil {
		log.Print("Error listening ZBX: ", err.Error())
		os.Exit(1)
	}

	// Close the listener when this application closes
	defer lsn.Close()

	log.Print("ZBX service is launched and available on " + zbxURL + "... OK")
	for {
		conn, err := lsn.Accept()

		if err != nil {
			log.Print("Error ZBX accepting connection: ", err.Error())
			os.Exit(1)
		}

		connTLS, status := conn.(*tls.Conn)
		if status {
			log.Print("ZBX TLS... OK")
			state := connTLS.ConnectionState()
			log.Printf("Version: %x", state.Version)
			log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
			log.Printf("DidResume: %t", state.DidResume)
			log.Printf("CipherSuite: %x", state.CipherSuite)
			log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)
		}

		go zbxHandleRequest(conn, &mutex01)
	}
}

func zbxHandleRequest(conn net.Conn, mutex01 *sync.Mutex) {
	defer conn.Close()
	ctx := context.Background()

	var requestAgentType RequestAgentTypeT
	var requestAgent RequestAgentT
	var responseServerData ResponseServerDataT
	var responseServerDataSlice []ResponseServerDataT
	var responseServer ResponseServerT
	var requestAgentMetrics RequestAgentMetricsT
	var responseServerMetrics ResponsServerMetricsT

	hdr := make([]byte, 13)
	tmp := make([]byte, 1024)
	buf := make([]byte, 0)

	log.Print(strings.Split(conn.RemoteAddr().String(), ":")[0])

	sizeRead, err := conn.Read(hdr)
	if err != nil {
		log.Print(err)
	} else {
		log.Print("ZBX Header read: ", sizeRead)

		protocolFlag := int(hdr[4])
		log.Print("ZBX Protocol FLAG: ", protocolFlag)
		protocolSize := (hdr[5:12])
		var protocolSizeInt uint32
		hdr01 := bytes.NewReader(protocolSize)
		err = binary.Read(hdr01, binary.LittleEndian, &protocolSizeInt)
		if err != nil {
			log.Print("Binary read failed: ", err)
		} else {
			buf = append(buf, hdr[:13]...)

			var dataSize uint32
			dataSize = 0
			for {
				if dataSize == protocolSizeInt+13 {
					break
				}
				sizeRead, err := conn.Read(tmp)
				if err != nil {
					log.Print("Error reading: ", err.Error())
					break
				}
				log.Print("Read: ", sizeRead)
				buf = append(buf, tmp[:sizeRead]...)
				dataSize = uint32(len(buf))
			}

			sizeRead := len(buf)
			log.Print("END Read: ", sizeRead)
			protocolData := (buf[13 : 13+protocolSizeInt])

			uncompressData, err := zbxUncompress(protocolData, protocolSizeInt, protocolFlag)
			if err != err {
			} else {
				if err := json.Unmarshal(uncompressData, &requestAgentType); err != nil {
					log.Print(err)
				} else {
					log.Print(requestAgentType.Request)
					if requestAgentType.Request == "active checks" {
						if err := json.Unmarshal(uncompressData, &requestAgent); err != nil {
							log.Print(err)
						} else {
							if registrationCheck(requestAgent.Host) {
								log.Print(requestAgent)

								scanner := Session.Query(`SELECT object, metric, status, int_id, interval FROM task WHERE module_id = ?`, requestAgent.Host).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
								for scanner.Next() {
									var object string
									var metric string
									var status int
									var intID int64
									var interval int64
									err := scanner.Scan(&object, &metric, &status, &intID, &interval)
									if err != nil {
										log.Print(err)
										return
									} else {
										if interval == 0 {
											interval = 10
										}
										if status == 1 && object == "zbxagent" {
											responseServerData.Itemid = intID
											responseServerData.Key = metric
											responseServerData.KeyOrig = metric
											responseServerData.Delay = fmt.Sprintf("%ds", interval)
											responseServerData.Lastlogsize = 0
											responseServerData.Lastlogsize = 0
											responseServerDataSlice = append(responseServerDataSlice, responseServerData)
										}
									}
								}
								responseServer.Response = "success"
								responseServer.Data = responseServerDataSlice

								responseServerJSON, err := json.Marshal(responseServer)
								if err != nil {
									log.Print(err)
								} else {
									log.Print(string(responseServerJSON))

									var buf02 bytes.Buffer
									flags := ZBXTCPProtocol
									if ZBXCompress {
										z := zlib.NewWriter(&buf02)
										if _, err = z.Write(responseServerJSON); err != nil {
											log.Print(err)
										}
										z.Close()
										flags |= ZBXZLibCompress
									} else {
										buf02.Write(responseServerJSON)
									}

									var buf03 bytes.Buffer
									buf03.Grow(buf02.Len() + ZBXHeaderSize)
									buf03.Write([]byte{'Z', 'B', 'X', 'D', flags})
									if err = binary.Write(&buf03, binary.LittleEndian, uint32(buf02.Len())); nil != err {
										log.Print(err)
									}
									if err = binary.Write(&buf03, binary.LittleEndian, uint32(len(responseServerJSON))); nil != err {
										log.Print(err)
									}
									buf03.Write(buf02.Bytes())
									sizeWrite, err := conn.Write(buf03.Bytes())
									if err != nil {
										log.Print(err)
									} else {
										log.Print(requestAgent.Host)
										AgentStart[requestAgent.Host] = true
										log.Print("Write: ", sizeWrite)
									}
								}
							} else {
								log.Print("Request from an unregistered agent!")
							}
						}
					} else if requestAgentType.Request == "agent data" {
						if err := json.Unmarshal(uncompressData, &requestAgentMetrics); err != nil {
							log.Print(err)
						} else {
							if registrationCheck(requestAgentMetrics.Host) {
								if AgentStart[requestAgentMetrics.Host] {
									log.Print(requestAgentMetrics)
									for _, value := range requestAgentMetrics.Data {
										spaceIDDB := ""
										metricDB := ""
										dataTypeDB := ""
										scanner := Session.Query(`SELECT space_id, status, metric, object, int_id, data_type FROM task WHERE module_id = ?`, requestAgentMetrics.Host).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
										for scanner.Next() {
											var spaceID string
											var status int
											var metric string
											var object string
											var intID int64
											var dataType string
											err := scanner.Scan(&spaceID, &status, &metric, &object, &intID, &dataType)
											if err != nil {
												log.Print(err)
											} else {
												if status == 1 && object == "zbxagent" && intID == int64(value.ItemID) {
													spaceIDDB = spaceID
													metricDB = metric
													dataTypeDB = dataType
												}
											}
										}

										if spaceIDDB != "" && metricDB != "" {
											if dataTypeDB == "float" {
												valueFloat, _ := strconv.ParseFloat(value.Value, 64)
												eventTimeFloat := (float64(value.Clock) * 1000) + (float64(value.NS) / 1000000)
												eventTime := int(math.Round(eventTimeFloat))

												spaceStatus := 0
												if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, spaceIDDB).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
													log.Print(err)
												}
												if spaceStatus == 1 {
													log.Print("The raw_data01, 02, 03 table is used...")
													err := Session.Query(`
												BEGIN BATCH
													INSERT INTO raw_data01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, 'zbxagent')
													INSERT INTO raw_data02 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, 'zbxagent')
													INSERT INTO raw_data03 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, 'zbxagent')
													INSERT INTO space_metric_data (space_id, metric) VALUES (?, ?)
													INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'data', 0)
												APPLY BATCH
												`,
														metricDB, float32(valueFloat),
														eventTime, spaceIDDB, time.Now().Format("2006.01"),
														metricDB, float32(valueFloat),
														eventTime, spaceIDDB, time.Now().Format("2006.01"),
														metricDB, float32(valueFloat),
														eventTime, spaceIDDB, time.Now().Format("2006.01"),
														spaceIDDB, metricDB,
														spaceIDDB, time.Now().Format("2006.01")).WithContext(ctx).Exec()
													if err != nil {
														log.Print(err)
													} else {
														log.Print("DATA SAVE " + time.Now().Format("2006.01"))
														mutex01.Lock()
														ZBXAgentMap[strings.Split(conn.RemoteAddr().String(), ":")[0]] = requestAgentMetrics.Host
														mutex01.Unlock()
														go problem(spaceIDDB, requestAgentMetrics.Host, metricDB, int64(eventTime), float32(valueFloat), "", "zbxagent")
													}
												}
											} else {
												eventTime, _ := strconv.ParseInt(fmt.Sprintf("%d%s", value.Clock, "000"), 10, 64)

												spaceStatus := 0
												if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, spaceIDDB).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
													log.Print(err)
												}
												if spaceStatus == 1 {
													log.Print("The raw_text01, 02 table is used...")
													err := Session.Query(`
													BEGIN BATCH
														INSERT INTO raw_text01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, 'zbxagent')
														INSERT INTO raw_text02 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), ?, ?, ?, 'zbxagent')
														INSERT INTO space_metric_text (space_id, metric) VALUES (?, ?)
														INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'text', 0)
													APPLY BATCH
													`,
														metricDB, value.Value,
														eventTime, spaceIDDB, time.Now().Format("2006.01"),
														metricDB, value.Value,
														eventTime, spaceIDDB, time.Now().Format("2006.01"),
														spaceIDDB, metricDB,
														spaceIDDB, time.Now().Format("2006.01")).WithContext(ctx).Exec()
													if err != nil {
														log.Print(err)
													} else {
														log.Print("TEXT SAVE")
													}
												}
											}
										}
									}

									responseServerMetrics.Response = "success"
									responseServerMetrics.Info = fmt.Sprintf("processed: %d; failed: 0; total: %d", len(requestAgentMetrics.Data), len(requestAgentMetrics.Data))
									responseServerJSON, err := json.Marshal(responseServerMetrics)
									log.Print(string(responseServerJSON))
									if err != nil {
										log.Print(err)
									} else {
										var buf04 bytes.Buffer
										flags := ZBXTCPProtocol
										if ZBXCompress {
											z := zlib.NewWriter(&buf04)
											if _, err = z.Write(responseServerJSON); err != nil {
												log.Print(err)
											}
											z.Close()
											flags |= ZBXZLibCompress
										} else {
											buf04.Write(responseServerJSON)
										}

										var buf05 bytes.Buffer
										buf05.Grow(buf04.Len() + ZBXHeaderSize)
										buf05.Write([]byte{'Z', 'B', 'X', 'D', flags})
										if err = binary.Write(&buf05, binary.LittleEndian, uint32(buf04.Len())); nil != err {
											log.Print(err)
										}
										if err = binary.Write(&buf05, binary.LittleEndian, uint32(len(responseServerJSON))); nil != err {
											log.Print(err)
										}
										buf05.Write(buf04.Bytes())
										sizeWrite, err := conn.Write(buf05.Bytes())
										if err != nil {
											log.Print(err)
										} else {
											log.Print("Write: ", sizeWrite)
										}
									}
								}
							} else {
								log.Print("Request from an unregistered agent")
							}
						}
					}
				}
			}
		}
	}
}

func zbxUncompress(protocolData []byte, protocolSizeInt uint32, protocolFlag int) ([]byte, error) {
	if protocolFlag == 3 || protocolFlag == 2 {
		var b bytes.Buffer

		b.Grow(int(protocolSizeInt))
		z, err := zlib.NewReader(bytes.NewReader(protocolData))
		if err != nil {
			log.Printf("Unable to uncompress message (1): '%s'\n", err)
			return nil, err
		}
		len, err := b.ReadFrom(z)
		z.Close()
		if err != nil {
			log.Printf("Unable to uncompress message (2): '%s'\n", err)
			return nil, err
		}
		if len != int64(protocolSizeInt) {
			log.Printf("Uncompressed message size %d instead of expected %d\n", len, protocolSizeInt)
		}
		return b.Bytes(), nil
	} else {
		return protocolData, nil
	}
}

func problem(spaceID string, modulID string, metric string, eventTime int64, valueF float32, valueS string, object string) {
	ctx := context.Background()

	scanner01 := Session.Query(`SELECT critical, warning, data_type FROM task WHERE module_id = ? AND space_id = ? AND object = ? AND metric = ?`, modulID, spaceID, object, metric).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
	for scanner01.Next() {
		var critical string
		var warning string
		var dataType string

		err := scanner01.Scan(&critical, &warning, &dataType)
		if err != nil {
			log.Print(err)
			return
		} else {
			status := "NORMAL"
			var valueAllType string
			if dataType == "float" {
				valueAllType = strconv.FormatFloat(float64(valueF), 'f', -1, 32)

				if critical != "" && warning != "" {
					criticalFloat, err01 := strconv.ParseFloat(critical, 32)
					if err01 != nil {
						log.Print(err01)
						return
					}
					warningFloat, err02 := strconv.ParseFloat(warning, 32)
					if err02 != nil {
						log.Print(err02)
					}
					if criticalFloat >= warningFloat {
						if valueF >= float32(criticalFloat) {
							log.Print("CRITICAL")
							status = "CRITICAL"
						} else {
							if valueF >= float32(warningFloat) {
								log.Print("WARNING")
								status = "WARNING"
							}
						}
					} else {
						if valueF <= float32(criticalFloat) {
							log.Print("CRITICAL")
							status = "CRITICAL"
						} else {
							if valueF <= float32(warningFloat) {
								log.Print("WARNING")
								status = "WARNING"
							}
						}
					}
				}
			} else {
				valueAllType = valueS
			}
			scanner02 := Session.Query(`SELECT status, event_time_start FROM problem WHERE module_id = ? AND space_id = ? AND metric = ?`, modulID, spaceID, metric).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
			countProblem := 0
			for scanner02.Next() {
				var statusDB string
				var eventTimeStartDB int64
				err := scanner02.Scan(&statusDB, &eventTimeStartDB)
				if err != nil {
					log.Print(err)
					return
				} else {
					if status != statusDB {
						err := Session.Query(`UPDATE problem SET event_time = ?, event_time_start = ?, status = ?, value = ? WHERE space_ID = ? AND module_id = ? AND metric = ?
						`, eventTime, eventTime, status, valueAllType, spaceID, modulID, metric).WithContext(ctx).Exec()
						if err != nil {
							log.Print(err)
							return
						} else {
							log.Print("PROBLEM UPDATE")
						}
					} else {
						log.Print("Problem time: ", eventTime-eventTimeStartDB)
						if statusDB == "NORMAL" && (eventTime-eventTimeStartDB) > 3600*1000 { // Добавить параметризацию времени хранения статуса восстановления (NORMAL)
							err := Session.Query(`DELETE FROM problem WHERE space_ID = ? AND module_id = ? AND metric = ?`, spaceID, modulID, metric).WithContext(ctx).Exec()
							if err != nil {
								log.Print(err)
								return
							} else {
								log.Print("PROBLEM DELETE (NORMAL)")
							}
						} else {
							err := Session.Query(`UPDATE problem SET event_time = ?, value = ? WHERE space_ID = ? AND module_id = ? AND metric = ?
							`, eventTime, valueAllType, spaceID, modulID, metric).WithContext(ctx).Exec()
							if err != nil {
								log.Print(err)
								return
							} else {
								log.Print("PROBLEM UPDATE")
							}
						}
					}
				}
				countProblem++
			}
			if countProblem == 0 && status != "NORMAL" {
				err := Session.Query(`INSERT INTO problem (space_id, module_id, metric, event_time, event_time_start, status, value) VALUES (?, ?, ?, ?, ?, ?, ?)
				`, spaceID, modulID, metric, eventTime, eventTime, status, valueAllType).WithContext(ctx).Exec()
				if err != nil {
					log.Print(err)
					return
				} else {
					log.Print("PROBLEM INSERT")
				}
			}
		}
	}
}

func workersRun() {
	var workerPool = AppConfig.WORKER.POOL
	log.Print("WORKERS START... ", workerPool, " OK")

	jobs := make(chan JobT, workerPool)
	defer close(jobs)

	for worker := 1; worker <= workerPool; worker++ {
		go zbxClient(worker, jobs)
	}

	ticker := time.NewTicker(time.Duration(AppConfig.WORKER.ACTIME) * time.Second)
	for ; true; <-ticker.C {
		if len(ZBXAgentMap) > 0 {
			log.Print("Summer agents for passive verification: ", len(ZBXAgentMap))
			for key, value := range ZBXAgentMap {
				var jobData JobT
				jobData.key = key
				jobData.value = value
				jobs <- jobData
			}
		}
	}
}

func zbxClient(worker int, jobs <-chan JobT) {
	log.Print("START WORKER ", worker, "... OK")
	defer log.Print("STOP WORKER ", worker, "... OK")
	for j := range jobs {
		address := j.key
		moduleID := j.value
		ctx := context.Background()

		log.Print("START PASSIVE CHECK, WORKER ", worker, " (", moduleID, "), ", address, "...")

		scanner := Session.Query(`SELECT space_id, status, metric, object, data_type FROM task WHERE module_id = ?`, moduleID).WithContext(ctx).Consistency(ConsistencyRead).Iter().Scanner()
		for scanner.Next() {
			var metricValue string

			var spaceID string
			var status int
			var metric string
			var object string
			var dataType string
			err := scanner.Scan(&spaceID, &status, &metric, &object, &dataType)
			if err != nil {
				log.Print(err)
			} else {
				if status == 1 && object == "zbxagentpassive" {
					conn, err := net.Dial("tcp", address+":10050")
					if err != nil {
						log.Print("ZBX Agent NOT AVAILABLE!")
					} else {
						var buf01 bytes.Buffer
						flags := ZBXTCPProtocol
						if ZBXCompress {
							z := zlib.NewWriter(&buf01)
							if _, err := z.Write([]byte(metric)); err != nil {
								log.Print(err)
							}
							z.Close()
							flags |= ZBXZLibCompress
						} else {
							buf01.Write([]byte(metric))
						}

						var buf02 bytes.Buffer
						buf02.Grow(buf01.Len() + ZBXHeaderSize)
						buf02.Write([]byte{'Z', 'B', 'X', 'D', flags})
						if err := binary.Write(&buf02, binary.LittleEndian, uint32(buf01.Len())); nil != err {
							log.Print(err)
						}
						if err := binary.Write(&buf02, binary.LittleEndian, uint32(len([]byte(metric)))); nil != err {
							log.Print(err)
						}
						buf02.Write(buf01.Bytes())
						sizeWrite, err := conn.Write(buf02.Bytes())
						if err != nil {
							log.Print(err)
						} else {
							log.Print("Write: ", sizeWrite)
						}

						hdr := make([]byte, 13)
						tmp := make([]byte, 1024)
						buf := make([]byte, 0)

						log.Print(conn.RemoteAddr())

						sizeRead, err := conn.Read(hdr)
						if err != nil {
							log.Print(err)
						} else {
							log.Print("ZBX Header read: ", sizeRead)

							protocolFlag := int(hdr[4])
							log.Print("ZBX Protocol FLAG: ", protocolFlag)
							protocolSize := (hdr[5:12])
							var protocolSizeInt uint32
							hdr01 := bytes.NewReader(protocolSize)
							err = binary.Read(hdr01, binary.LittleEndian, &protocolSizeInt)
							if err != nil {
								log.Print("Binary read failed: ", err)
							} else {
								buf = append(buf, hdr[:13]...)

								var dataSize uint32
								dataSize = 0
								for {
									if dataSize == protocolSizeInt+13 {
										break
									}
									sizeRead, err := conn.Read(tmp)
									if err != nil {
										log.Print("Error reading: ", err.Error())
										break
									}
									log.Print("Read: ", sizeRead)
									buf = append(buf, tmp[:sizeRead]...)
									dataSize = uint32(len(buf))
								}

								sizeRead := len(buf)
								log.Print("END Read: ", sizeRead)
								protocolData := (buf[13 : 13+protocolSizeInt])

								uncompressData, err := zbxUncompress(protocolData, protocolSizeInt, protocolFlag)
								if err != err {
								} else {
									log.Print("ZBX PASSIVE RESPONSE: ", string(uncompressData))
									metricValue = string(uncompressData)
								}
							}
						}
					}

					log.Print("PASSIVE AGENT ", moduleID, " DATA SAVE")

					if spaceID != "" && metric != "" {
						if dataType == "float" {
							valueFloat, _ := strconv.ParseFloat(metricValue, 64)

							spaceStatus := 0
							if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
								log.Print(err)
							}
							if spaceStatus == 1 {
								log.Print("The raw_data01, 02, 03 table is used...")
								err := Session.Query(`
						BEGIN BATCH
							INSERT INTO raw_data01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), toTimestamp(now()), ?, ?, 'zbxagentpassive')
							INSERT INTO raw_data02 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), toTimestamp(now()), ?, ?, 'zbxagentpassive')
							INSERT INTO raw_data03 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), toTimestamp(now()), ?, ?, 'zbxagentpassive')
							INSERT INTO space_metric_data (space_id, metric) VALUES (?, ?)
							INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'data', 0)
						APPLY BATCH
						`,
									metric, float32(valueFloat),
									spaceID, time.Now().Format("2006.01"),
									metric, float32(valueFloat),
									spaceID, time.Now().Format("2006.01"),
									metric, float32(valueFloat),
									spaceID, time.Now().Format("2006.01"),
									spaceID, metric,
									spaceID, time.Now().Format("2006.01")).WithContext(ctx).Exec()
								if err != nil {
									log.Print(err)
								} else {
									log.Print("PASSIVE DATA SAVE")
									go problem(spaceID, moduleID, metric, makeTimestamp(), float32(valueFloat), "", "zbxagentpassive")
								}
							}
						} else {
							spaceStatus := 0
							if err := Session.Query(`SELECT status FROM space WHERE id = ? LIMIT 1`, spaceID).WithContext(ctx).Consistency(ConsistencyRead).Scan(&spaceStatus); err != nil {
								log.Print(err)
							}
							if spaceStatus == 1 {
								log.Print("The raw_data01, 02, 03 table is used...")
								err := Session.Query(`
							BEGIN BATCH
								INSERT INTO raw_text01 (metric, value, status, create_time, event_time, space_id, synt_key, object) VALUES (?, ?, 1, toTimestamp(now()), toTimestamp(now()), ?, ?, 'zbxagentpassive')
								INSERT INTO space_metric_text (space_id, metric) VALUES (?, ?)
								INSERT INTO raw_key (space_id, synt_key, type_data, flag) values (?, ?, 'text', 0)
							APPLY BATCH
							`,
									metric, metricValue,
									spaceID, time.Now().Format("2006.01"),
									spaceID, metric,
									spaceID, time.Now().Format("2006.01")).WithContext(ctx).Exec()
								if err != nil {
									log.Print(err)
								} else {
									log.Print("PASSIVE TEXT SAVE")
								}
							}
						}
					}
				}
			}
		}
	}
}
