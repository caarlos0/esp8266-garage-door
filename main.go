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

	a := accessory.NewGarageDoorOpener(accessory.Info{
		Name:         "Garage 2",
		Manufacturer: "Becker Software",
		Model:        "1",
		Firmware:     "0.0.1",
	})

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.BrokerHost, cfg.BrokerPort))
	opts.SetClientID("homekit-garage")
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

	ping := func(from string) {
		log.Info("ping", "from", from)
		_ = cli.Publish(topicAct, 1, false, "ping")
	}

	ping("main")
	go func() {
		// when power goes out, sometimes things are not up yet... this gives
		// an extra ping after a couple of minutes on startup... should help
		// sync things up...
		time.Sleep(time.Minute * 5)
		ping("delayed-main")
	}()

	var once sync.Once
	lastAction := &atomicTime{}

	cli.Subscribe(topicSensor, 1, func(_ mqtt.Client, msg mqtt.Message) {
		msg.Ack()
		if len(msg.Payload()) == 0 {
			return
		}

		log.Info("got msg", "payload", string(msg.Payload()))
		garage := a.GarageDoorOpener

		once.Do(func() {
			// startup...
			state := stateToInt(string(msg.Payload()))
			_ = garage.CurrentDoorState.SetValue(state)
			_ = garage.TargetDoorState.SetValue(state)
		})

		switch string(msg.Payload()) {
		case "open":
			if lastAction.IsZero() || lastAction.Since() >= operationTimeout {
				log.Info("open")
				lastAction.Zero()
				_ = garage.TargetDoorState.SetValue(characteristic.CurrentDoorStateOpen)
				_ = garage.CurrentDoorState.SetValue(characteristic.CurrentDoorStateOpen)
				garage.ObstructionDetected.SetValue(false)
				return
			}
			log.Info("opening")
			_ = garage.TargetDoorState.SetValue(characteristic.CurrentDoorStateOpening)
			_ = garage.CurrentDoorState.SetValue(characteristic.CurrentDoorStateOpening)
			go func() {
				time.Sleep(operationTimeout - lastAction.Since())
				ping("open msg")
			}()
		case "closed":
			if lastAction.IsZero() || lastAction.Since() >= operationTimeout {
				log.Info("closed")
				lastAction.Zero()
				_ = garage.TargetDoorState.SetValue(characteristic.CurrentDoorStateClosed)
				_ = garage.CurrentDoorState.SetValue(characteristic.CurrentDoorStateClosed)
				garage.ObstructionDetected.SetValue(false)
				return
			}
			log.Info("closing")
			_ = garage.CurrentDoorState.SetValue(characteristic.CurrentDoorStateClosing)
			go func() {
				time.Sleep(operationTimeout - lastAction.Since())
				ping("closed msg")
			}()
		}
	})

	a.GarageDoorOpener.TargetDoorState.OnSetRemoteValue(func(v int) error {
		lastAction.Set(time.Now())
		garage := a.GarageDoorOpener
		garage.ObstructionDetected.SetValue(false)
		switch v {
		case characteristic.TargetDoorStateOpen:
			if garage.CurrentDoorState.Value() == characteristic.CurrentDoorStateOpen ||
				garage.CurrentDoorState.Value() == characteristic.CurrentDoorStateOpening {
				log.Info("already open/opening")
				return nil
			}
			log.Info("request open")
			if token := cli.Publish(topicAct, 1, false, "open"); token.Wait() && token.Error() != nil {
				return fmt.Errorf("failed to change status")
			}
		case characteristic.TargetDoorStateClosed:
			if garage.CurrentDoorState.Value() == characteristic.CurrentDoorStateClosed ||
				garage.CurrentDoorState.Value() == characteristic.CurrentDoorStateClosing {
				log.Info("already closed/closing")
				return nil
			}
			log.Info("request close")
			if token := cli.Publish(topicAct, 1, false, "close"); token.Wait() && token.Error() != nil {
				return fmt.Errorf("failed to change status")
			}
		}
		return nil
	})

	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			garage := a.GarageDoorOpener
			if garage.TargetDoorState.Value() == garage.CurrentDoorState.Value() {
				if garage.ObstructionDetected.Value() {
					log.Info("not obstructed")
					garage.ObstructionDetected.SetValue(false)
				}
				continue
			}

			if lastAction.Since() > operationTimeout+5*time.Second {
				if garage.ObstructionDetected.Value() {
					log.Info("obstructed")
					garage.ObstructionDetected.SetValue(true)
				}
				ping("obstructed")
			}
		}
	}()

	server, err := hap.NewServer(fs, a.A)
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

type atomicTime struct {
	t time.Time
	m sync.Mutex
}

func (a *atomicTime) Since() time.Duration {
	return time.Since(a.Get())
}

func (a *atomicTime) Zero() {
	a.Set(time.Time{})
}

func (a *atomicTime) IsZero() bool {
	return a.Get().IsZero()
}

func (a *atomicTime) Get() time.Time {
	a.m.Lock()
	defer a.m.Unlock()
	return a.t
}

func (a *atomicTime) Set(t time.Time) {
	a.m.Lock()
	defer a.m.Unlock()
	a.t = t
}
