package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//go:embed index.html
var index []byte

const (
	operationTimeout = 30 * time.Second
	topicSensor      = "espgate/sensor"
	topicAct         = "espgate/act"
)

const (
	normallyOpen   = "open"
	normallyClosed = "closed"
	stateOpen      = "open"
	stateClosed    = "closed"
)

type Config struct {
	BrokerHost string `env:"MQTT_HOST" envDefault:"localhost"`
	BrokerPort int    `env:"MQTT_PORT" envDefault:"1883"`
	Address    string `env:"LISTEN" envDefault:":8989"`
	Normally   string `env:"NORMALLY" envDefault:"open"`
}

func main() {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
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

		adjustedState := adjustState(string(msg.Payload()), cfg.Normally)
		log.Info("got msg", "payload", string(msg.Payload()), "adjusted-state", adjustedState)

		garage := a.GarageDoorOpener

		once.Do(func() {
			// startup...
			state := stateToInt(adjustedState)
			_ = garage.CurrentDoorState.SetValue(state)
			_ = garage.TargetDoorState.SetValue(state)
		})

		switch adjustedState {
		case stateOpen:
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
		case stateClosed:
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server, err := hap.NewServer(fs, a.A)
	if err != nil {
		log.Fatal("fail to start server", "error", err)
	}
	server.Addr = cfg.Address
	server.ServeMux().Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := [5]string{"open", "closed", "opening", "closed", "stopped"}[a.GarageDoorOpener.CurrentDoorState.Value()]
		tp := template.Must(template.New("index").Parse(string(index)))
		_ = tp.Execute(w, struct {
			State string
		}{
			state,
		})
	}))
	server.ServeMux().Handle("/open", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request open")
		if token := cli.Publish(topicAct, 1, false, "open"); token.Wait() && token.Error() != nil {
			log.Error("could not open", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	server.ServeMux().Handle("/close", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request close")
		if token := cli.Publish(topicAct, 1, false, "close"); token.Wait() && token.Error() != nil {
			log.Error("could not close", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	log.Info("starting server...", "addr", server.Addr)
	if err := server.ListenAndServe(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to close server", "err", err)
	}
}

func stateToInt(s string) int {
	if s == stateOpen {
		return characteristic.CurrentDoorStateOpen
	}
	return characteristic.CurrentDoorStateClosed
}

func adjustState(state, normally string) string {
	if normally == normallyOpen {
		if state == stateClosed {
			return stateOpen
		}
		return stateClosed
	}
	return state
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
