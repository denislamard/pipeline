package pipeline

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"log/syslog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func parseEnvironment(name string) bool {
	switch strings.ToUpper(name) {
	case "DEBUG":
		return true
	case "INTEGRATION":
		return true
	case "PRODUCTION":
		return true
	default:
		return false
	}
}

func parseLogOutput(name string) bool {
	switch strings.ToLower(name) {
	case "stdout":
		return true
	case "syslog":
		return true
	default:
		return false
	}
}

type Pipeline struct {
	channelQuit        chan os.Signal
	channelFileCreated chan string
	waitGroup          sync.WaitGroup
	ctx                context.Context
	cancel             context.CancelFunc

	tasks map[string]*Task
}

func NewPipeline() *Pipeline {
	return new(Pipeline).init()
}

func (pipeline *Pipeline) init() *Pipeline {
	log.Info("Initializing pipeline")

	pipeline.channelQuit = make(chan os.Signal, 1)
	pipeline.channelFileCreated = make(chan string, 1)

	pipeline.ctx, pipeline.cancel = context.WithCancel(context.Background())
	signal.Notify(pipeline.channelQuit, os.Interrupt, syscall.SIGTERM)

	pipeline.tasks = make(map[string]*Task)
	log.Info("pipeline is initialized")
	return pipeline
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

func init() {
	configPath := flag.String("config", ".", "path for config file")
	flag.Parse()
	env := os.Getenv("PIPELINE_ENV")
	if len(env) == 0 {
		env = "DEBUG"
	}
	if !parseEnvironment(env) {
		panic(fmt.Sprintf("not a valid environment : %s", env))
	}
	readConfigFile(*configPath, fmt.Sprintf("config.%s", strings.ToLower(env)))
	logOutput := viper.GetString("LOG_OUTPUT")
	if !parseLogOutput(logOutput) {
		panic(fmt.Sprintf("not a valid logging output : %s", logOutput))
	}
	setLogger(env, logOutput)
}

func setLogger(levelName string, logOutput string) {
	level, err := log.ParseLevel(levelName)
	if err != nil {
		panic(err)
	}
	switch logOutput {
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
	}
	log.SetLevel(level)
}

func readConfigFile(path string, configName string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName(configName)
	viper.AddConfigPath("/etc/pipeline/")
	viper.AddConfigPath("$HOME/.pipeline")
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
