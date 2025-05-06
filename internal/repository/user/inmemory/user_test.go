package inmemory

import (
	"context"
	"fmt"
	"homework/internal/domain"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository_SaveUser(t *testing.T) {
	t.Run("err, user is nil", func(t *testing.T) {
		sr := NewUserRepository()
		err := sr.SaveUser(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sr := NewUserRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := sr.SaveUser(ctx, &domain.User{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sr := NewUserRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := sr.SaveUser(ctx, &domain.User{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save", func(t *testing.T) {
		sr := NewUserRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		user := domain.User{
			Name: "User Name",
		}

		err := sr.SaveUser(ctx, &user)
		assert.NoError(t, err)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		sr := NewUserRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		for i := 0; i < 1000; i++ {
			user := &domain.User{
				Name: fmt.Sprintf("use name #%d", i),
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, sr.SaveUser(ctx, user))
			}()
		}

		wg.Wait()
	})
}
