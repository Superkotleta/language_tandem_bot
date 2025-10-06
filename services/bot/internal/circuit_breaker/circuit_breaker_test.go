package circuit_breaker //nolint:testpackage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSuccessValue = "success"

func TestCircuitBreaker_StateClosed(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
	})

	// В закрытом состоянии запросы должны проходить
	result, err := cb.Execute(func() (interface{}, error) {
		return testSuccessValue, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, testSuccessValue, result)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_StateOpen(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 2
		},
	})

	// Вызываем несколько неудачных запросов
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, errors.New("test error")
		})
		assert.Error(t, err)
	}

	// Circuit Breaker должен быть в открытом состоянии
	assert.Equal(t, StateOpen, cb.State())

	// Следующий запрос должен быть заблокирован
	_, err := cb.Execute(func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is OPEN")
}

func TestCircuitBreaker_StateHalfOpen(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 2,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 1
		},
	})

	// Переводим в открытое состояние
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	// Ждем перехода в полуоткрытое состояние
	time.Sleep(150 * time.Millisecond)

	// Проверяем, что состояние изменилось
	assert.Equal(t, StateHalfOpen, cb.State())

	// Успешный запрос должен перевести в закрытое состояние
	result, err := cb.Execute(func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, testSuccessValue, result)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ExecuteWithContext(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
	})

	// Тест с отмененным контекстом
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cb.ExecuteWithContext(ctx, func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Тест с таймаутом контекста
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	_, err = cb.ExecuteWithContext(ctx, func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestCircuitBreaker_Counts(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
	})

	// Успешный запрос
	_, err := cb.Execute(func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.NoError(t, err)

	// Неудачный запрос
	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	// Проверяем счетчики
	counts := cb.Counts()
	assert.Equal(t, uint32(2), counts.Requests)
	assert.Equal(t, uint32(1), counts.TotalSuccesses)
	assert.Equal(t, uint32(1), counts.TotalFailures)
	assert.Equal(t, uint32(0), counts.ConsecutiveSuccesses)
	assert.Equal(t, uint32(1), counts.ConsecutiveFailures)
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	t.Parallel()

	var stateChanges []string

	cb := NewCircuitBreaker(Config{
		Name: "test",
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 1
		},
		OnStateChange: func(name string, from State, to State) {
			stateChanges = append(stateChanges, name+":"+from.String()+"->"+to.String())
		},
	})

	// Переводим в открытое состояние
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	// Проверяем, что callback был вызван
	assert.Len(t, stateChanges, 1)
	assert.Contains(t, stateChanges[0], "test:CLOSED->OPEN")
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, "default", cb.name)
	assert.Equal(t, uint32(3), cb.maxRequests)
	assert.Equal(t, 60*time.Second, cb.interval)
	assert.Equal(t, 60*time.Second, cb.timeout)
	assert.NotNil(t, cb.readyToTrip)
}

func TestCircuitBreaker_TelegramConfig(t *testing.T) {
	t.Parallel()

	config := TelegramConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, "telegram", cb.name)
	assert.Equal(t, uint32(5), cb.maxRequests)
	assert.Equal(t, 30*time.Second, cb.interval)
	assert.Equal(t, 30*time.Second, cb.timeout)
}

func TestCircuitBreaker_DatabaseConfig(t *testing.T) {
	t.Parallel()

	config := DatabaseConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, "database", cb.name)
	assert.Equal(t, uint32(10), cb.maxRequests)
	assert.Equal(t, 60*time.Second, cb.interval)
	assert.Equal(t, 30*time.Second, cb.timeout)
}

func TestCircuitBreaker_RedisConfig(t *testing.T) {
	t.Parallel()

	config := RedisConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, "redis", cb.name)
	assert.Equal(t, uint32(5), cb.maxRequests)
	assert.Equal(t, 30*time.Second, cb.interval)
	assert.Equal(t, 20*time.Second, cb.timeout)
}

func TestCircuitBreaker_PanicRecovery(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
	})

	// Проверяем, что паника перехватывается и возвращается как ошибка
	result, err := cb.Execute(func() (interface{}, error) {
		panic("test panic")
	})

	// Отладка
	t.Logf("Result: %v, Error: %v", result, err)

	// CircuitBreaker должен вернуть ошибку с информацией о панике
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "panic recovered")
		assert.Contains(t, err.Error(), "test panic")
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name: "test",
	})

	// Запускаем несколько горутин для конкурентного доступа
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			result, err := cb.Execute(func() (interface{}, error) {
				return id, nil
			})
			require.NoError(t, err)
			assert.Equal(t, id, result)
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что все запросы были обработаны
	counts := cb.Counts()
	assert.Equal(t, uint32(10), counts.Requests)
	assert.Equal(t, uint32(10), counts.TotalSuccesses)
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 2,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 1
		},
	})

	// Начальное состояние - закрытое
	assert.Equal(t, StateClosed, cb.State())

	// Переводим в открытое состояние
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	_, err = cb.Execute(func() (interface{}, error) {
		return nil, errors.New("test error")
	})
	assert.Error(t, err)

	// Должно быть открытое состояние
	assert.Equal(t, StateOpen, cb.State())

	// Ждем перехода в полуоткрытое состояние
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, StateHalfOpen, cb.State())

	// Успешный запрос должен перевести в закрытое состояние
	_, err = cb.Execute(func() (interface{}, error) {
		return testSuccessValue, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}
