package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	// TAppConfigHTTP Параметры запуска HTTP сервера
	TAppConfigHTTP struct {
		HOST         string `json:"HOST"`
		PORT         string `json:"PORT"`
		READTIMEOUT  int    `json:"READTIMEOUT"`
		WRITETIMEOUT int    `json:"WRITETIMEOUT"`
		RL           int    `json:"RL"`
		TLS          bool   `json:"TLS"`
		CERTPATH     string `json:"CERTPATH"`
		KEYPATH      string `json:"KEYPATH"`
		CAPATH       string `json:"CAPATH"`
	}
	TAppConfigWorker struct {
		POOL   int `json:"POOL"`
		ACTIME int `json:"ACTIME"`
	}

	/*
		Consistensy:
		Any
		One
		Two
		Three
		Quorum
		All
		LocalQuorum
		EachQuorum
		LocalOn
	*/
	TAppConfigDB struct {
		HOSTS            []string `json:"HOSTS"`
		USERNAME         string   `json:"USERNAME"`
		PASSWORD         string   `json:"PASSWORD"`
		CERTPATH         string   `json:"CERTPATH"`
		KEYPATH          string   `json:"KEYPATH"`
		CAPATH           string   `json:"CAPATH"`
		KEYSPACE         string   `json:"KEYSPACE"`
		CONSISTENCY      string   `json:"CONSISTENCY"`
		CONSISTENCYREAD  string   `json:"CONSISTENCYREAD"`
		TIMEOUT          int      `json:"TIMEOUT"`
		TLS              bool     `json:"TLS"`
		HOSTVERIFICATION bool     `json:"HOSTVERIFICATION"`
	}

	// TAppConfigZBX Параметры запуска ZBX сервера
	TAppConfigZBX struct {
		HOST string `json:"HOST"`
		PORT string `json:"PORT"`
		TYPE string `json:"TYPE"`
	}

	// TAppConfigSWG Параметры Swagger
	TAppConfigSWG struct {
		TITLE       string `json:"TITLE"`
		DESCRIPTION string `json:"DESCRIPTION"`
		VERSION     string `json:"VERSION"`
		HOST        string `json:"HOST"`
		BASEPATH    string `json:"BASEPATH"`
		SCHEME      string `json:"SCHEME"`
	}

	// TAppConfig Общая структура параметров сервиса
	TAppConfig struct {
		HTTP   TAppConfigHTTP   `json:"HTTP"`
		WORKER TAppConfigWorker `json:"WORKER"`
		DB     TAppConfigDB     `json:"DB"`
		ZBX    TAppConfigZBX    `json:"ZBX"`
		SWG    TAppConfigSWG    `json:"SWG"`
	}
)

var (
	// AppConfig Структура TAppConfig
	AppConfig TAppConfig
)

// LoadConfig Загрузка конфигурации, для Docker версии конфиг собирается из переменного окружения
func (appCfg *TAppConfig) LoadConfig() (err error) {

	log.Print("Download configuration from the file...")
	data, err := os.ReadFile(ConfigFile)
	if err == nil {
		err = json.Unmarshal(data, &appCfg)
		if err != nil {
			return err
		}
	} else {
		log.Print("Failed to read the configuration from the file (" + err.Error() + ")!")
	}

	/*
		Переменное окружение (для Docker версии):

		CHRDS_HTTP_HOST Адрес для входящих соединений API
		CHRDS_HTTP_PORT Порт для входящих соединений API
		CHRDS_HTTP_READTIMEOUT Read Timeout для HTTP сервера
		CHRDS_HTTP_WRITETIMEOUT Write Timeout для HTTP сервера
		CHRDS_HTTP_RL Requests Limit
		CHRDS_HTTP_TLS Включение TLS взаимодействия с http сервером (true/false)
		CHRDS_HTTP_CERTPATH Путь к сертификату для подключения к http серверу
		CHRDS_HTTP_KEYPATH Путь к приватному ключу для подключения к http серверу
		CHRDS_HTTP_CAPATH Путь к CA сертификату для подключения к http серверу

		CHRDS_WORKER_POOL Колличество воркеров запускаемых при старте для активных проверок со стороны сервера
		CHRDS_WORKER_ACTIME Интервал запуска активных проверок со стороны сервера в секундах

		CHRDS_DB_HOST Массив адресов для подключения к БД, разделитель запятая (127.0.0.1:19042,127.0.0.1:29042)
		CHRDS_DB_USERNAME Имя пользователя для подключения к БД
		CHRDS_DB_PASSWORD Пароль для подключения к БД
		CHRDS_DB_TLS Включение TLS взаимодействия с кластером БД (true/false)
		CHRDS_DB_CERTPATH Путь к сертификату для подключения к БД
		CHRDS_DB_KEYPATH Путь к приватному ключу для подключения к БД
		CHRDS_DB_CAPATH Путь к CA сертификату для подключения к БД
		CHRDS_DB_HOSTVERIFICATION Включение проверки HostVerification по сертификату при подключении к БД (true/false)
		CHRDS_DB_KEYSPACE Пространство ключей (KeySpace) для подключения
		CHRDS_DB_CONSISTENCY Параметр Consistency для данных в кластере БД
		CHRDS_DB_CONSISTENCYREAD Параметр Consistency для выборки данных из БД
		CHRDS_DB_TIMEOUT Значение времени ожидания выполнения запроса до принудительного завершения

		CHRDS_ZBX_HOST Адрес для входящих соединений ZBX сервера
		CHRDS_ZBX_PORT Порт для входящих соединений ZBX сервера
		CHRDS_ZBX_TYPE Протокол ZBX сервера

		CHRDS_SWG_TITLE Заголовок на странице Swagger
		CHRDS_SWG_DESCRIPTION Описание на странице Swagger
		CHRDS_SWG_VERSION Версия API на странице  Swagger
		CHRDS_SWG_HOST Полный адрес до API для Swagger
		CHRDS_SWG_BASEPATH Базовый путь до API для Swagger
		CHRDS_SWG_SCHEME Базовый протокол для Swagger
	*/

	log.Print("Loading configuration from alternating environment...")
	// Основной API
	var httpHostENV string
	httpHostENV, _ = os.LookupEnv("CHRDS_HTTP_HOST")
	if httpHostENV != "" {
		appCfg.HTTP.HOST = httpHostENV
	}
	var httpPortENV string
	httpPortENV, _ = os.LookupEnv("CHRDS_HTTP_PORT")
	if httpPortENV != "" {
		appCfg.HTTP.PORT = httpPortENV
	}
	var httpReadTimeoutENV string
	httpReadTimeoutENV, _ = os.LookupEnv("CHRDS_HTTP_READTIMEOUT")
	if httpReadTimeoutENV != "" {
		var httpReadTimeoutINT int
		httpReadTimeoutINT, err = strconv.Atoi(httpReadTimeoutENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_HTTP_READTIMEOUT (" + err.Error() + ")!")
		} else {
			appCfg.HTTP.READTIMEOUT = httpReadTimeoutINT
		}
	}
	var httpWriteTimeoutENV string
	httpWriteTimeoutENV, _ = os.LookupEnv("CHRDS_HTTP_READTIMEOUT")
	if httpWriteTimeoutENV != "" {
		var httpWriteTimeoutINT int
		httpWriteTimeoutINT, err = strconv.Atoi(httpWriteTimeoutENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_HTTP_WRITETIMEOUT (" + err.Error() + ")!")
		} else {
			appCfg.HTTP.WRITETIMEOUT = httpWriteTimeoutINT
		}
	}
	var httpRLENV string
	httpRLENV, _ = os.LookupEnv("CHRDS_HTTP_RL")
	if httpRLENV != "" {
		var httpRLINT int
		httpRLINT, err = strconv.Atoi(httpRLENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_HTTP_RL (" + err.Error() + ")!")
		} else {
			appCfg.HTTP.RL = httpRLINT
		}
	}
	var httpTLSENV string
	httpTLSENV, _ = os.LookupEnv("CHRDS_HTTP_TLS")
	if httpTLSENV != "" {
		boolValue, err := strconv.ParseBool(httpTLSENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_HTTP_TLS (" + err.Error() + ")!")
		} else {
			appCfg.HTTP.TLS = boolValue
		}
	}
	var httpCertPathENV string
	httpCertPathENV, _ = os.LookupEnv("CHRDS_HTTP_CERTPATH")
	if httpCertPathENV != "" {
		appCfg.HTTP.CERTPATH = httpCertPathENV
	}
	var httpKeyPathENV string
	httpKeyPathENV, _ = os.LookupEnv("CHRDS_HTTP_KEYPATH")
	if httpKeyPathENV != "" {
		appCfg.HTTP.KEYPATH = httpKeyPathENV
	}
	var httpCAPathENV string
	httpCAPathENV, _ = os.LookupEnv("CHRDS_HTTP_CAPATH")
	if httpCAPathENV != "" {
		appCfg.HTTP.CAPATH = httpCAPathENV
	}

	//Worker
	var workerPoolENV string
	workerPoolENV, _ = os.LookupEnv("CHRDS_WORKER_POOL")
	if workerPoolENV != "" {
		var workerPoolINT int
		workerPoolINT, err = strconv.Atoi(workerPoolENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_WORKER_POOL (" + err.Error() + ")!")
		} else {
			appCfg.WORKER.POOL = workerPoolINT
		}
	}
	var workerACTimeENV string
	workerACTimeENV, _ = os.LookupEnv("CHRDS_WORKER_ACTIME")
	if workerACTimeENV != "" {
		var workerACTimeINT int
		workerACTimeINT, err = strconv.Atoi(workerACTimeENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_WORKER_ACTIME (" + err.Error() + ")!")
		} else {
			appCfg.WORKER.ACTIME = workerACTimeINT
		}
	}

	// Подключение к БД
	var dbHostsENV string
	dbHostsENV, _ = os.LookupEnv("CHRDS_DB_HOST")
	if dbHostsENV != "" {
		dbHostsENV = strings.TrimSpace(dbHostsENV)       // Удаление возможных пробелов
		appCfg.DB.HOSTS = strings.Split(dbHostsENV, ",") // Разбиваем строку по символу ","
	}
	var dbUserNameENV string
	dbUserNameENV, _ = os.LookupEnv("CHRDS_DB_USERNAME")
	if dbUserNameENV != "" {
		appCfg.DB.USERNAME = dbUserNameENV
	}
	var dbPasswordENV string
	dbPasswordENV, _ = os.LookupEnv("CHRDS_DB_PASSWORD")
	if dbPasswordENV != "" {
		appCfg.DB.PASSWORD = dbPasswordENV
	}
	var dbCertPathENV string
	dbCertPathENV, _ = os.LookupEnv("CHRDS_DB_CERTPATH")
	if dbCertPathENV != "" {
		appCfg.DB.CERTPATH = dbCertPathENV
	}
	var dbKeyPathENV string
	dbKeyPathENV, _ = os.LookupEnv("CHRDS_DB_KEYPATH")
	if dbKeyPathENV != "" {
		appCfg.DB.KEYPATH = dbKeyPathENV
	}
	var dbCAPathENV string
	dbCAPathENV, _ = os.LookupEnv("CHRDS_DB_CAPATH")
	if dbCAPathENV != "" {
		appCfg.DB.CAPATH = dbCAPathENV
	}
	var dbKeySpaceENV string
	dbKeySpaceENV, _ = os.LookupEnv("CHRDS_DB_KEYSPACE")
	if dbKeySpaceENV != "" {
		appCfg.DB.KEYSPACE = dbKeySpaceENV
	}
	var dbConsistensyENV string
	dbConsistensyENV, _ = os.LookupEnv("CHRDS_DB_CONSISTENCY")
	if dbConsistensyENV != "" {
		appCfg.DB.CONSISTENCY = dbConsistensyENV
	}
	var dbConsistensyReadENV string
	dbConsistensyReadENV, _ = os.LookupEnv("CHRDS_DB_CONSISTENCYREAD")
	if dbConsistensyReadENV != "" {
		appCfg.DB.CONSISTENCYREAD = dbConsistensyReadENV
	}
	var dbTimeoutENV string
	dbTimeoutENV, _ = os.LookupEnv("CHRDS_DB_TIMEOUT")
	if dbTimeoutENV != "" {
		var dbTimeoutINT int
		dbTimeoutINT, err = strconv.Atoi(dbTimeoutENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_DB_TIMEOUT (" + err.Error() + ")!")
		} else {
			appCfg.DB.TIMEOUT = dbTimeoutINT
		}
	}
	var dbTLSENV string
	dbTLSENV, _ = os.LookupEnv("CHRDS_DB_TLS")
	if dbTLSENV != "" {
		boolValue, err := strconv.ParseBool(dbTLSENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_DB_TLS (" + err.Error() + ")!")
		} else {
			appCfg.DB.TLS = boolValue
		}
	}
	var dbHOSTVERIFICATIONENV string
	dbHOSTVERIFICATIONENV, _ = os.LookupEnv("CHRDS_DB_HOSTVERIFICATION")
	if dbHOSTVERIFICATIONENV != "" {
		boolValue, err := strconv.ParseBool(dbHOSTVERIFICATIONENV)
		if err != nil {
			log.Print("It was not possible to convert CHRDS_DB_HOSTVERIFICATION (" + err.Error() + ")!")
		} else {
			appCfg.DB.HOSTVERIFICATION = boolValue
		}
	}

	// ZBX сервер
	var zbxHostENV string
	zbxHostENV, _ = os.LookupEnv("CHRDS_ZBX_HOST")
	if zbxHostENV != "" {
		appCfg.ZBX.HOST = zbxHostENV
	}
	var zbxPortENV string
	zbxPortENV, _ = os.LookupEnv("CHRDS_ZBX_PORT")
	if zbxPortENV != "" {
		appCfg.ZBX.PORT = zbxPortENV
	}
	var zbxTypeENV string
	zbxTypeENV, _ = os.LookupEnv("CHRDS_ZBX_TYPE")
	if zbxTypeENV != "" {
		appCfg.ZBX.TYPE = zbxTypeENV
	}

	// Swagger
	var swgTitleENV string
	swgTitleENV, _ = os.LookupEnv("CHRDS_SWG_TITLE")
	if swgTitleENV != "" {
		appCfg.SWG.TITLE = swgTitleENV
	}
	var swgDescriptionENV string
	swgDescriptionENV, _ = os.LookupEnv("CHRDS_SWG_DESCRIPTION")
	if swgDescriptionENV != "" {
		appCfg.SWG.DESCRIPTION = swgDescriptionENV
	}
	var swgVersionENV string
	swgVersionENV, _ = os.LookupEnv("CHRDS_SWG_VERSION")
	if swgVersionENV != "" {
		appCfg.SWG.VERSION = swgVersionENV
	}
	var swgHostENV string
	swgHostENV, _ = os.LookupEnv("CHRDS_SWG_HOST")
	if swgHostENV != "" {
		appCfg.SWG.HOST = swgHostENV
	}
	var swgBasePathENV string
	swgBasePathENV, _ = os.LookupEnv("CHRDS_SWG_BASEPATH")
	if swgBasePathENV != "" {
		appCfg.SWG.BASEPATH = swgBasePathENV
	}
	var swgSchemeENV string
	swgSchemeENV, _ = os.LookupEnv("CHRDS_SWG_SCHEME")
	if swgSchemeENV != "" {
		appCfg.SWG.SCHEME = swgSchemeENV
	}

	log.Print("Configuration parameters:")
	log.Print("HTTP.HOST: ", appCfg.HTTP.HOST)
	log.Print("HTTP.PORT: ", appCfg.HTTP.PORT)
	log.Print("HTTP.READTIMEOUT: ", appCfg.HTTP.READTIMEOUT)
	log.Print("HTTP.WRITETIMEOUT: ", appCfg.HTTP.WRITETIMEOUT)
	log.Print("HTTP.RL: ", appCfg.HTTP.RL)
	log.Print("HTTP.TLS: ", appCfg.HTTP.TLS)
	log.Print("HTTP.CERTPATH: ", appCfg.HTTP.CERTPATH)
	log.Print("HTTP.KEYPATH: ", appCfg.HTTP.KEYPATH)
	log.Print("HTTP.CAPATH: ", appCfg.HTTP.CAPATH)

	log.Print("WORKER.POOL: ", appCfg.WORKER.POOL)
	log.Print("WORKER.POOL: ", appCfg.WORKER.ACTIME)

	log.Print("DB.HOSTS: ", appCfg.DB.HOSTS)
	log.Print("DB.USERNAME: ", appCfg.DB.USERNAME)
	log.Print("DB.TLS: ", appCfg.DB.TLS)
	log.Print("DB.CERTPATH: ", appCfg.DB.CERTPATH)
	log.Print("DB.KEYPATH: ", appCfg.DB.KEYPATH)
	log.Print("DB.CAPATH: ", appCfg.DB.CAPATH)
	log.Print("DB.HOSTVERIFICATION: ", appCfg.DB.HOSTVERIFICATION)
	log.Print("DB.KEYSPACE: ", appCfg.DB.KEYSPACE)
	log.Print("DB.CONSISTENCY: ", appCfg.DB.CONSISTENCY)
	log.Print("DB.CONSISTENCYREAD: ", appCfg.DB.CONSISTENCYREAD)
	log.Print("DB.TIMEOUT: ", appCfg.DB.TIMEOUT)

	log.Print("ZBX.HOST: ", appCfg.ZBX.HOST)
	log.Print("ZBX.PORT: ", appCfg.ZBX.PORT)
	log.Print("ZBX.TYPE: ", appCfg.ZBX.TYPE)

	log.Print("SWG.TITLE: ", appCfg.SWG.TITLE)
	log.Print("SWG.DESCRIPTION: ", appCfg.SWG.DESCRIPTION)
	log.Print("SWG.VERSION: ", appCfg.SWG.VERSION)
	log.Print("SWG.HOST: ", appCfg.SWG.HOST)
	log.Print("SWG.BASEPATH: ", appCfg.SWG.BASEPATH)
	log.Print("SWG.SCHEME: ", appCfg.SWG.SCHEME)

	return nil
}
