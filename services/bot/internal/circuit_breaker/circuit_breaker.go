// Package circuit_breaker implements circuit breaker pattern for fault tolerance.
package circuit_breaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"language-exchange-bot/internal/localization"
)

// State представляет состояние Circuit Breaker.
type State int

const (
	// StateClosed - закрытое состояние, запросы проходят нормально.
	StateClosed State = iota
	// StateOpen - открытое состояние, запросы блокируются.
	StateOpen
	// StateHalfOpen - полуоткрытое состояние, ограниченное количество запросов.
	StateHalfOpen
)

// Circuit Breaker константы теперь в localization/constants.go

// String возвращает строковое представление состояния.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config содержит конфигурацию Circuit Breaker.
type Config struct {
	// Name - имя Circuit Breaker для логирования.
	Name string
	// MaxRequests - максимальное количество запросов в полуоткрытом состоянии.
	MaxRequests uint32
	// Interval - интервал для сброса счетчика ошибок.
	Interval time.Duration
	// Timeout - время ожидания в открытом состоянии перед переходом в полуоткрытое.
	Timeout time.Duration
	// ReadyToTrip - функция для определения готовности к переходу в открытое состояние.
	ReadyToTrip func(counts Counts) bool
	// OnStateChange - callback при изменении состояния.
	OnStateChange func(name string, from State, to State)
}

// Counts содержит счетчики для Circuit Breaker.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreaker реализует паттерн Circuit Breaker.
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	onStateChange func(name string, from State, to State)

	mutex      sync.Mutex
	state      State
	generation uint32
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker создает новый Circuit Breaker.
func NewCircuitBreaker(config Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:          config.Name,
		maxRequests:   config.MaxRequests,
		interval:      config.Interval,
		timeout:       config.Timeout,
		readyToTrip:   config.ReadyToTrip,
		onStateChange: config.OnStateChange,
		state:         StateClosed,
		generation:    0,
		counts:        Counts{},
		expiry:        time.Time{},
	}

	// Устанавливаем значения по умолчанию
	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}

	if cb.interval == 0 {
		cb.interval = localization.DefaultIntervalSeconds * time.Second
	}

	if cb.timeout == 0 {
		cb.timeout = localization.DefaultTimeoutSeconds * time.Second
	}

	if cb.readyToTrip == nil {
		cb.readyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > localization.DefaultConsecutiveFailures
		}
	}

	return cb
}

// Execute выполняет функцию с защитой Circuit Breaker.
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	// Переменные для результата
	var result interface{}
	var execErr error
	panicOccurred := false

	// Функция для выполнения с перехватом паники
	func() {
		defer func() {
			if e := recover(); e != nil {
				panicOccurred = true
				// Любая паника считается recovered
				execErr = fmt.Errorf("panic recovered: %v", e)
				result = nil
			}
		}()

		result, execErr = req()
	}()

	// Обновляем счетчики на основе результата
	if panicOccurred || execErr != nil {
		cb.afterRequest(generation, false)
	} else {
		cb.afterRequest(generation, true)
	}

	return result, execErr
}

// ExecuteWithContext выполняет функцию с контекстом и защитой Circuit Breaker.
func (cb *CircuitBreaker) ExecuteWithContext(
	ctx context.Context,
	req func() (interface{}, error),
) (result interface{}, err error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	// Проверяем контекст
	select {
	case <-ctx.Done():
		cb.afterRequest(generation, false)
		return nil, ctx.Err()
	default:
	}

	// Переменные для результата
	var execErr error
	panicOccurred := false

	// Функция для выполнения с перехватом паники
	func() {
		defer func() {
			if e := recover(); e != nil {
				panicOccurred = true
				// Любая паника считается recovered
				execErr = fmt.Errorf("panic recovered: %v", e)
				result = nil
			}
		}()

		result, execErr = req()
	}()

	// Обновляем счетчики на основе результата
	if panicOccurred || execErr != nil {
		cb.afterRequest(generation, false)
	} else {
		cb.afterRequest(generation, true)
	}

	return result, execErr
}

// State возвращает текущее состояние Circuit Breaker.
func (cb *CircuitBreaker) State() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)

	return state
}

// Counts возвращает текущие счетчики.
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	_, _ = cb.currentState(now)

	return cb.counts
}

// beforeRequest проверяет возможность выполнения запроса.
func (cb *CircuitBreaker) beforeRequest() (uint32, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, errors.New("circuit breaker is OPEN")
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, errors.New("circuit breaker is HALF_OPEN and max requests reached")
	}

	cb.counts.Requests++

	return generation, nil
}

// afterRequest обновляет счетчики после выполнения запроса.
func (cb *CircuitBreaker) afterRequest(before uint32, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	_, generation := cb.currentState(now)

	if generation != before {
		return
	}

	if success {
		cb.onSuccess(now)
	} else {
		cb.onFailure(now)
	}
}

// currentState возвращает текущее состояние и поколение.
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint32) {
	if cb.expiry.Before(now) && !cb.expiry.IsZero() {
		cb.toNewGeneration(now)

		// Переход из Open в HalfOpen при истечении timeout
		if cb.state == StateOpen {
			cb.state = StateHalfOpen
		}
	}

	return cb.state, cb.generation
}

// toNewGeneration сбрасывает счетчики и обновляет поколение.
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time

	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	case StateHalfOpen:
		cb.expiry = zero
	}
}

// onSuccess обрабатывает успешный запрос.
func (cb *CircuitBreaker) onSuccess(now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if cb.state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
		cb.setState(StateClosed, now)
	}
}

// onFailure обрабатывает неудачный запрос.
func (cb *CircuitBreaker) onFailure(now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if cb.readyToTrip(cb.counts) {
		cb.setState(StateOpen, now)
	}
}

// setState изменяет состояние Circuit Breaker.
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() Config {
	return Config{
		Name:        "default",
		MaxRequests: localization.DefaultMaxRequests,
		Interval:    localization.DefaultIntervalSeconds * time.Second,
		Timeout:     localization.DefaultTimeoutSeconds * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > localization.DefaultConsecutiveFailures
		},
	}
}

// TelegramConfig возвращает конфигурацию для Telegram API.
func TelegramConfig() Config {
	return Config{
		Name:        "telegram",
		MaxRequests: localization.TelegramMaxRequests,
		Interval:    localization.TelegramIntervalSeconds * time.Second,
		Timeout:     localization.TelegramTimeoutSeconds * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > localization.TelegramFailureThreshold
		},
	}
}

// DatabaseConfig возвращает конфигурацию для базы данных.
func DatabaseConfig() Config {
	return Config{
		Name:        "database",
		MaxRequests: localization.DatabaseMaxRequests,
		Interval:    localization.DatabaseIntervalSeconds * time.Second,
		Timeout:     localization.DatabaseTimeoutSeconds * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > localization.DatabaseFailureThreshold
		},
	}
}

// RedisConfig возвращает конфигурацию для Redis.
func RedisConfig() Config {
	return Config{
		Name:        "redis",
		MaxRequests: localization.MatcherMaxRequests,
		Interval:    localization.MatcherIntervalSeconds * time.Second,
		Timeout:     localization.MatcherTimeoutSeconds * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > localization.MatcherFailureThreshold
		},
	}
}
