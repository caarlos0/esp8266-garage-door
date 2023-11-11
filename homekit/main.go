package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/caarlos0/env/v10"
	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	operationTimeout = 30 * time.Second
	topicSensor      = "espgate/sensor"
	topicAct         = "espgate/act"
)

type Config struct {
	BrokerHost string `env:"MQTT_HOST" envDefault:"localhost"`
	BrokerPort int    `env:"MQTT_PORT" envDefault:"1883"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("failed to parse config", "err", err)
	}

	fs := hap.NewFsStore("./db")

	garage := accessory.NewGarageDoorOpener(accessory.Info{
		Name:         "Garage 2",
		Manufacturer: "Becker Software",
		Model:        "1",
		Firmware:     "0.0.1",
	})

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.BrokerHost, cfg.BrokerPort))
	opts.SetClientID("homekit_espgate")
	opts.OnConnect = func(_ mqtt.Client) {
		log.Info("connected to mqtt", "host", cfg.BrokerHost, "port", cfg.BrokerPort)
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Error("connection to mqtt lost", "err", err)
	}
	cli := mqtt.NewClient(opts)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("could not connect to mqtt", "token", token)
	}

	_ = cli.Publish(topicAct, 1, false, "ping")

	var once sync.Once

	cli.Subscribe(topicSensor, 1, func(_ mqtt.Client, msg mqtt.Message) {
		msg.Ack()
		if len(msg.Payload()) == 0 {
			return
		}
		if err := fs.Set("state", msg.Payload()); err != nil {
			log.Warn(
				"could not store event in cache",
				"err", err,
				"payload", string(msg.Payload()),
			)
		}
		log.Info("got msg", "payload", string(msg.Payload()))
		state := stateToInt(string(msg.Payload()))
		_ = garage.GarageDoorOpener.CurrentDoorState.SetValue(state)
		once.Do(func() {
			_ = garage.GarageDoorOpener.TargetDoorState.SetValue(state)
		})
	})

	garage.GarageDoorOpener.TargetDoorState.OnSetRemoteValue(func(v int) error {
		garage.GarageDoorOpener.ObstructionDetected.SetValue(false)
		switch v {
		case characteristic.TargetDoorStateOpen:
			if token := cli.Publish(topicAct, 1, false, "open"); token.Wait() && token.Error() != nil {
				return fmt.Errorf("failed to change status")
			}
		case characteristic.TargetDoorStateClosed:
			if token := cli.Publish(topicAct, 1, false, "close"); token.Wait() && token.Error() != nil {
				return fmt.Errorf("failed to change status")
			}
		}

		go func() {
			time.Sleep(operationTimeout)
			if garage.GarageDoorOpener.CurrentDoorState.Value() != garage.GarageDoorOpener.TargetDoorState.Value() {
				garage.GarageDoorOpener.ObstructionDetected.SetValue(true)
			}
		}()
		return nil
	})

	server, err := hap.NewServer(fs, garage.A)
	if err != nil {
		log.Fatal("fail to start server", "error", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		log.Info("stopping server...")
		signal.Stop(c)
		cancel()
	}()

	log.Info("starting server...")
	if err := server.ListenAndServe(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to close server", "err", err)
	}
}

func stateToInt(s string) int {
	if s == "open" {
		return characteristic.CurrentDoorStateOpen
	}
	return characteristic.CurrentDoorStateClosed
}
