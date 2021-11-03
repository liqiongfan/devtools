package core

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Message all messages delivered with the following structure
// When sending message to core kernel, omit ID parameter
type Message struct {
	Id         int64  `json:"id"`
	Identifier int64  `json:"identifier"`
	Data       string `json:"data"`
	Signal     int    `json:"signal"`
}

// Tasks all tasks need to be running each seconds
type Tasks struct {
	m      sync.Mutex
	runner []func()
}

// idType inner struct used for global message identifier
type idType struct {
	M  sync.Mutex
	Id int64
}

var (
	// inMessage messages from other plugins
	inMessage chan Message

	// modules plugins that installed
	// modules map[int64]chan Message
	modules sync.Map

	// initialised once used sync.Once.Do(func(){})
	once sync.Once

	// _gId global id(identifier)
	_gId idType

	// runner 1s tasks
	runner Tasks

	// Logger kernel logger
	Logger *zap.Logger

	// err kernel error
	err error
)

const (
	// NORMAL common signal
	NORMAL = 0
	// SignalKill exit signal
	// This signal will need the plugin to close the message channel
	SignalKill = 1
)

// increaseId Increase global id and return the result
func increaseId() int64 {
	_gId.M.Lock()
	defer _gId.M.Unlock()
	if _gId.Id >= 2<<60 {
		_gId.Id = 0
	}
	_gId.Id++
	return _gId.Id
}

// init
func init() {
	once.Do(func() {
		if inMessage == nil {
			inMessage = make(chan Message, 1000)
		}

		Logger, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	})
}

// InstallModule install plugin into core kernel
func InstallModule(id int64, queue chan Message) bool {
	if _, ok := modules.Load(id); ok {
		Logger.Error(fmt.Sprintf("Plugins already installed with the identifier: %d", id))
		return false
	}
	modules.Store(id, queue)
	return true
}

// IsExitSignal check whether signal is exit signal
// all sub-goroutine need be exit when receiving this signal
func IsExitSignal(m *Message) bool {
	if SignalKill == m.Signal {
		return true
	}
	return false
}

// PrintModule Print all modules which installed in core-kernel
func PrintModule() {
	modules.Range(func(key, value interface{}) bool {
		Logger.Info(fmt.Sprintf("Installed module[id: %d]", key.(int64)))
		return true
	})
}

// UninstallModule uninstall plugin from core kernel
func UninstallModule(id int64) bool {
	if _, ok := modules.Load(id); ok {
		modules.Delete(id)
	}
	return true
}

type option struct {
	Identifier int64
	Data       string
	Signal     int
}

type sendOption func(*option)

func WithIdInt64(identifier int64) sendOption {
	return func(o *option) {
		o.Identifier = identifier
	}
}

func WithIdInt(identifier int) sendOption {
	return func(o *option) {
		o.Identifier = int64(identifier)
	}
}

func WithData(data string) sendOption {
	return func(o *option) {
		o.Data = data
	}
}

func WithSignal(signal int) sendOption {
	return func(o *option) {
		o.Signal = signal
	}
}

// SendMessage send message to core
func SendMessage(opts ...sendOption) bool {

	opt := &option{
		Identifier: 0,
		Data:       "",
		Signal:     NORMAL,
	}

	for _, o := range opts {
		o(opt)
	}

	if _, ok := modules.Load(opt.Identifier); !ok {
		Logger.Info(fmt.Sprintf("Please install plugin to deal with the message with identifier: %d", opt.Identifier))
		return false
	}

	defer func() {
		if r := recover(); r != nil {

		}
	}()

	inMessage <- Message{
		Signal:     opt.Signal,
		Identifier: opt.Identifier,
		Data:       opt.Data,
	}

	return true
}

// deliverMessage Deliver message into different message queue
func deliverMessage(message Message) bool {
	if queue, exists := modules.Load(message.Identifier); exists {
		defer func() {
			// IsExitSignal will uninstall module when meeting the SignalKill
			if IsExitSignal(&message) {
				UninstallModule(message.Identifier)
			}
		}()
		queue.(chan Message) <- Message{
			Id:         increaseId(),
			Identifier: message.Identifier,
			Data:       message.Data,
			Signal:     message.Signal,
		}
		return true
	} else {
		// Message discard immediately
		Logger.Info(fmt.Sprintf("Message from: %d value:[%s] discarded", message.Identifier, message.Data))
		return false
	}
}

// startInMessage this function will run forever
func startInMessage(w *sync.WaitGroup) {
	for {
		select {
		case message, ok := <-inMessage:
			if ok {
				deliverMessage(message)
			} else {
				w.Done()
				return
			}
		case <-time.NewTimer(time.Second).C:
			for _, t := range runner.runner {
				t()
			}
		}
	}
}

// Shutdown the core kernel
// very dangerous, when called, all goroutine will exit
func Shutdown() {
	close(inMessage)
}

// AppendToSecondTask Append task to kernel 1s task
func AppendToSecondTask(_t func()) int {
	runner.m.Lock()
	defer runner.m.Unlock()

	runner.runner = append(runner.runner, _t)
	return len(runner.runner) - 1
}

// RemoveSecondTask Remove task from kernel 1s task
func RemoveSecondTask(i int) bool {
	if len(runner.runner) <= i {
		return false
	}
	runner.m.Lock()
	defer runner.m.Unlock()

	var _v []func()
	for _i, t := range runner.runner {
		if _i == i {
			continue
		}
		_v = append(_v, t)
	}
	runner.runner = _v
	return true
}

// Start the engine core
func Start() {
	defer func() {
		_ = Logger.Sync()
	}()
	wait := sync.WaitGroup{}
	wait.Add(1)
	go startInMessage(&wait)
	wait.Wait()
}
