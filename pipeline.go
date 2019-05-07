package pipeline

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Pipeline struct {
	channelQuit        chan os.Signal
	channelFileCreated chan string
	waitGroup          sync.WaitGroup
	ctx                context.Context
	cancel             context.CancelFunc
	environment        string
}

func NewPipeline(configPath string) *Pipeline {
	return new(Pipeline).init(configPath)
}

func (pipeline *Pipeline) init(configPath string) *Pipeline {

	pipeline.environment = pipeline.readEnvName()
	pipeline.readConfigFile(configPath)
	pipeline.setLogger()

	log.Info("Initializing pipeline")

	pipeline.channelQuit = make(chan os.Signal, 1)
	pipeline.channelFileCreated = make(chan string, 1)

	pipeline.ctx, pipeline.cancel = context.WithCancel(context.Background())
	signal.Notify(pipeline.channelQuit, os.Interrupt, syscall.SIGTERM)

	log.Info("pipeline is initialized")
	return pipeline
}

func (pipeline *Pipeline) readEnvName() string {
	environment := os.Getenv("PIPELINE_ENV")
	if len(environment) == 0 {
		environment = "DEBUG"
	} else {
		if !parseEnvironment(environment) {
			panic(fmt.Sprintf("not a valid environment : %s", environment))
		}
	}
	return environment
}

func (pipeline *Pipeline) readConfigFile(path string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/pipeline/")
	viper.AddConfigPath("$HOME/.pipeline")
	viper.AddConfigPath(".")
	if path != "." {
		viper.AddConfigPath(path)
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func (pipeline *Pipeline) setLogger() {
	output := viper.GetString("LOG_OUTPUT")
	switch output {
	case "stdout":
		log.SetFormatter(&easy.Formatter{TimestampFormat: "2006/01/02 15:04:05", LogFormat: "[%lvl%]: %time% - %msg%\n"})
		log.SetOutput(os.Stdout)
	case "syslog":
		sysLogger, err := syslog.Dial("", "", syslog.LOG_INFO|syslog.LOG_DAEMON, "PIPIELINE")
		if err != nil {
			log.Fatal(err)
		}
		log.SetFormatter(&easy.Formatter{TimestampFormat: "2006/01/02 15:04:05", LogFormat: "%msg%\n"})
		log.SetOutput(sysLogger)
	default:
		panic(fmt.Sprintf("not a valid logging output : %s", output))
	}
	level, err := log.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)
}

func (pipeline *Pipeline) Run() {
	log.Info("Running pipeline")
	<-pipeline.channelQuit
}

func (pipeline *Pipeline) Done() <-chan struct{} {
	return pipeline.ctx.Done()
}

func (pipeline *Pipeline) Context() context.Context {
	return pipeline.ctx
}

func (pipeline *Pipeline) AddTask(e Executor, name string) {
	if e == nil || name == "" {
		log.Fatal("Executor or Name must be valued")
	}
	e.Init(pipeline, name)
	pipeline.waitGroup.Add(1)
	log.Infof("Add Job \"%s\" id: %s", e.Name(), e.ID())
	go e.Run()
}

func (pipeline *Pipeline) TearDown(name string, uuid uuid.UUID) {
	log.Infof("teardown job \"%s\" id: %s", name, uuid.String())
	pipeline.waitGroup.Done()
}

func (pipeline *Pipeline) Stop() {
	log.Info("Stopping pipeline")
	pipeline.cancel()
	close(pipeline.channelQuit)
	pipeline.waitGroup.Wait()
	log.Info("pipeline is stopped")
}
